package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"
	"strings"

	"github.com/TrebuchetDynamics/polygolem/internal/auth"
	"github.com/TrebuchetDynamics/polygolem/internal/clob"
	"github.com/spf13/cobra"
)

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}

func parseClobTokenIDs(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var ids []string
	if err := json.Unmarshal([]byte(raw), &ids); err != nil {
		return []string{raw}
	}
	return ids
}

type batchOrderJSON struct {
	Token          string `json:"token"`
	TokenID        string `json:"tokenID"`
	TokenIDSnake   string `json:"token_id"`
	Side           string `json:"side"`
	Price          string `json:"price"`
	Size           string `json:"size"`
	OrderType      string `json:"orderType"`
	OrderTypeSnake string `json:"order_type"`
	Expiration     string `json:"expiration"`
	PostOnly       bool   `json:"postOnly"`
	PostOnlySnake  bool   `json:"post_only"`
}

func openBatchOrdersInput(cmd *cobra.Command, path string) (io.Reader, func(), error) {
	if strings.TrimSpace(path) == "-" {
		return cmd.InOrStdin(), nil, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	return file, func() { _ = file.Close() }, nil
}

func parseBatchOrderParams(r io.Reader) ([]clob.CreateOrderParams, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var inputs []batchOrderJSON
	if err := json.Unmarshal(raw, &inputs); err != nil {
		var wrapped struct {
			Orders []batchOrderJSON `json:"orders"`
		}
		if wrappedErr := json.Unmarshal(raw, &wrapped); wrappedErr != nil {
			return nil, err
		}
		inputs = wrapped.Orders
	}
	if len(inputs) == 0 {
		return nil, fmt.Errorf("orders file must contain at least one order")
	}
	out := make([]clob.CreateOrderParams, len(inputs))
	for i, in := range inputs {
		out[i] = clob.CreateOrderParams{
			TokenID:    firstNonEmptyCLI(in.TokenID, in.TokenIDSnake, in.Token),
			Side:       in.Side,
			Price:      in.Price,
			Size:       in.Size,
			OrderType:  firstNonEmptyCLI(in.OrderType, in.OrderTypeSnake),
			Expiration: in.Expiration,
			PostOnly:   in.PostOnly || in.PostOnlySnake,
		}
	}
	return out, nil
}

func firstNonEmptyCLI(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func builderCodeFromFlagOrEnv(flagValue string) string {
	if value := strings.TrimSpace(flagValue); value != "" {
		return value
	}
	return firstEnv("POLYMARKET_BUILDER_CODE", "POLYMARKET_CLOB_BUILDER_CODE")
}

func privateKeyFromEnv() (string, error) {
	key := strings.TrimSpace(os.Getenv("POLYMARKET_PRIVATE_KEY"))
	if key == "" {
		return "", fmt.Errorf("POLYMARKET_PRIVATE_KEY is required")
	}
	return key, nil
}

func clobL2CredentialsFromEnv() (auth.APIKey, bool) {
	key := auth.APIKey{
		Key:        firstEnv("POLYMARKET_CLOB_API_KEY", "CLOB_API_KEY"),
		Secret:     firstEnv("POLYMARKET_CLOB_SECRET", "CLOB_SECRET"),
		Passphrase: firstEnv("POLYMARKET_CLOB_PASSPHRASE", "CLOB_PASSPHRASE", "CLOB_PASS_PHRASE"),
	}
	if strings.TrimSpace(key.Key) == "" &&
		strings.TrimSpace(key.Secret) == "" &&
		strings.TrimSpace(key.Passphrase) == "" {
		return auth.APIKey{}, false
	}
	return key, true
}

func validateBuilderCodeForCLI(builderCode string) error {
	value := strings.TrimSpace(builderCode)
	if value == "" {
		return nil
	}
	if !strings.HasPrefix(value, "0x") {
		return fmt.Errorf("builder code must be a 0x-prefixed bytes32 hex string")
	}
	hexValue := value[2:]
	if len(hexValue) != 64 {
		return fmt.Errorf("builder code must be 32 bytes, got %d hex characters", len(hexValue))
	}
	if _, err := hex.DecodeString(hexValue); err != nil {
		return fmt.Errorf("builder code must be hex: %w", err)
	}
	return nil
}

func balanceResponseMap(res *clob.BalanceAllowanceResponse) map[string]interface{} {
	out := map[string]interface{}{}
	if res == nil {
		return out
	}
	out["balance"] = res.Balance
	if len(res.Allowances) > 0 {
		out["allowances"] = res.Allowances
	}
	if strings.TrimSpace(res.Allowance) != "" {
		out["allowance"] = res.Allowance
	}
	return out
}

func normalizeCollateralBalanceResponse(raw map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(raw))
	for key, value := range raw {
		out[key] = value
	}
	balance, ok := out["balance"].(string)
	if !ok {
		return out
	}
	if scaled, ok := scaleBaseUnits(balance, 6); ok {
		out["balance"] = scaled
	}
	return out
}

func scaleBaseUnits(value string, decimals int) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "" || strings.Contains(value, ".") || decimals <= 0 {
		return value, false
	}
	n := new(big.Int)
	if _, ok := n.SetString(value, 10); !ok {
		return value, false
	}
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	whole := new(big.Int).Quo(new(big.Int).Set(n), scale)
	frac := new(big.Int).Mod(new(big.Int).Set(n), scale).String()
	if len(frac) < decimals {
		frac = strings.Repeat("0", decimals-len(frac)) + frac
	}
	return whole.String() + "." + frac, true
}
