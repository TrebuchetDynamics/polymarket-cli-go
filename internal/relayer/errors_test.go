package relayer

import (
	"errors"
	"fmt"
	"testing"
)

func TestClassifyAllowlistErrorMatchesUpstreamRejectionStrings(t *testing.T) {
	cases := []struct {
		name    string
		err     error
		blocked bool
	}{
		{
			name:    "setApprovalForAll operator not in the allowed list",
			err:     errors.New(`HTTP 400 https://relayer-v2.polymarket.com/submit: {"error":"call blocked: setApprovalForAll operator 0xAdA100Db00Ca00073811820692005400218FcE1f is not in the allowed list"}`),
			blocked: true,
		},
		{
			name:    "calls to address are not permitted",
			err:     errors.New(`HTTP 400 https://relayer-v2.polymarket.com/submit: {"error":"call blocked: call[0] blocked: calls to 0xAdA100Db00Ca00073811820692005400218FcE1f are not permitted"}`),
			blocked: true,
		},
		{
			name:    "case-insensitive match",
			err:     errors.New("HTTP 400: NOT IN THE ALLOWED LIST"),
			blocked: true,
		},
		{
			name:    "wrapped through fmt.Errorf",
			err:     fmt.Errorf("relayer: WALLET batch: %w", errors.New(`HTTP 400: {"error":"call blocked: are not permitted"}`)),
			blocked: true,
		},
		{
			name:    "server error not classified",
			err:     errors.New("HTTP 500: internal server error"),
			blocked: false,
		},
		{
			name:    "nonce error not classified",
			err:     errors.New("relayer: fetch nonce: HTTP 401: unauthorized"),
			blocked: false,
		},
		{
			name:    "nil passes through",
			err:     nil,
			blocked: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyAllowlistError(tc.err)
			if tc.err == nil {
				if got != nil {
					t.Fatalf("nil err must classify to nil, got %v", got)
				}
				return
			}
			isBlocked := errors.Is(got, ErrRelayerAllowlistBlocked)
			if isBlocked != tc.blocked {
				t.Fatalf("errors.Is(ErrRelayerAllowlistBlocked) = %v, want %v (err=%q)", isBlocked, tc.blocked, got.Error())
			}
			// Underlying error must remain accessible regardless of classification.
			if got.Error() == "" {
				t.Fatalf("classified error has empty message")
			}
		})
	}
}
