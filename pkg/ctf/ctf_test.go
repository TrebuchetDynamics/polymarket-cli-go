package ctf

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestSplitPositionDataEncodesSelector(t *testing.T) {
	conditionID := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	amount := big.NewInt(1000000)
	partition := []*big.Int{big.NewInt(1), big.NewInt(2)}

	data, err := SplitPositionData(USDC, common.Hash{}, conditionID, partition, amount)
	if err != nil {
		t.Fatal(err)
	}
	selector := crypto.Keccak256Hash([]byte("splitPosition(address,bytes32,bytes32,uint256[],uint256)")).Bytes()[:4]
	if !bytes.Equal(data[:4], selector) {
		t.Fatalf("wrong selector: got %x want %x", data[:4], selector)
	}
}

func TestMergePositionsDataEncodesSelector(t *testing.T) {
	conditionID := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	amount := big.NewInt(1000000)
	partition := []*big.Int{big.NewInt(1), big.NewInt(2)}

	data, err := MergePositionsData(USDC, common.Hash{}, conditionID, partition, amount)
	if err != nil {
		t.Fatal(err)
	}
	selector := crypto.Keccak256Hash([]byte("mergePositions(address,bytes32,bytes32,uint256[],uint256)")).Bytes()[:4]
	if !bytes.Equal(data[:4], selector) {
		t.Fatalf("wrong selector: got %x want %x", data[:4], selector)
	}
}

func TestRedeemPositionsDataEncodesSelector(t *testing.T) {
	conditionID := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	indexSets := []*big.Int{big.NewInt(1), big.NewInt(2)}

	data, err := RedeemPositionsData(USDC, common.Hash{}, conditionID, indexSets)
	if err != nil {
		t.Fatal(err)
	}
	selector := crypto.Keccak256Hash([]byte("redeemPositions(address,bytes32,bytes32,uint256[])")).Bytes()[:4]
	if !bytes.Equal(data[:4], selector) {
		t.Fatalf("wrong selector: got %x want %x", data[:4], selector)
	}
}

func TestPositionIDCalculation(t *testing.T) {
	collateral := USDC
	collectionID := common.HexToHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")
	expected := PositionID(collateral, collectionID)
	if expected == (common.Hash{}) {
		t.Fatal("position ID should not be zero")
	}
}

func TestCollectionIDCalculation(t *testing.T) {
	parent := common.Hash{}
	conditionID := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	indexSet := big.NewInt(1)
	expected := CollectionID(parent, conditionID, indexSet)
	if expected == (common.Hash{}) {
		t.Fatal("collection ID should not be zero")
	}
}
