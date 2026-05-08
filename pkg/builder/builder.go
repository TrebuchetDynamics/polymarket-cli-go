package builder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/TrebuchetDynamics/polygolem/internal/auth"
)

const (
	POLY_BUILDER_API_KEY    = "POLY_BUILDER_API_KEY"
	POLY_BUILDER_PASSPHRASE = "POLY_BUILDER_PASSPHRASE"
	POLY_BUILDER_TIMESTAMP  = "POLY_BUILDER_TIMESTAMP"
	POLY_BUILDER_SIGNATURE  = "POLY_BUILDER_SIGNATURE"
)

// Signer creates POLY_BUILDER_* headers for relayer/builder requests.
type Signer interface {
	CreateHeaders(method, path string, body *string, timestamp *int64) (map[string]string, error)
}

type LocalSignerConfig struct {
	Key        string
	Secret     string
	Passphrase string
}

type LocalSigner struct {
	config LocalSignerConfig
}

func NewLocalSigner(config LocalSignerConfig) (*LocalSigner, error) {
	if config.Key == "" || config.Secret == "" || config.Passphrase == "" {
		return nil, fmt.Errorf("builder signer config incomplete")
	}
	return &LocalSigner{config: config}, nil
}

func (s *LocalSigner) CreateHeaders(method, path string, body *string, timestamp *int64) (map[string]string, error) {
	ts := time.Now().Unix()
	if timestamp != nil {
		ts = *timestamp
	}
	return auth.BuildBuilderHeaders(&auth.BuilderConfig{
		Key:        s.config.Key,
		Secret:     s.config.Secret,
		Passphrase: s.config.Passphrase,
	}, ts, method, path, body)
}

type RemoteSignerConfig struct {
	URL   string
	Token string
}

type RemoteSigner struct {
	config RemoteSignerConfig
	client *http.Client
}

func NewRemoteSigner(config RemoteSignerConfig) (*RemoteSigner, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("remote signer URL is required")
	}
	if config.Token == "" {
		return nil, fmt.Errorf("remote signer token is required")
	}
	return &RemoteSigner{config: config, client: http.DefaultClient}, nil
}

func (s *RemoteSigner) CreateHeaders(method, path string, body *string, timestamp *int64) (map[string]string, error) {
	payload := map[string]interface{}{
		"method": method,
		"path":   path,
	}
	if body != nil {
		payload["body"] = *body
	}
	if timestamp != nil {
		payload["timestamp"] = strconv.FormatInt(*timestamp, 10)
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, s.config.URL, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.config.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("remote signer returned HTTP %d", resp.StatusCode)
	}
	var headers map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&headers); err != nil {
		return nil, err
	}
	return headers, nil
}

func GenSignature(secret string, timestamp int64, method, path string, body *string) string {
	return auth.SignHMAC(secret, timestamp, method, path, body)
}
