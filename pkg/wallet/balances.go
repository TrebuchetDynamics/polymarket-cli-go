package wallet

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	PUSDAddress  = "0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB"
	PUSDDecimals = 6
)

type Balance struct {
	POL       float64
	PUSD      float64
	PUSDRaw   string
}

func GetBalances(ctx context.Context, rpcURL, address string) (*Balance, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("dial rpc: %w", err)
	}
	defer client.Close()

	addr := common.HexToAddress(address)
	
	polWei, err := client.BalanceAt(ctx, addr, nil)
	if err != nil {
		return nil, fmt.Errorf("pol balance: %w", err)
	}
	pol := new(big.Float).Quo(new(big.Float).SetInt(polWei), big.NewFloat(1e18))
	polFloat, _ := pol.Float64()

	pUSDCtr := common.HexToAddress(PUSDAddress)
	data := common.Hex2Bytes("70a08231")
	data = append(data, common.LeftPadBytes(addr.Bytes(), 32)...)
	
	msg := ethereum.CallMsg{To: &pUSDCtr, Data: data}
	result, err := client.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("pusd balance: %w", err)
	}
	
	pUSDRaw := new(big.Int).SetBytes(result)
	pUSDFloat := new(big.Float).Quo(new(big.Float).SetInt(pUSDRaw), big.NewFloat(1e6))
	pusd, _ := pUSDFloat.Float64()

	return &Balance{
		POL:     polFloat,
		PUSD:    pusd,
		PUSDRaw: pUSDRaw.String(),
	}, nil
}
