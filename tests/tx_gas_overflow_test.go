// Modifications Copyright 2024 The Kaia Authors
// Copyright 2019 The klaytn Authors
// This file is part of the klaytn library.
//
// The klaytn library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The klaytn library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the klaytn library. If not, see <http://www.gnu.org/licenses/>.
// Modified and improved for the Kaia development.

package tests

import (
	"testing"

	"github.com/kaiachain/kaia/blockchain"
	"github.com/kaiachain/kaia/blockchain/types"
	"github.com/kaiachain/kaia/blockchain/types/accountkey"
	"github.com/kaiachain/kaia/common/math"
	"github.com/kaiachain/kaia/params"
)

type overflowCheckFunc func(t *testing.T)

func TestGasOverflow(t *testing.T) {
	testFunctions := []struct {
		Name                 string
		gasOverflowCheckFunc overflowCheckFunc
	}{
		{"LegacyTransaction", testGasOverflowLegacyTransaction},

		{"ValueTransfer", testGasOverflowValueTransfer},
		{"FeeDelegatedValueTransfer", testGasOverflowFeeDelegatedValueTransfer},
		{"FeeDelegatedWithRatioValueTransfer", testGasOverflowFeeDelegatedWithRatioValueTransfer},

		{"ValueTransferWithMemo", testGasOverflowValueTransferWithMemo},
		{"FeeDelegatedValueTransferWithMemo", testGasOverflowFeeDelegatedValueTransferWithMemo},
		{"FeeDelegatedWithRatioValueTransferWithMemo", testGasOverflowFeeDelegatedWithRatioValueTransferWithMemo},

		{"AccountUpdate", testGasOverflowAccountUpdate},
		{"FeeDelegatedAccountUpdate", testGasOverflowFeeDelegatedAccountUpdate},
		{"FeeDelegatedWithRatioAccountUpdate", testGasOverflowFeeDelegatedWithRatioAccountUpdate},

		{"SmartContractDeploy", testGasOverflowSmartContractDeploy},
		{"FeeDelegatedSmartContractDeploy", testGasOverflowFeeDelegatedSmartContractDeploy},
		{"FeeDelegatedWithRatioSmartContractDeploy", testGasOverflowFeeDelegatedWithRatioSmartContractDeploy},

		{"SmartContractExecution", testGasOverflowSmartContractExecution},
		{"FeeDelegatedSmartContractExecution", testGasOverflowFeeDelegatedSmartContractExecution},
		{"FeeDelegatedWithRatioSmartContractExecution", testGasOverflowFeeDelegatedWithRatioSmartContractExecution},

		{"Cancel", testGasOverflowCancel},
		{"FeeDelegatedCancel", testGasOverflowFeeDelegatedCancel},
		{"FeeDelegatedWithRatioCancel", testGasOverflowFeeDelegatedWithRatioCancel},

		{"ChainDataAnchoring", testGasOverflowChainDataAnchoring},
		{"FeeDelegatedChainDataAnchoring", testGasOverflowFeeDelegatedChainDataAnchoring},
		{"FeeDelegatedWithRatioChainDataAnchoring", testGasOverflowFeeDelegatedWithRatioChainDataAnchoring},
	}

	for _, f := range testFunctions {
		t.Run(f.Name, func(t *testing.T) {
			f.gasOverflowCheckFunc(t)
		})
	}
}

func testGasOverflowLegacyTransaction(t *testing.T) {
	intrinsic, _ := types.GetTxGasForTxType(types.TxTypeLegacyTransaction)
	senderValidationGas := getMaxValidationKeyGas(t)

	maxDataGas := mulUint64(t, blockchain.MaxTxDataSize, params.TxDataNonZeroGasFrontier)

	gas := addUint64(t, intrinsic, senderValidationGas)
	gas = addUint64(t, gas, maxDataGas)
}

func testGasOverflowValueTransfer(t *testing.T) {
	intrinsic, _ := types.GetTxGasForTxType(types.TxTypeValueTransfer)
	senderValidationGas := getMaxValidationKeyGas(t)

	addUint64(t, intrinsic, senderValidationGas)
}

func testGasOverflowFeeDelegatedValueTransfer(t *testing.T) {
	intrinsic, _ := types.GetTxGasForTxType(types.TxTypeFeeDelegatedValueTransfer)
	senderValidationGas := getMaxValidationKeyGas(t)
	payerValidationGas := getMaxValidationKeyGas(t)

	gas := addUint64(t, intrinsic, senderValidationGas)
	gas = addUint64(t, gas, payerValidationGas)
}

func testGasOverflowFeeDelegatedWithRatioValueTransfer(t *testing.T) {
	intrinsic, _ := types.GetTxGasForTxType(types.TxTypeFeeDelegatedValueTransferWithRatio)
	senderValidationGas := getMaxValidationKeyGas(t)
	payerValidationGas := getMaxValidationKeyGas(t)

	gas := addUint64(t, intrinsic, senderValidationGas)
	gas = addUint64(t, gas, payerValidationGas)
}

func testGasOverflowValueTransferWithMemo(t *testing.T) {
	intrinsic, _ := types.GetTxGasForTxType(types.TxTypeValueTransferMemo)
	senderValidationGas := getMaxValidationKeyGas(t)

	maxDataGas := mulUint64(t, blockchain.MaxTxDataSize, params.TxDataGas)

	gas := addUint64(t, intrinsic, senderValidationGas)
	gas = addUint64(t, gas, maxDataGas)
}

func testGasOverflowFeeDelegatedValueTransferWithMemo(t *testing.T) {
	intrinsic, _ := types.GetTxGasForTxType(types.TxTypeFeeDelegatedValueTransferMemo)
	senderValidationGas := getMaxValidationKeyGas(t)
	payerValidationGas := getMaxValidationKeyGas(t)

	maxDataGas := mulUint64(t, blockchain.MaxTxDataSize, params.TxDataGas)

	gas := addUint64(t, intrinsic, senderValidationGas)
	gas = addUint64(t, gas, payerValidationGas)
	gas = addUint64(t, gas, maxDataGas)
}

func testGasOverflowFeeDelegatedWithRatioValueTransferWithMemo(t *testing.T) {
	intrinsic, _ := types.GetTxGasForTxType(types.TxTypeFeeDelegatedValueTransferMemoWithRatio)
	senderValidationGas := getMaxValidationKeyGas(t)
	payerValidationGas := getMaxValidationKeyGas(t)

	maxDataGas := mulUint64(t, blockchain.MaxTxDataSize, params.TxDataGas)

	gas := addUint64(t, intrinsic, senderValidationGas)
	gas = addUint64(t, gas, payerValidationGas)
	gas = addUint64(t, gas, maxDataGas)
}

func testGasOverflowAccountUpdate(t *testing.T) {
	intrinsic, _ := types.GetTxGasForTxType(types.TxTypeAccountUpdate)
	senderValidationGas := getMaxValidationKeyGas(t)

	maxCreationGas := getMaxCreationKeyGas(t)

	gas := addUint64(t, intrinsic, senderValidationGas)
	gas = addUint64(t, gas, maxCreationGas)
}

func testGasOverflowFeeDelegatedAccountUpdate(t *testing.T) {
	intrinsic, _ := types.GetTxGasForTxType(types.TxTypeFeeDelegatedAccountUpdate)
	senderValidationGas := getMaxValidationKeyGas(t)
	payerValidationGas := getMaxValidationKeyGas(t)

	maxCreationGas := getMaxCreationKeyGas(t)

	gas := addUint64(t, intrinsic, senderValidationGas)
	gas = addUint64(t, gas, payerValidationGas)
	gas = addUint64(t, gas, maxCreationGas)
}

func testGasOverflowFeeDelegatedWithRatioAccountUpdate(t *testing.T) {
	intrinsic, _ := types.GetTxGasForTxType(types.TxTypeFeeDelegatedAccountUpdateWithRatio)
	senderValidationGas := getMaxValidationKeyGas(t)
	payerValidationGas := getMaxValidationKeyGas(t)

	maxCreationGas := getMaxCreationKeyGas(t)

	gas := addUint64(t, intrinsic, senderValidationGas)
	gas = addUint64(t, gas, payerValidationGas)
	gas = addUint64(t, gas, maxCreationGas)
}

func testGasOverflowSmartContractDeploy(t *testing.T) {
	intrinsic, _ := types.GetTxGasForTxType(types.TxTypeSmartContractDeploy)
	senderValidationGas := getMaxValidationKeyGas(t)

	payloadGas := mulUint64(t, blockchain.MaxTxDataSize, params.TxDataGas)

	humanReadableGas := params.TxGasHumanReadable

	gas := addUint64(t, intrinsic, senderValidationGas)
	gas = addUint64(t, gas, payloadGas)
	gas = addUint64(t, gas, humanReadableGas)
}

func testGasOverflowFeeDelegatedSmartContractDeploy(t *testing.T) {
	intrinsic, _ := types.GetTxGasForTxType(types.TxTypeFeeDelegatedSmartContractDeploy)
	senderValidationGas := getMaxValidationKeyGas(t)
	payerValidationGas := getMaxValidationKeyGas(t)

	payloadGas := mulUint64(t, blockchain.MaxTxDataSize, params.TxDataGas)

	humanReadableGas := params.TxGasHumanReadable

	gas := addUint64(t, intrinsic, senderValidationGas)
	gas = addUint64(t, gas, payerValidationGas)
	gas = addUint64(t, gas, payloadGas)
	gas = addUint64(t, gas, humanReadableGas)
}

func testGasOverflowFeeDelegatedWithRatioSmartContractDeploy(t *testing.T) {
	intrinsic, _ := types.GetTxGasForTxType(types.TxTypeFeeDelegatedSmartContractDeployWithRatio)
	senderValidationGas := getMaxValidationKeyGas(t)
	payerValidationGas := getMaxValidationKeyGas(t)

	payloadGas := mulUint64(t, blockchain.MaxTxDataSize, params.TxDataGas)

	humanReadableGas := params.TxGasHumanReadable

	gas := addUint64(t, intrinsic, senderValidationGas)
	gas = addUint64(t, gas, payerValidationGas)
	gas = addUint64(t, gas, payloadGas)
	gas = addUint64(t, gas, humanReadableGas)
}

func testGasOverflowSmartContractExecution(t *testing.T) {
	intrinsic, _ := types.GetTxGasForTxType(types.TxTypeSmartContractExecution)
	senderValidationGas := getMaxValidationKeyGas(t)

	payloadGas := mulUint64(t, blockchain.MaxTxDataSize, params.TxDataGas)

	gas := addUint64(t, intrinsic, senderValidationGas)
	gas = addUint64(t, gas, payloadGas)
}

func testGasOverflowFeeDelegatedSmartContractExecution(t *testing.T) {
	intrinsic, _ := types.GetTxGasForTxType(types.TxTypeFeeDelegatedSmartContractExecution)
	senderValidationGas := getMaxValidationKeyGas(t)
	payerValidationGas := getMaxValidationKeyGas(t)

	payloadGas := mulUint64(t, blockchain.MaxTxDataSize, params.TxDataGas)

	gas := addUint64(t, intrinsic, senderValidationGas)
	gas = addUint64(t, gas, payerValidationGas)
	gas = addUint64(t, gas, payloadGas)
}

func testGasOverflowFeeDelegatedWithRatioSmartContractExecution(t *testing.T) {
	intrinsic, _ := types.GetTxGasForTxType(types.TxTypeFeeDelegatedSmartContractExecutionWithRatio)
	senderValidationGas := getMaxValidationKeyGas(t)
	payerValidationGas := getMaxValidationKeyGas(t)

	payloadGas := mulUint64(t, blockchain.MaxTxDataSize, params.TxDataGas)

	gas := addUint64(t, intrinsic, senderValidationGas)
	gas = addUint64(t, gas, payerValidationGas)
	gas = addUint64(t, gas, payloadGas)
}

func testGasOverflowCancel(t *testing.T) {
	intrinsic, _ := types.GetTxGasForTxType(types.TxTypeCancel)
	senderValidationGas := getMaxValidationKeyGas(t)

	addUint64(t, intrinsic, senderValidationGas)
}

func testGasOverflowFeeDelegatedCancel(t *testing.T) {
	intrinsic, _ := types.GetTxGasForTxType(types.TxTypeFeeDelegatedCancel)
	senderValidationGas := getMaxValidationKeyGas(t)
	payerValidationGas := getMaxValidationKeyGas(t)

	gas := addUint64(t, intrinsic, senderValidationGas)
	gas = addUint64(t, gas, payerValidationGas)
}

func testGasOverflowFeeDelegatedWithRatioCancel(t *testing.T) {
	intrinsic, _ := types.GetTxGasForTxType(types.TxTypeFeeDelegatedCancelWithRatio)
	senderValidationGas := getMaxValidationKeyGas(t)
	payerValidationGas := getMaxValidationKeyGas(t)

	gas := addUint64(t, intrinsic, senderValidationGas)
	gas = addUint64(t, gas, payerValidationGas)
}

func testGasOverflowChainDataAnchoring(t *testing.T) {
	intrinsic, _ := types.GetTxGasForTxType(types.TxTypeChainDataAnchoring)
	senderValidationGas := getMaxValidationKeyGas(t)

	maxDataGas := mulUint64(t, blockchain.MaxTxDataSize, params.TxDataGas)

	gas := addUint64(t, intrinsic, senderValidationGas)
	gas = addUint64(t, gas, maxDataGas)
}

func testGasOverflowFeeDelegatedChainDataAnchoring(t *testing.T) {
	intrinsic, _ := types.GetTxGasForTxType(types.TxTypeFeeDelegatedChainDataAnchoring)
	senderValidationGas := getMaxValidationKeyGas(t)
	payerValidationGas := getMaxValidationKeyGas(t)

	maxDataGas := mulUint64(t, blockchain.MaxTxDataSize, params.TxDataGas)

	gas := addUint64(t, intrinsic, senderValidationGas)
	gas = addUint64(t, gas, payerValidationGas)
	gas = addUint64(t, gas, maxDataGas)
}

func testGasOverflowFeeDelegatedWithRatioChainDataAnchoring(t *testing.T) {
	intrinsic, _ := types.GetTxGasForTxType(types.TxTypeFeeDelegatedChainDataAnchoringWithRatio)
	senderValidationGas := getMaxValidationKeyGas(t)
	payerValidationGas := getMaxValidationKeyGas(t)

	maxDataGas := mulUint64(t, blockchain.MaxTxDataSize, params.TxDataGas)

	gas := addUint64(t, intrinsic, senderValidationGas)
	gas = addUint64(t, gas, payerValidationGas)
	gas = addUint64(t, gas, maxDataGas)
}

func getMaxValidationKeyGas(t *testing.T) uint64 {
	return mulUint64(t, uint64(accountkey.MaxNumKeysForMultiSig), params.TxValidationGasPerKey)
}

func getMaxCreationKeyGas(t *testing.T) uint64 {
	txKeyGas := mulUint64(t, uint64(accountkey.MaxNumKeysForMultiSig), params.TxAccountCreationGasPerKey)
	updateKeyGas := mulUint64(t, uint64(accountkey.MaxNumKeysForMultiSig), params.TxAccountCreationGasPerKey)
	feeKeysGas := mulUint64(t, uint64(accountkey.MaxNumKeysForMultiSig), params.TxAccountCreationGasPerKey)

	creationKey := addUint64(t, txKeyGas, updateKeyGas)
	creationKey = addUint64(t, creationKey, feeKeysGas)

	return creationKey
}

func addUint64(t *testing.T, a uint64, b uint64) uint64 {
	c, overflow := math.SafeAdd(a, b)
	if overflow {
		t.Error("gas overflow ", a, "+", b)
	}
	return c
}

func mulUint64(t *testing.T, a uint64, b uint64) uint64 {
	c, overflow := math.SafeMul(a, b)
	if overflow {
		t.Error("gas overflow ", a, "*", b)
	}
	return c
}
