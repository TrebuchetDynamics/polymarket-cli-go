package web3

import (
	"embed"
	"encoding/json"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

//go:embed abis/*.json
var abiFS embed.FS

// 预加载的 ABI
var (
	USDCABI              abi.ABI
	ConditionalTokensABI abi.ABI
	CTFExchangeABI       abi.ABI
	NegRiskExchangeABI   abi.ABI
	NegRiskAdapterABI    abi.ABI
	ProxyFactoryABI      abi.ABI
	SafeProxyFactoryABI  abi.ABI
	SafeABI              abi.ABI
)

func init() {
	var err error

	USDCABI, err = loadABI("UChildERC20Proxy")
	if err != nil {
		panic("failed to load USDC ABI: " + err.Error())
	}

	ConditionalTokensABI, err = loadABI("ConditionalTokens")
	if err != nil {
		panic("failed to load ConditionalTokens ABI: " + err.Error())
	}

	CTFExchangeABI, err = loadABI("CTFExchange")
	if err != nil {
		panic("failed to load CTFExchange ABI: " + err.Error())
	}

	NegRiskExchangeABI, err = loadABI("NegRiskCtfExchange")
	if err != nil {
		panic("failed to load NegRiskCtfExchange ABI: " + err.Error())
	}

	NegRiskAdapterABI, err = loadABI("NegRiskAdapter")
	if err != nil {
		panic("failed to load NegRiskAdapter ABI: " + err.Error())
	}

	ProxyFactoryABI, err = loadABI("ProxyWalletFactory")
	if err != nil {
		panic("failed to load ProxyWalletFactory ABI: " + err.Error())
	}

	SafeProxyFactoryABI, err = loadABI("SafeProxyFactory")
	if err != nil {
		panic("failed to load SafeProxyFactory ABI: " + err.Error())
	}

	SafeABI, err = loadABI("Safe")
	if err != nil {
		panic("failed to load Safe ABI: " + err.Error())
	}
}

func loadABI(contractName string) (abi.ABI, error) {
	data, err := abiFS.ReadFile("abis/" + contractName + ".json")
	if err != nil {
		return abi.ABI{}, err
	}

	var abiJSON []interface{}
	if err := json.Unmarshal(data, &abiJSON); err != nil {
		return abi.ABI{}, err
	}

	// 重新序列化为标准格式
	abiBytes, err := json.Marshal(abiJSON)
	if err != nil {
		return abi.ABI{}, err
	}

	return abi.JSON(strings.NewReader(string(abiBytes)))
}

