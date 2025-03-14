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

package types

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/kaiachain/kaia/blockchain/types/account"
	"github.com/kaiachain/kaia/blockchain/types/accountkey"
	"github.com/kaiachain/kaia/common"
	"github.com/kaiachain/kaia/crypto"
	"github.com/kaiachain/kaia/kerrors"
	"github.com/kaiachain/kaia/params"
	"github.com/kaiachain/kaia/rlp"
)

// MaxFeeRatio is the maximum value of feeRatio. Since it is represented in percentage,
// the maximum value is 100.
const MaxFeeRatio FeeRatio = 100

const SubTxTypeBits uint = 3

var emptyCodeHash = crypto.Keccak256(nil)

type TxType uint16

const (
	// TxType declarations
	// There are three type declarations at each line:
	//   <base type>, <fee-delegated type>, and <fee-delegated type with a fee ratio>
	// If types other than <base type> are not useful, they are declared with underscore(_).
	// Each base type is self-descriptive.
	TxTypeLegacyTransaction, _, _ TxType = iota << SubTxTypeBits, iota<<SubTxTypeBits + 1, iota<<SubTxTypeBits + 2
	TxTypeValueTransfer, TxTypeFeeDelegatedValueTransfer, TxTypeFeeDelegatedValueTransferWithRatio
	TxTypeValueTransferMemo, TxTypeFeeDelegatedValueTransferMemo, TxTypeFeeDelegatedValueTransferMemoWithRatio
	TxTypeAccountCreation, _, _
	TxTypeAccountUpdate, TxTypeFeeDelegatedAccountUpdate, TxTypeFeeDelegatedAccountUpdateWithRatio
	TxTypeSmartContractDeploy, TxTypeFeeDelegatedSmartContractDeploy, TxTypeFeeDelegatedSmartContractDeployWithRatio
	TxTypeSmartContractExecution, TxTypeFeeDelegatedSmartContractExecution, TxTypeFeeDelegatedSmartContractExecutionWithRatio
	TxTypeCancel, TxTypeFeeDelegatedCancel, TxTypeFeeDelegatedCancelWithRatio
	TxTypeBatch, _, _
	TxTypeChainDataAnchoring, TxTypeFeeDelegatedChainDataAnchoring, TxTypeFeeDelegatedChainDataAnchoringWithRatio
	TxTypeKaiaLast, _, _
	TxTypeEthereumAccessList = TxType(0x7801)
	TxTypeEthereumDynamicFee = TxType(0x7802)
	// EIP-4844 BLOB_TX_TYPE not supported in Kaia.
	_                     = TxType(0x7803)
	TxTypeEthereumSetCode = TxType(0x7804)
	TxTypeEthereumLast    = TxType(0x7805)
)

type TxValueKeyType uint

const EthereumTxTypeEnvelope = TxType(0x78)

const (
	TxValueKeyNonce TxValueKeyType = iota
	TxValueKeyTo
	TxValueKeyAmount
	TxValueKeyGasLimit
	TxValueKeyGasPrice
	TxValueKeyData
	TxValueKeyFrom
	TxValueKeyAnchoredData
	TxValueKeyHumanReadable
	TxValueKeyAccountKey
	TxValueKeyFeePayer
	TxValueKeyFeeRatioOfFeePayer
	TxValueKeyCodeFormat
	TxValueKeyAccessList
	TxValueKeyChainID
	TxValueKeyGasTipCap
	TxValueKeyGasFeeCap
	TxValueKeyAuthorizationList
)

type TxTypeMask uint8

const (
	TxFeeDelegationBitMask          TxTypeMask = 1
	TxFeeDelegationWithRatioBitMask TxTypeMask = 2
)

var (
	errNotTxTypeValueTransfer                 = errors.New("not value transfer transaction type")
	errNotTxTypeValueTransferWithFeeDelegator = errors.New("not a fee-delegated value transfer transaction")
	errNotTxTypeAccountCreation               = errors.New("not account creation transaction type")
	errUndefinedTxType                        = errors.New("undefined tx type")
	errCannotBeSignedByFeeDelegator           = errors.New("this transaction type cannot be signed by a fee delegator")
	errUndefinedKeyRemains                    = errors.New("undefined key remains")

	errValueKeyHumanReadableMustBool     = errors.New("HumanReadable must be a type of bool")
	errValueKeyAccountKeyMustAccountKey  = errors.New("AccountKey must be a type of AccountKey")
	errValueKeyAnchoredDataMustByteSlice = errors.New("AnchoredData must be a slice of bytes")
	errValueKeyNonceMustUint64           = errors.New("Nonce must be a type of uint64")
	errValueKeyToMustAddress             = errors.New("To must be a type of common.Address")
	errValueKeyToMustAddressPointer      = errors.New("To must be a type of *common.Address")
	errValueKeyAmountMustBigInt          = errors.New("Amount must be a type of *big.Int")
	errValueKeyGasLimitMustUint64        = errors.New("GasLimit must be a type of uint64")
	errValueKeyGasPriceMustBigInt        = errors.New("GasPrice must be a type of *big.Int")
	errValueKeyFromMustAddress           = errors.New("From must be a type of common.Address")
	errValueKeyFeePayerMustAddress       = errors.New("FeePayer must be a type of common.Address")
	errValueKeyDataMustByteSlice         = errors.New("Data must be a slice of bytes")
	errValueKeyFeeRatioMustUint8         = errors.New("FeeRatio must be a type of uint8")
	errValueKeyCodeFormatInvalid         = errors.New("The smart contract code format is invalid")
	errValueKeyAccessListInvalid         = errors.New("AccessList must be a type of AccessList")
	errValueKeyAuthorizationListInvalid  = errors.New("AuthorizationList must be a type of AuthorizationList")
	errValueKeyChainIDInvalid            = errors.New("ChainID must be a type of ChainID")
	errValueKeyGasTipCapMustBigInt       = errors.New("GasTipCap must be a type of *big.Int")
	errValueKeyGasFeeCapMustBigInt       = errors.New("GasFeeCap must be a type of *big.Int")

	ErrTxTypeNotSupported         = errors.New("transaction type not supported")
	ErrSenderPubkeyNotSupported   = errors.New("SenderPubkey is not supported for this signer")
	ErrSenderFeePayerNotSupported = errors.New("SenderFeePayer is not supported for this signer")
	ErrHashFeePayerNotSupported   = errors.New("HashFeePayer is not supported for this signer")

	// ErrGasUintOverflow is returned when calculating gas usage.
	ErrGasUintOverflow = errors.New("gas uint64 overflow")
)

func (t TxValueKeyType) String() string {
	switch t {
	case TxValueKeyNonce:
		return "TxValueKeyNonce"
	case TxValueKeyTo:
		return "TxValueKeyTo"
	case TxValueKeyAmount:
		return "TxValueKeyAmount"
	case TxValueKeyGasLimit:
		return "TxValueKeyGasLimit"
	case TxValueKeyGasPrice:
		return "TxValueKeyGasPrice"
	case TxValueKeyData:
		return "TxValueKeyData"
	case TxValueKeyFrom:
		return "TxValueKeyFrom"
	case TxValueKeyAnchoredData:
		return "TxValueKeyAnchoredData"
	case TxValueKeyHumanReadable:
		return "TxValueKeyHumanReadable"
	case TxValueKeyAccountKey:
		return "TxValueKeyAccountKey"
	case TxValueKeyFeePayer:
		return "TxValueKeyFeePayer"
	case TxValueKeyFeeRatioOfFeePayer:
		return "TxValueKeyFeeRatioOfFeePayer"
	case TxValueKeyCodeFormat:
		return "TxValueKeyCodeFormat"
	case TxValueKeyChainID:
		return "TxValueKeyChainID"
	case TxValueKeyAccessList:
		return "TxValueKeyAccessList"
	case TxValueKeyGasTipCap:
		return "TxValueKeyGasTipCap"
	case TxValueKeyGasFeeCap:
		return "TxValueKeyGasFeeCap"
	case TxValueKeyAuthorizationList:
		return "TxValueKeyAuthorizationList"
	}

	return "UndefinedTxValueKeyType"
}

func (t TxType) String() string {
	switch t {
	case TxTypeLegacyTransaction:
		return "TxTypeLegacyTransaction"
	case TxTypeValueTransfer:
		return "TxTypeValueTransfer"
	case TxTypeFeeDelegatedValueTransfer:
		return "TxTypeFeeDelegatedValueTransfer"
	case TxTypeFeeDelegatedValueTransferWithRatio:
		return "TxTypeFeeDelegatedValueTransferWithRatio"
	case TxTypeValueTransferMemo:
		return "TxTypeValueTransferMemo"
	case TxTypeFeeDelegatedValueTransferMemo:
		return "TxTypeFeeDelegatedValueTransferMemo"
	case TxTypeFeeDelegatedValueTransferMemoWithRatio:
		return "TxTypeFeeDelegatedValueTransferMemoWithRatio"
	case TxTypeAccountCreation:
		return "TxTypeAccountCreation"
	case TxTypeAccountUpdate:
		return "TxTypeAccountUpdate"
	case TxTypeFeeDelegatedAccountUpdate:
		return "TxTypeFeeDelegatedAccountUpdate"
	case TxTypeFeeDelegatedAccountUpdateWithRatio:
		return "TxTypeFeeDelegatedAccountUpdateWithRatio"
	case TxTypeSmartContractDeploy:
		return "TxTypeSmartContractDeploy"
	case TxTypeFeeDelegatedSmartContractDeploy:
		return "TxTypeFeeDelegatedSmartContractDeploy"
	case TxTypeFeeDelegatedSmartContractDeployWithRatio:
		return "TxTypeFeeDelegatedSmartContractDeployWithRatio"
	case TxTypeSmartContractExecution:
		return "TxTypeSmartContractExecution"
	case TxTypeFeeDelegatedSmartContractExecution:
		return "TxTypeFeeDelegatedSmartContractExecution"
	case TxTypeFeeDelegatedSmartContractExecutionWithRatio:
		return "TxTypeFeeDelegatedSmartContractExecutionWithRatio"
	case TxTypeCancel:
		return "TxTypeCancel"
	case TxTypeFeeDelegatedCancel:
		return "TxTypeFeeDelegatedCancel"
	case TxTypeFeeDelegatedCancelWithRatio:
		return "TxTypeFeeDelegatedCancelWithRatio"
	case TxTypeBatch:
		return "TxTypeBatch"
	case TxTypeChainDataAnchoring:
		return "TxTypeChainDataAnchoring"
	case TxTypeFeeDelegatedChainDataAnchoring:
		return "TxTypeFeeDelegatedChainDataAnchoring"
	case TxTypeFeeDelegatedChainDataAnchoringWithRatio:
		return "TxTypeFeeDelegatedChainDataAnchoringWithRatio"
	case TxTypeEthereumAccessList:
		return "TxTypeEthereumAccessList"
	case TxTypeEthereumDynamicFee:
		return "TxTypeEthereumDynamicFee"
	case TxTypeEthereumSetCode:
		return "TxTypeEthereumSetCode"
	}

	return "UndefinedTxType"
}

func (t TxType) IsAccountCreation() bool {
	return t == TxTypeAccountCreation
}

func (t TxType) IsAccountUpdate() bool {
	return (t &^ ((1 << SubTxTypeBits) - 1)) == TxTypeAccountUpdate
}

func (t TxType) IsContractDeploy() bool {
	return (t &^ ((1 << SubTxTypeBits) - 1)) == TxTypeSmartContractDeploy
}

func (t TxType) IsCancelTransaction() bool {
	return (t &^ ((1 << SubTxTypeBits) - 1)) == TxTypeCancel
}

func (t TxType) IsLegacyTransaction() bool {
	return t == TxTypeLegacyTransaction
}

func (t TxType) IsFeeDelegatedTransaction() bool {
	return (TxTypeMask(t)&(TxFeeDelegationBitMask|TxFeeDelegationWithRatioBitMask)) != 0x0 && !t.IsEthereumTransaction()
}

func (t TxType) IsFeeDelegatedWithRatioTransaction() bool {
	return (TxTypeMask(t)&TxFeeDelegationWithRatioBitMask) != 0x0 && !t.IsEthereumTransaction()
}

func (t TxType) IsEthTypedTransaction() bool {
	return (t & 0xff00) == (EthereumTxTypeEnvelope << 8)
}

func (t TxType) IsEthereumTransaction() bool {
	return t.IsLegacyTransaction() || t.IsEthTypedTransaction()
}

func (t TxType) IsChainDataAnchoring() bool {
	return (t &^ ((1 << SubTxTypeBits) - 1)) == TxTypeChainDataAnchoring
}

type FeeRatio uint8

// FeeRatio is valid where it is [1,99].
func (f FeeRatio) IsValid() bool {
	return 1 <= f && f <= 99
}

// TxInternalData is an interface for an internal data structure of a Transaction
type TxInternalData interface {
	Type() TxType

	GetAccountNonce() uint64
	GetPrice() *big.Int
	GetGasLimit() uint64
	GetRecipient() *common.Address
	GetAmount() *big.Int
	GetHash() *common.Hash

	SetHash(*common.Hash)
	SetSignature(TxSignatures)

	// RawSignatureValues returns signatures as a slice of `*big.Int`.
	// Due to multi signatures, it is not good to return three values of `*big.Int`.
	// The format would be something like [["V":v, "R":r, "S":s}, {"V":v, "R":r, "S":s}].
	RawSignatureValues() TxSignatures

	// ValidateSignature returns true if the signature is valid.
	ValidateSignature() bool

	// RecoverAddress returns address derived from txhash and signatures(r, s, v).
	// Since EIP155Signer modifies V value during recovering while other signers don't, it requires vfunc for the treatment.
	RecoverAddress(txhash common.Hash, homestead bool, vfunc func(*big.Int) *big.Int) (common.Address, error)

	// RecoverPubkey returns a public key derived from txhash and signatures(r, s, v).
	// Since EIP155Signer modifies V value during recovering while other signers don't, it requires vfunc for the treatment.
	RecoverPubkey(txhash common.Hash, homestead bool, vfunc func(*big.Int) *big.Int) ([]*ecdsa.PublicKey, error)

	// ChainId returns which chain id this transaction was signed for (if at all)
	ChainId() *big.Int

	// Equal returns true if all attributes are the same.
	Equal(t TxInternalData) bool

	// IntrinsicGas computes additional 'intrinsic gas' based on tx types.
	IntrinsicGas(currentBlockNumber uint64) (uint64, error)

	// SerializeForSign returns a slice containing attributes to make its tx signature.
	SerializeForSign() []interface{}

	// SenderTxHash returns a hash of the tx without the fee payer's address and signature.
	SenderTxHash() common.Hash

	// Validate returns nil if tx is validated with the given stateDB and currentBlockNumber.
	// Otherwise, it returns an error.
	// This function is called in TxPool.validateTx() and TxInternalData.Execute().
	Validate(stateDB StateDB, currentBlockNumber uint64) error

	// ValidateMutableValue returns nil if tx is validated. Otherwise, it returns an error.
	// The function validates tx values associated with mutable values in the state.
	// MutableValues: accountKey, the existence of creating address, feePayer's balance, etc.
	ValidateMutableValue(stateDB StateDB, currentBlockNumber uint64) error

	// IsLegacyTransaction returns true if the tx type is a legacy transaction (TxInternalDataLegacy) object.
	IsLegacyTransaction() bool

	// GetRoleTypeForValidation returns RoleType to validate this transaction.
	GetRoleTypeForValidation() accountkey.RoleType

	// String returns a string containing information about the fields of the object.
	String() string

	// Execute performs execution of the transaction according to the transaction type.
	Execute(sender ContractRef, vm VM, stateDB StateDB, currentBlockNumber uint64, gas uint64, value *big.Int) (ret []byte, usedGas uint64, err error)

	MakeRPCOutput() map[string]interface{}
}

type TxInternalDataContractAddressFiller interface {
	// FillContractAddress fills contract address to receipt. This only works for types deploying a smart contract.
	FillContractAddress(from common.Address, r *Receipt)
}

type TxInternalDataSerializeForSignToByte interface {
	SerializeForSignToBytes() []byte
}

// TxInternalDataFeePayer has functions related to fee delegated transactions.
type TxInternalDataFeePayer interface {
	GetFeePayer() common.Address

	// GetFeePayerRawSignatureValues returns fee payer's signatures as a slice of `*big.Int`.
	// Due to multi signatures, it is not good to return three values of `*big.Int`.
	// The format would be something like [["V":v, "R":r, "S":s}, {"V":v, "R":r, "S":s}].
	GetFeePayerRawSignatureValues() TxSignatures

	// RecoverFeePayerPubkey returns the fee payer's public key derived from txhash and signatures(r, s, v).
	RecoverFeePayerPubkey(txhash common.Hash, homestead bool, vfunc func(*big.Int) *big.Int) ([]*ecdsa.PublicKey, error)

	SetFeePayerSignatures(s TxSignatures)
}

// TxInternalDataFeeRatio has a function `GetFeeRatio`.
type TxInternalDataFeeRatio interface {
	// GetFeeRatio returns a ratio of tx fee paid by the fee payer in percentage.
	// For example, if it is 30, 30% of tx fee will be paid by the fee payer.
	// 70% will be paid by the sender.
	GetFeeRatio() FeeRatio
}

// TxInternalDataFrom has a function `GetFrom()`.
// All other transactions to be implemented will have `from` field, but
// `TxInternalDataLegacy` (a legacy transaction type) does not have the field.
// Hence, this function is defined in another interface TxInternalDataFrom.
type TxInternalDataFrom interface {
	GetFrom() common.Address
}

// TxInternalDataPayload has a function `GetPayload()`.
// Since the payload field is not a common field for all tx types, we provide
// an interface `TxInternalDataPayload` to obtain the payload.
type TxInternalDataPayload interface {
	GetPayload() []byte
}

// TxInternalDataEthTyped has a function related to EIP-2718 Ethereum typed transaction.
// For supporting new typed transaction defined EIP-2718, We provide an interface `TxInternalDataEthTyped `
type TxInternalDataEthTyped interface {
	setSignatureValues(chainID, v, r, s *big.Int)
	GetAccessList() AccessList
	TxHash() common.Hash
}

// TxInternalDataBaseFee has a function related to EIP-1559 Ethereum typed transaction.
type TxInternalDataBaseFee interface {
	GetGasTipCap() *big.Int
	GetGasFeeCap() *big.Int
}

// Since we cannot access the package `blockchain/vm` directly, an interface `VM` is introduced.
// TODO-Kaia-Refactoring: Transaction and related data structures should be a new package.
type VM interface {
	Create(caller ContractRef, code []byte, gas uint64, value *big.Int, codeFormat params.CodeFormat) (ret []byte, contractAddr common.Address, leftOverGas uint64, err error)
	CreateWithAddress(caller ContractRef, code []byte, gas uint64, value *big.Int, contractAddr common.Address, humanReadable bool, codeFormat params.CodeFormat) ([]byte, common.Address, uint64, error)
	Call(caller ContractRef, addr common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, leftOverGas uint64, err error)
}

// Since we cannot access the package `blockchain/state` directly, an interface `StateDB` is introduced.
// TODO-Kaia-Refactoring: Transaction and related data structures should be a new package.
type StateDB interface {
	IncNonce(common.Address)
	Exist(common.Address) bool
	UpdateKey(addr common.Address, key accountkey.AccountKey, currentBlockNumber uint64) error
	CreateEOA(addr common.Address, humanReadable bool, key accountkey.AccountKey)
	CreateSmartContractAccount(addr common.Address, format params.CodeFormat, r params.Rules)
	CreateSmartContractAccountWithKey(addr common.Address, humanReadable bool, key accountkey.AccountKey, format params.CodeFormat, r params.Rules)
	IsProgramAccount(addr common.Address) bool
	IsContractAvailable(addr common.Address) bool
	IsValidCodeFormat(addr common.Address) bool
	GetKey(addr common.Address) accountkey.AccountKey
	GetAccount(addr common.Address) account.Account
}

func NewTxInternalData(t TxType) (TxInternalData, error) {
	switch t {
	case TxTypeLegacyTransaction:
		return newTxInternalDataLegacy(), nil
	case TxTypeValueTransfer:
		return newTxInternalDataValueTransfer(), nil
	case TxTypeFeeDelegatedValueTransfer:
		return newTxInternalDataFeeDelegatedValueTransfer(), nil
	case TxTypeFeeDelegatedValueTransferWithRatio:
		return NewTxInternalDataFeeDelegatedValueTransferWithRatio(), nil
	case TxTypeValueTransferMemo:
		return newTxInternalDataValueTransferMemo(), nil
	case TxTypeFeeDelegatedValueTransferMemo:
		return newTxInternalDataFeeDelegatedValueTransferMemo(), nil
	case TxTypeFeeDelegatedValueTransferMemoWithRatio:
		return newTxInternalDataFeeDelegatedValueTransferMemoWithRatio(), nil
	// case TxTypeAccountCreation:
	//	return newTxInternalDataAccountCreation(), nil
	case TxTypeAccountUpdate:
		return newTxInternalDataAccountUpdate(), nil
	case TxTypeFeeDelegatedAccountUpdate:
		return newTxInternalDataFeeDelegatedAccountUpdate(), nil
	case TxTypeFeeDelegatedAccountUpdateWithRatio:
		return newTxInternalDataFeeDelegatedAccountUpdateWithRatio(), nil
	case TxTypeSmartContractDeploy:
		return newTxInternalDataSmartContractDeploy(), nil
	case TxTypeFeeDelegatedSmartContractDeploy:
		return newTxInternalDataFeeDelegatedSmartContractDeploy(), nil
	case TxTypeFeeDelegatedSmartContractDeployWithRatio:
		return newTxInternalDataFeeDelegatedSmartContractDeployWithRatio(), nil
	case TxTypeSmartContractExecution:
		return newTxInternalDataSmartContractExecution(), nil
	case TxTypeFeeDelegatedSmartContractExecution:
		return newTxInternalDataFeeDelegatedSmartContractExecution(), nil
	case TxTypeFeeDelegatedSmartContractExecutionWithRatio:
		return newTxInternalDataFeeDelegatedSmartContractExecutionWithRatio(), nil
	case TxTypeCancel:
		return newTxInternalDataCancel(), nil
	case TxTypeFeeDelegatedCancel:
		return newTxInternalDataFeeDelegatedCancel(), nil
	case TxTypeFeeDelegatedCancelWithRatio:
		return newTxInternalDataFeeDelegatedCancelWithRatio(), nil
	case TxTypeChainDataAnchoring:
		return newTxInternalDataChainDataAnchoring(), nil
	case TxTypeFeeDelegatedChainDataAnchoring:
		return newTxInternalDataFeeDelegatedChainDataAnchoring(), nil
	case TxTypeFeeDelegatedChainDataAnchoringWithRatio:
		return newTxInternalDataFeeDelegatedChainDataAnchoringWithRatio(), nil
	case TxTypeEthereumAccessList:
		return newTxInternalDataEthereumAccessList(), nil
	case TxTypeEthereumDynamicFee:
		return newTxInternalDataEthereumDynamicFee(), nil
	case TxTypeEthereumSetCode:
		return newTxInternalDataEthereumSetCode(), nil
	}

	return nil, errUndefinedTxType
}

func NewTxInternalDataWithMap(t TxType, values map[TxValueKeyType]interface{}) (TxInternalData, error) {
	switch t {
	case TxTypeLegacyTransaction:
		return newTxInternalDataLegacyWithMap(values)
	case TxTypeValueTransfer:
		return newTxInternalDataValueTransferWithMap(values)
	case TxTypeFeeDelegatedValueTransfer:
		return newTxInternalDataFeeDelegatedValueTransferWithMap(values)
	case TxTypeFeeDelegatedValueTransferWithRatio:
		return newTxInternalDataFeeDelegatedValueTransferWithRatioWithMap(values)
	case TxTypeValueTransferMemo:
		return newTxInternalDataValueTransferMemoWithMap(values)
	case TxTypeFeeDelegatedValueTransferMemo:
		return newTxInternalDataFeeDelegatedValueTransferMemoWithMap(values)
	case TxTypeFeeDelegatedValueTransferMemoWithRatio:
		return newTxInternalDataFeeDelegatedValueTransferMemoWithRatioWithMap(values)
	// case TxTypeAccountCreation:
	//	return newTxInternalDataAccountCreationWithMap(values)
	case TxTypeAccountUpdate:
		return newTxInternalDataAccountUpdateWithMap(values)
	case TxTypeFeeDelegatedAccountUpdate:
		return newTxInternalDataFeeDelegatedAccountUpdateWithMap(values)
	case TxTypeFeeDelegatedAccountUpdateWithRatio:
		return newTxInternalDataFeeDelegatedAccountUpdateWithRatioWithMap(values)
	case TxTypeSmartContractDeploy:
		return newTxInternalDataSmartContractDeployWithMap(values)
	case TxTypeFeeDelegatedSmartContractDeploy:
		return newTxInternalDataFeeDelegatedSmartContractDeployWithMap(values)
	case TxTypeFeeDelegatedSmartContractDeployWithRatio:
		return newTxInternalDataFeeDelegatedSmartContractDeployWithRatioWithMap(values)
	case TxTypeSmartContractExecution:
		return newTxInternalDataSmartContractExecutionWithMap(values)
	case TxTypeFeeDelegatedSmartContractExecution:
		return newTxInternalDataFeeDelegatedSmartContractExecutionWithMap(values)
	case TxTypeFeeDelegatedSmartContractExecutionWithRatio:
		return newTxInternalDataFeeDelegatedSmartContractExecutionWithRatioWithMap(values)
	case TxTypeCancel:
		return newTxInternalDataCancelWithMap(values)
	case TxTypeFeeDelegatedCancel:
		return newTxInternalDataFeeDelegatedCancelWithMap(values)
	case TxTypeFeeDelegatedCancelWithRatio:
		return newTxInternalDataFeeDelegatedCancelWithRatioWithMap(values)
	case TxTypeChainDataAnchoring:
		return newTxInternalDataChainDataAnchoringWithMap(values)
	case TxTypeFeeDelegatedChainDataAnchoring:
		return newTxInternalDataFeeDelegatedChainDataAnchoringWithMap(values)
	case TxTypeFeeDelegatedChainDataAnchoringWithRatio:
		return newTxInternalDataFeeDelegatedChainDataAnchoringWithRatioWithMap(values)
	case TxTypeEthereumAccessList:
		return newTxInternalDataEthereumAccessListWithMap(values)
	case TxTypeEthereumDynamicFee:
		return newTxInternalDataEthereumDynamicFeeWithMap(values)
	case TxTypeEthereumSetCode:
		return newTxInternalDataEthereumSetCodeWithMap(values)
	}

	return nil, errUndefinedTxType
}

// toWordSize returns the ceiled word size required for init code payment calculation.
func toWordSize(size uint64) uint64 {
	if size > math.MaxUint64-31 {
		return math.MaxUint64/32 + 1
	}

	return (size + 31) / 32
}

// Klaytn-TxTypes since genesis, and EthTxTypes since istanbul use this.
func IntrinsicGasPayload(gas uint64, data []byte, isContractCreation bool, rules params.Rules) (uint64, error) {
	// Bump the required gas by the amount of transactional data
	length := uint64(len(data))
	if length > 0 {
		// Zero and non-zero bytes are priced differently
		z := uint64(bytes.Count(data, []byte{0}))
		nz := length - z

		// Since the genesis block, a flat 100 gas is paid
		// regardless of whether the value is zero or non-zero.
		nonZeroGas, zeroGas := params.TxDataGas, params.TxDataGas
		if rules.IsPrague {
			nonZeroGas = params.TxDataNonZeroGasEIP2028
			zeroGas = params.TxDataZeroGas
		}
		// Make sure we don't exceed uint64 for all data combinations
		if (math.MaxUint64-gas)/nonZeroGas < nz {
			return 0, ErrGasUintOverflow
		}
		gas += nz * nonZeroGas

		if (math.MaxUint64-gas)/zeroGas < z {
			return 0, ErrGasUintOverflow
		}
		gas += z * zeroGas
	}

	if isContractCreation && rules.IsShanghai {
		lenWords := toWordSize(length)
		if (math.MaxUint64-gas)/params.InitCodeWordGas < lenWords {
			return 0, ErrGasUintOverflow
		}
		gas += lenWords * params.InitCodeWordGas
	}
	return gas, nil
}

// Eth-TxTypes before istanbul use this. Only 0 tx type exists before Istanbul (No dynamic and access list types correspond to)
// Calculate gas cost for type 0 transactions:
// 68 gas for each non-zero byte and 16 gas for each zero byte in the data field.
func IntrinsicGasPayloadLegacy(gas uint64, data []byte) (uint64, error) {
	length := uint64(len(data))
	if length > 0 {
		// Zero and non-zero bytes are priced differently
		z := uint64(bytes.Count(data, []byte{0}))
		nz := length - z

		// Make sure we don't exceed uint64 for all data combinations
		if (math.MaxUint64-gas)/params.TxDataNonZeroGasFrontier < nz {
			return 0, ErrGasUintOverflow
		}
		gas += nz * params.TxDataNonZeroGasFrontier

		if (math.MaxUint64-gas)/params.TxDataZeroGas < z {
			return 0, ErrGasUintOverflow
		}
		gas += z * params.TxDataZeroGas
	}

	return gas, nil
}

// IntrinsicGas computes the 'intrinsic gas' for a message with the given data.
func IntrinsicGas(data []byte, accessList AccessList, authorizationList []SetCodeAuthorization, contractCreation bool, r params.Rules) (uint64, error) {
	// Set the starting gas for the raw transaction
	var gas uint64

	if contractCreation {
		gas = params.TxGasContractCreation
	} else {
		gas = params.TxGas
	}

	var gasPayloadWithGas uint64
	var err error
	if r.IsIstanbul {
		// tx types 1,2 only exist after istanbul; so they take this path.
		// tx types 8+ take this path as well.
		gasPayloadWithGas, err = IntrinsicGasPayload(gas, data, contractCreation, r)
	} else {
		// only for tx type 0 before istanbul.
		gasPayloadWithGas, err = IntrinsicGasPayloadLegacy(gas, data)
	}

	if err != nil {
		return 0, err
	}

	// We charge additional gas for the accessList:
	// ACCESS_LIST_ADDRESS_COST : gas per address in AccessList
	// ACCESS_LIST_STORAGE_KEY_COST : gas per storage key in AccessList
	if accessList != nil {
		gasPayloadWithGas += uint64(len(accessList)) * params.TxAccessListAddressGas
		gasPayloadWithGas += uint64(accessList.StorageKeys()) * params.TxAccessListStorageKeyGas
	}

	// We charge additional gas for the authorizationList:
	// PER_EMPTY_ACCOUNT_COST : gas per address in authorizationList
	// Since this is the same value as CallNewAccountGas, we will use this.
	if authorizationList != nil {
		gasPayloadWithGas += uint64(len(authorizationList)) * params.CallNewAccountGas
	}

	return gasPayloadWithGas, nil
}

var txTypeToGasMap = map[TxType]uint64{
	TxTypeLegacyTransaction:                           params.TxGas,
	TxTypeValueTransfer:                               params.TxGasValueTransfer,
	TxTypeFeeDelegatedValueTransfer:                   params.TxGasValueTransfer + params.TxGasFeeDelegated,
	TxTypeFeeDelegatedValueTransferWithRatio:          params.TxGasValueTransfer + params.TxGasFeeDelegatedWithRatio,
	TxTypeValueTransferMemo:                           params.TxGasValueTransfer,
	TxTypeFeeDelegatedValueTransferMemo:               params.TxGasValueTransfer + params.TxGasFeeDelegated,
	TxTypeFeeDelegatedValueTransferMemoWithRatio:      params.TxGasValueTransfer + params.TxGasFeeDelegatedWithRatio,
	TxTypeAccountCreation:                             params.TxGasAccountCreation,
	TxTypeAccountUpdate:                               params.TxGasAccountUpdate,
	TxTypeFeeDelegatedAccountUpdate:                   params.TxGasAccountUpdate + params.TxGasFeeDelegated,
	TxTypeFeeDelegatedAccountUpdateWithRatio:          params.TxGasAccountUpdate + params.TxGasFeeDelegatedWithRatio,
	TxTypeSmartContractDeploy:                         params.TxGasContractCreation,
	TxTypeFeeDelegatedSmartContractDeploy:             params.TxGasContractCreation + params.TxGasFeeDelegated,
	TxTypeFeeDelegatedSmartContractDeployWithRatio:    params.TxGasContractCreation + params.TxGasFeeDelegatedWithRatio,
	TxTypeSmartContractExecution:                      params.TxGasContractExecution,
	TxTypeFeeDelegatedSmartContractExecution:          params.TxGasContractExecution + params.TxGasFeeDelegated,
	TxTypeFeeDelegatedSmartContractExecutionWithRatio: params.TxGasContractExecution + params.TxGasFeeDelegatedWithRatio,
	TxTypeCancel:                                  params.TxGasCancel,
	TxTypeFeeDelegatedCancel:                      params.TxGasCancel + params.TxGasFeeDelegated,
	TxTypeFeeDelegatedCancelWithRatio:             params.TxGasCancel + params.TxGasFeeDelegatedWithRatio,
	TxTypeChainDataAnchoring:                      params.TxChainDataAnchoringGas,
	TxTypeFeeDelegatedChainDataAnchoring:          params.TxChainDataAnchoringGas + params.TxGasFeeDelegated,
	TxTypeFeeDelegatedChainDataAnchoringWithRatio: params.TxChainDataAnchoringGas + params.TxGasFeeDelegatedWithRatio,
	TxTypeEthereumAccessList:                      params.TxGas,
	TxTypeEthereumDynamicFee:                      params.TxGas,
	TxTypeEthereumSetCode:                         params.TxGas,
}

func GetTxGasForTxType(txType TxType) (uint64, error) {
	if gas, exists := txTypeToGasMap[txType]; exists {
		return gas, nil
	}
	return 0, fmt.Errorf("cannot find txGas for txType %s", txType.String())
}

func GetTxGasForTxTypeWithAccountKey(txType TxType, accountKey accountkey.AccountKey, currentBlockNumber uint64, humanReadable bool) (uint64, error) {
	gas, err := GetTxGasForTxType(txType)
	if err != nil {
		return 0, err
	}
	var gasKey uint64
	if accountKey != nil {
		gasKey, err = accountKey.AccountCreationGas(currentBlockNumber)
		if err != nil {
			return 0, err
		}
	}
	gas += gasKey
	if humanReadable {
		gas += params.TxGasHumanReadable
	}
	return gas, nil
}

// CalcFeeWithRatio returns feePayer's fee and sender's fee based on feeRatio.
// For example, if fee = 100 and feeRatio = 30, feePayer = 30 and feeSender = 70.
func CalcFeeWithRatio(feeRatio FeeRatio, fee *big.Int) (*big.Int, *big.Int) {
	// feePayer = fee * ratio / 100
	feePayer := new(big.Int).Div(new(big.Int).Mul(fee, new(big.Int).SetUint64(uint64(feeRatio))), common.Big100)
	// feeSender = fee - feePayer
	feeSender := new(big.Int).Sub(fee, feePayer)

	return feePayer, feeSender
}

func equalRecipient(a, b *common.Address) bool {
	if a == nil && b == nil {
		return true
	}

	if a != nil && b != nil && bytes.Equal(a.Bytes(), b.Bytes()) {
		return true
	}

	return false
}

// NewAccountCreationTransactionWithMap is a test only function since the accountCreation tx is disabled.
// The function generates an accountCreation function like 'NewTxInternalDataWithMap()'.
func NewAccountCreationTransactionWithMap(values map[TxValueKeyType]interface{}) (*Transaction, error) {
	txData, err := newTxInternalDataAccountCreationWithMap(values)
	if err != nil {
		return nil, err
	}

	return NewTx(txData), nil
}

func calculateTxSize(data TxInternalData) common.StorageSize {
	c := writeCounter(0)
	rlp.Encode(&c, data)
	return common.StorageSize(c)
}

func validate7702(stateDB StateDB, txType TxType, from, to common.Address) error {
	switch txType {
	// Group 1: Recipient must be EOA without code
	case TxTypeValueTransfer,
		TxTypeFeeDelegatedValueTransfer,
		TxTypeFeeDelegatedValueTransferWithRatio,
		TxTypeValueTransferMemo,
		TxTypeFeeDelegatedValueTransferMemo,
		TxTypeFeeDelegatedValueTransferMemoWithRatio:
		acc := stateDB.GetAccount(to)
		if acc == nil {
			return nil
		}
		if acc.Type() == account.SmartContractAccountType {
			return kerrors.ErrToMustBeEOAWithoutCode
		}
		eoa, ok := acc.(*account.ExternallyOwnedAccount)
		if !ok || !bytes.Equal(eoa.GetCodeHash(), emptyCodeHash) {
			return kerrors.ErrToMustBeEOAWithoutCode
		}

		return nil

	// Group 2: From must be EOA without code
	case TxTypeAccountUpdate,
		TxTypeFeeDelegatedAccountUpdate,
		TxTypeFeeDelegatedAccountUpdateWithRatio:
		acc := stateDB.GetAccount(from)
		if acc == nil {
			return nil
		}
		if acc.Type() == account.SmartContractAccountType {
			return kerrors.ErrFromMustBeEOAWithoutCode
		}
		eoa, ok := acc.(*account.ExternallyOwnedAccount)
		if !ok || !bytes.Equal(eoa.GetCodeHash(), emptyCodeHash) {
			return kerrors.ErrFromMustBeEOAWithoutCode
		}

		return nil

	// Group 3: Recipient must be EOA with code or SCA
	case TxTypeSmartContractExecution,
		TxTypeFeeDelegatedSmartContractExecution,
		TxTypeFeeDelegatedSmartContractExecutionWithRatio:
		acc := stateDB.GetAccount(to)
		if acc == nil {
			return kerrors.ErrToMustBeEOAWithCodeOrSCA
		}
		if acc.Type() == account.SmartContractAccountType {
			return nil
		}
		eoa, ok := acc.(*account.ExternallyOwnedAccount)
		if !ok || !bytes.Equal(eoa.GetCodeHash(), emptyCodeHash) {
			return nil
		}

		return kerrors.ErrToMustBeEOAWithCodeOrSCA

	default:
		return nil
	}
}
