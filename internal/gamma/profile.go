package gamma

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// CreateProfileRequest mirrors the JSON body the polymarket.com web UI
// posts to gamma-api.polymarket.com/profiles after SIWE login. The shape
// was decoded from a Playwright capture (see scripts/playwright-capture/
// and BLOCKERS.md "CORRECTION 2026-05-08").
//
// `ProxyWallet` is the maker address polygolem will register: for the
// deposit-wallet path it's `MakerAddressForSignatureType(eoa, 137, 3)`;
// for the legacy proxy path it's the sigtype-1 CREATE2 proxy. The web UI
// only ever registers sigtype-1, but the backend accepts the sigtype-3
// form too because /profiles just persists whatever address the client
// sends (the backend doesn't re-derive).
type CreateProfileRequest struct {
	DisplayUsernamePublic bool                `json:"displayUsernamePublic"`
	EmailOptIn            bool                `json:"emailOptIn"`
	WalletActivated       bool                `json:"walletActivated"`
	Name                  string              `json:"name"`
	Pseudonym             string              `json:"pseudonym"`
	ProxyWallet           string              `json:"proxyWallet"`
	Users                 []CreateProfileUser `json:"users"`
}

// CreateProfileUser is a single user-row inside a CreateProfileRequest.
// `Address` is the EOA, `ProxyWallet` is the maker the EOA owns.
type CreateProfileUser struct {
	Address           string            `json:"address"`
	Email             string            `json:"email"`
	IsExternalAuth    bool              `json:"isExternalAuth"`
	ProxyWallet       string            `json:"proxyWallet"`
	Username          string            `json:"username"`
	Provider          string            `json:"provider"`
	Preferences       []json.RawMessage `json:"preferences,omitempty"`
	WalletPreferences []json.RawMessage `json:"walletPreferences,omitempty"`
}

// CreateProfileResponse is the 201 body returned by /profiles. Only the
// fields polygolem actually reads are typed; the rest is preserved in
// Raw for callers that need to inspect server-set defaults.
type CreateProfileResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ProxyWallet string `json:"proxyWallet"`
	Pseudonym   string `json:"pseudonym"`
	Raw         json.RawMessage
}

// UnmarshalJSON keeps Raw populated alongside the typed fields.
func (r *CreateProfileResponse) UnmarshalJSON(data []byte) error {
	type alias CreateProfileResponse
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*r = CreateProfileResponse(a)
	r.Raw = append(r.Raw[:0], data...)
	return nil
}

// DefaultEmailNotificationPreferences is the shape the web UI ships when
// creating a new profile. Encoded as JSON-in-string the way Polymarket's
// schema requires.
const DefaultEmailNotificationPreferences = `{"generalEmail":{"sendEmails":false},"marketEmails":{"sendEmails":false},"newsletterEmails":{"sendEmails":false},"promotionalEmails":{"sendEmails":false},"eventEmails":{"sendEmails":false,"tagIds":[]},"orderFillEmails":{"sendEmails":false,"hideSmallFills":true},"resolutionEmails":{"sendEmails":false}}`

// DefaultAppNotificationPreferences mirrors the web UI's app-notif defaults.
const DefaultAppNotificationPreferences = `{"eventApp":{"sendApp":true,"tagIds":[]},"marketPriceChangeApp":{"sendApp":true},"orderFillApp":{"sendApp":true,"hideSmallFills":true},"resolutionApp":{"sendApp":true}}`

// DefaultPreferencesBlock returns the "preferences" array element the web
// UI sends in the users[] entry — a single object describing notification
// settings as JSON-in-string.
func DefaultPreferencesBlock() json.RawMessage {
	body, _ := json.Marshal(map[string]interface{}{
		"preferencesStatus":            "New/Existing - Created Prefs",
		"subscriptionStatus":           false,
		"emailNotificationPreferences": DefaultEmailNotificationPreferences,
		"appNotificationPreferences":   DefaultAppNotificationPreferences,
		"marketInterests":              "[]",
	})
	return body
}

// DefaultWalletPreferencesBlock returns the "walletPreferences" element
// the web UI sends.
func DefaultWalletPreferencesBlock() json.RawMessage {
	body, _ := json.Marshal(map[string]interface{}{
		"advancedMode":            false,
		"customGasPrice":          "30",
		"gasPreference":           "fast",
		"walletPreferencesStatus": "New/Existing - Created Wallet Prefs",
	})
	return body
}

// NewCreateProfileRequest builds a request that mirrors the captured web
// UI signup payload. The pseudonym/name prefix uses the proxyWallet
// address + a millisecond timestamp (web UI convention). Provider names
// the integration source — pass "metamask" for an injected EIP-1193
// provider, which is what the web UI sends for sigtype-1 signups.
func NewCreateProfileRequest(eoaAddress, proxyWallet, provider string, nowMillis int64) CreateProfileRequest {
	username := fmt.Sprintf("%s-%d", proxyWallet, nowMillis)
	return CreateProfileRequest{
		DisplayUsernamePublic: true,
		EmailOptIn:            false,
		WalletActivated:       false,
		Name:                  username,
		Pseudonym:             proxyWallet,
		ProxyWallet:           proxyWallet,
		Users: []CreateProfileUser{
			{
				Address:           eoaAddress,
				Email:             "",
				IsExternalAuth:    true,
				ProxyWallet:       proxyWallet,
				Username:          username,
				Provider:          provider,
				Preferences:       []json.RawMessage{DefaultPreferencesBlock()},
				WalletPreferences: []json.RawMessage{DefaultWalletPreferencesBlock()},
			},
		},
	}
}

// CreateProfile registers a fresh EOA + proxyWallet pair with Polymarket's
// gamma backend. The supplied http.Client must already carry the SIWE
// session cookie (call [auth.SIWESession.Login] first; pass session.HTTPClient()).
//
// Returns the created profile (id, server-confirmed proxyWallet, etc.).
//
// Idempotency: gamma returns 409 if a profile already exists for the
// signed-in EOA. Callers that want create-or-fetch behaviour should
// catch the 409 and fall back to PublicProfile.
func CreateProfile(ctx context.Context, client *http.Client, gammaURL string, body CreateProfileRequest) (CreateProfileResponse, error) {
	if client == nil {
		return CreateProfileResponse{}, fmt.Errorf("client is required")
	}
	if strings.TrimSpace(gammaURL) == "" {
		return CreateProfileResponse{}, fmt.Errorf("gammaURL is required")
	}
	if strings.TrimSpace(body.ProxyWallet) == "" {
		return CreateProfileResponse{}, fmt.Errorf("ProxyWallet is required")
	}
	if len(body.Users) == 0 || strings.TrimSpace(body.Users[0].Address) == "" {
		return CreateProfileResponse{}, fmt.Errorf("Users[0].Address is required")
	}

	url := strings.TrimRight(gammaURL, "/") + "/profiles"
	raw, err := json.Marshal(body)
	if err != nil {
		return CreateProfileResponse{}, fmt.Errorf("marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return CreateProfileResponse{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return CreateProfileResponse{}, fmt.Errorf("create profile: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return CreateProfileResponse{}, fmt.Errorf("HTTP %d %s: %s", resp.StatusCode, url, string(respBody))
	}

	var out CreateProfileResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return CreateProfileResponse{}, fmt.Errorf("decode profile response: %w", err)
	}
	return out, nil
}
