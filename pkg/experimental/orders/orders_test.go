package orders

import "testing"

func TestOrderIntentValidate(t *testing.T) {
	tests := []struct {
		name    string
		intent  OrderIntent
		wantErr bool
	}{
		{
			name: "valid limit buy",
			intent: OrderIntent{
				TokenID:   "123",
				Side:      SideBuy,
				Price:     "0.5",
				Size:      "10",
				TickSize:  "0.01",
				OrderType: OrderTypeGTC,
			},
			wantErr: false,
		},
		{
			name: "missing token_id",
			intent: OrderIntent{
				Side:     SideBuy,
				Price:    "0.5",
				Size:     "10",
				TickSize: "0.01",
			},
			wantErr: true,
		},
		{
			name: "invalid side",
			intent: OrderIntent{
				TokenID:   "123",
				Side:      "HOLD",
				Price:     "0.5",
				Size:      "10",
				TickSize:  "0.01",
				OrderType: OrderTypeGTC,
			},
			wantErr: true,
		},
		{
			name: "missing price and size",
			intent: OrderIntent{
				TokenID:   "123",
				Side:      SideBuy,
				TickSize:  "0.01",
				OrderType: OrderTypeGTC,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.intent.Validate()
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
