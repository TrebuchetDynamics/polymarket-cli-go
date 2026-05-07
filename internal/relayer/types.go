package relayer

// RelayerError is a structured error returned by the relayer API.
type RelayerError struct {
	Error string `json:"error"`
	Code  int    `json:"code,omitempty"`
}

// RelayerTransactionState maps the relayer's lifecycle states.
type RelayerTransactionState string

const (
	StateNew       RelayerTransactionState = "STATE_NEW"
	StateExecuted  RelayerTransactionState = "STATE_EXECUTED"
	StateMined     RelayerTransactionState = "STATE_MINED"
	StateInvalid   RelayerTransactionState = "STATE_INVALID"
	StateConfirmed RelayerTransactionState = "STATE_CONFIRMED"
	StateFailed    RelayerTransactionState = "STATE_FAILED"
)

func (s RelayerTransactionState) IsTerminal() bool {
	switch s {
	case StateMined, StateConfirmed, StateFailed, StateInvalid:
		return true
	}
	return false
}

func (s RelayerTransactionState) IsSuccess() bool {
	return s == StateMined || s == StateConfirmed
}

// DepositWalletCall is a single call within a WALLET batch.
type DepositWalletCall struct {
	Target string `json:"target"`
	Value  string `json:"value"`
	Data   string `json:"data"`
}

type WalletCreateRequest struct {
	Type string `json:"type"`
	From string `json:"from"`
	To   string `json:"to"`
}

// depositWalletParams is the nested params object within a WALLET batch.
type depositWalletParams struct {
	DepositWallet string              `json:"depositWallet"`
	Deadline      string              `json:"deadline"`
	Calls         []DepositWalletCall `json:"calls"`
}

// WalletBatchRequest is the payload for POST /submit with type=WALLET.
type WalletBatchRequest struct {
	Type                string              `json:"type"`
	From                string              `json:"from"`
	To                  string              `json:"to"`
	Nonce               string              `json:"nonce"`
	Signature           string              `json:"signature"`
	DepositWalletParams depositWalletParams `json:"depositWalletParams"`
}

// RelayerTransaction represents a submitted transaction tracked by the relayer.
type RelayerTransaction struct {
	TransactionID   string `json:"transactionID"`
	TransactionHash string `json:"transactionHash,omitempty"`
	From            string `json:"from"`
	To              string `json:"to"`
	ProxyAddress    string `json:"proxyAddress,omitempty"`
	Data            string `json:"data,omitempty"`
	Nonce           string `json:"nonce,omitempty"`
	Value           string `json:"value,omitempty"`
	State           string `json:"state"`
	Type            string `json:"type"`
	Metadata        string `json:"metadata,omitempty"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

// NonceResponse is the response from GET /nonce?address=...&type=WALLET.
type NonceResponse struct {
	Nonce string `json:"nonce"`
}

// DeployedResponse is the response from GET /deployed?address=...
type DeployedResponse struct {
	Deployed bool   `json:"deployed"`
	Address  string `json:"address,omitempty"`
}

// WalletCreateResponse is returned when WALLET-CREATE is accepted by the relayer.
// The relayer returns a transaction object with the created wallet address.
type WalletCreateResponse struct {
	TransactionID string `json:"transactionID"`
	WalletAddress string `json:"proxyAddress"`
	State         string `json:"state"`
}
