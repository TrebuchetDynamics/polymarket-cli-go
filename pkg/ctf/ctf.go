// Package ctf provides utilities for Conditional Token Framework (CTF)
// on-chain operations on Polymarket. It includes contract addresses,
// transaction data encoding for split/merge/redeem, and CTF ID calculations.
package ctf

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Polygon Mainnet contract addresses.
var (
	ConditionalTokens = common.HexToAddress("0x4D97DCd97eC945f40cF65F87097ACe5EA0476045")
	NegRiskAdapter    = common.HexToAddress("0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296")
	USDC              = common.HexToAddress("0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174")
)

func mustType(t string) abi.Type {
	ty, err := abi.NewType(t, "", nil)
	if err != nil {
		panic(err)
	}
	return ty
}

var (
	splitPositionArgs = abi.Arguments{
		{Type: mustType("address")},
		{Type: mustType("bytes32")},
		{Type: mustType("bytes32")},
		{Type: mustType("uint256[]")},
		{Type: mustType("uint256")},
	}
	mergePositionsArgs = abi.Arguments{
		{Type: mustType("address")},
		{Type: mustType("bytes32")},
		{Type: mustType("bytes32")},
		{Type: mustType("uint256[]")},
		{Type: mustType("uint256")},
	}
	redeemPositionsArgs = abi.Arguments{
		{Type: mustType("address")},
		{Type: mustType("bytes32")},
		{Type: mustType("bytes32")},
		{Type: mustType("uint256[]")},
	}
)

func selector(sig string) []byte {
	return crypto.Keccak256Hash([]byte(sig)).Bytes()[:4]
}

var (
	splitPositionSelector   = selector("splitPosition(address,bytes32,bytes32,uint256[],uint256)")
	mergePositionsSelector  = selector("mergePositions(address,bytes32,bytes32,uint256[],uint256)")
	redeemPositionsSelector = selector("redeemPositions(address,bytes32,bytes32,uint256[])")
)

// SplitPositionData returns the ABI-encoded transaction data for splitting
// a collateral token into conditional tokens.
func SplitPositionData(collateralToken common.Address, parentCollectionID common.Hash, conditionID common.Hash, partition []*big.Int, amount *big.Int) ([]byte, error) {
	data, err := splitPositionArgs.Pack(collateralToken, parentCollectionID, conditionID, partition, amount)
	if err != nil {
		return nil, err
	}
	return append(splitPositionSelector, data...), nil
}

// MergePositionsData returns the ABI-encoded transaction data for merging
// conditional tokens back into collateral.
func MergePositionsData(collateralToken common.Address, parentCollectionID common.Hash, conditionID common.Hash, partition []*big.Int, amount *big.Int) ([]byte, error) {
	data, err := mergePositionsArgs.Pack(collateralToken, parentCollectionID, conditionID, partition, amount)
	if err != nil {
		return nil, err
	}
	return append(mergePositionsSelector, data...), nil
}

// RedeemPositionsData returns the ABI-encoded transaction data for redeeming
// winning conditional tokens for collateral after market resolution.
func RedeemPositionsData(collateralToken common.Address, parentCollectionID common.Hash, conditionID common.Hash, indexSets []*big.Int) ([]byte, error) {
	data, err := redeemPositionsArgs.Pack(collateralToken, parentCollectionID, conditionID, indexSets)
	if err != nil {
		return nil, err
	}
	return append(redeemPositionsSelector, data...), nil
}

// PositionID calculates the CTF position ID from collateral token and collection ID.
func PositionID(collateralToken common.Address, collectionID common.Hash) common.Hash {
	return crypto.Keccak256Hash(collateralToken.Bytes(), collectionID.Bytes())
}

// CollectionID calculates the CTF collection ID from parent collection ID,
// condition ID, and index set.
func CollectionID(parentCollectionID common.Hash, conditionID common.Hash, indexSet *big.Int) common.Hash {
	return crypto.Keccak256Hash(parentCollectionID.Bytes(), conditionID.Bytes(), indexSet.Bytes())
}
