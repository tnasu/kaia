// Modifications Copyright 2024 The Kaia Authors
// Modifications Copyright 2019 The klaytn Authors
// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
//
// This file is derived from params/protocol_params.go (2018/06/04).
// Modified and improved for the klaytn development.
// Modified and improved for the Kaia development.

package params

import "github.com/kaiachain/kaia/common/math"

const (
	// Computation cost for opcodes
	ExtCodeHashComputationCost    = 1000
	Create2ComputationCost        = 10000
	StaticCallComputationCost     = 10000
	ReturnDataSizeComputationCost = 10
	ReturnDataCopyComputationCost = 40
	RevertComputationCost         = 0
	DelegateCallComputationCost   = 696
	StopComputationCost           = 0
	AddComputationCost            = 150
	MulComputationCost            = 200
	SubComputationCost            = 219
	DivComputationCost            = 404
	SdivComputationCost           = 739
	ModComputationCost            = 812
	SmodComputationCost           = 560
	AddmodComputationCost         = 3349
	MulmodComputationCost         = 4757
	ExpComputationCost            = 5000
	ShlComputationCost            = 1603
	ShrComputationCost            = 1346
	SarComputationCost            = 1815
	SignExtendComputationCost     = 481
	LtComputationCost             = 201
	GtComputationCost             = 264
	SltComputationCost            = 176
	SgtComputationCost            = 222
	EqComputationCost             = 220
	IszeroComputationCost         = 165
	AndComputationCost            = 288
	XorComputationCost            = 657
	OrComputationCost             = 160
	NotComputationCost            = 1289
	ByteComputationCost           = 589
	Sha3ComputationCost           = 2465
	AddressComputationCost        = 284
	BalanceComputationCost        = 1407
	OriginComputationCost         = 210
	CallerComputationCost         = 188
	CallValueComputationCost      = 149
	CallDataLoadComputationCost   = 596
	CallDataSizeComputationCost   = 194
	CallDataCopyComputationCost   = 100
	CodeSizeComputationCost       = 145
	CodeCopyComputationCost       = 898
	GasPriceComputationCost       = 131
	ExtCodeSizeComputationCost    = 1481
	ExtCodeCopyComputationCost    = 1000
	BlockHashComputationCost      = 500
	CoinbaseComputationCost       = 189
	TimestampComputationCost      = 265
	NumberComputationCost         = 202
	DifficultyComputationCost     = 180
	GasLimitComputationCost       = 166
	PopComputationCost            = 140
	MloadComputationCost          = 376
	MstoreComputationCost         = 288
	Mstore8ComputationCost        = 5142
	SloadComputationCost          = 835
	SstoreComputationCost         = 1548
	JumpComputationCost           = 253
	JumpiComputationCost          = 176
	PcComputationCost             = 147
	MsizeComputationCost          = 137
	GasComputationCost            = 230
	JumpDestComputationCost       = 10
	PushComputationCost           = 120
	Dup1ComputationCost           = 190
	Dup2ComputationCost           = 190
	Dup3ComputationCost           = 176
	Dup4ComputationCost           = 142
	Dup5ComputationCost           = 177
	Dup6ComputationCost           = 165
	Dup7ComputationCost           = 147
	Dup8ComputationCost           = 157
	Dup9ComputationCost           = 138
	Dup10ComputationCost          = 174
	Dup11ComputationCost          = 141
	Dup12ComputationCost          = 144
	Dup13ComputationCost          = 157
	Dup14ComputationCost          = 143
	Dup15ComputationCost          = 237
	Dup16ComputationCost          = 149
	Swap1ComputationCost          = 141
	Swap2ComputationCost          = 156
	Swap3ComputationCost          = 145
	Swap4ComputationCost          = 135
	Swap5ComputationCost          = 115
	Swap6ComputationCost          = 146
	Swap7ComputationCost          = 199
	Swap8ComputationCost          = 130
	Swap9ComputationCost          = 160
	Swap10ComputationCost         = 134
	Swap11ComputationCost         = 147
	Swap12ComputationCost         = 128
	Swap13ComputationCost         = 121
	Swap14ComputationCost         = 114
	Swap15ComputationCost         = 197
	Swap16ComputationCost         = 128
	Log0ComputationCost           = 100
	Log1ComputationCost           = 1000
	Log2ComputationCost           = 1000
	Log3ComputationCost           = 1000
	Log4ComputationCost           = 1000
	CreateComputationCost         = 2094
	CallComputationCost           = 5000
	CallCodeComputationCost       = 4000
	ReturnComputationCost         = 0
	SelfDestructComputationCost   = 0

	// Computation cost for precompiled contracts
	EcrecoverComputationCost                       = 113150
	Sha256PerWordComputationCost                   = 100
	Sha256BaseComputationCost                      = 1000
	Ripemd160PerWordComputationCost                = 10
	Ripemd160BaseComputationCost                   = 100
	IdentityPerWordComputationCost                 = 0
	IdentityBaseComputationCost                    = 0
	BigModExpPerGasComputationCost                 = 10
	BigModExpBaseComputationCost                   = 100
	Bn256AddComputationCost                        = 8000
	Bn256ScalarMulComputationCost                  = 100000
	Bn256ParingBaseComputationCost                 = 2000000
	Bn256ParingPerPointComputationCost             = 1000000
	BlobTxPointEvaluationPrecompileComputationCost = 2200000
	VMLogPerByteComputationCost                    = 0
	VMLogBaseComputationCost                       = 10
	FeePayerComputationCost                        = 10
	ValidateSenderPerSigComputationCost            = 180000
	ValidateSenderBaseComputationCost              = 10000

	// computation costs added at istanbulCompatible
	ChainIDComputationCost      = 120
	SelfBalanceComputationCost  = 374
	Blake2bBaseComputationCost  = 10000
	Blake2bScaleComputationCost = 10

	// computation cost added at londonCompatible
	BaseFeeComputationCost = 198

	// computation cost added at KoreCompatible
	RandomComputationCost = 1498

	// computation cost added at ShanghaiCompatible
	Push0ComputationCost = 80

	// computation cost added at CancunCompatible
	McopyComputationCost       = 250
	TloadComputationCost       = 220
	TstoreComputationCost      = 280
	BlobHashComptationCost     = 165
	BlobBaseFeeComputationCost = 120

	// opcode computation cost modification - istanbul
	AddmodComputationCostIstanbul = 1410
	MulmodComputationCostIstanbul = 1760
	ShlComputationCostIstanbul    = 478
	ShrComputationCostIstanbul    = 498
	SarComputationCostIstanbul    = 834
	XorComputationCostIstanbul    = 454
	NotComputationCostIstanbul    = 364

	// opcode computation cost codification - cancun
	SdivComputationCostCancun    = 360
	ModComputationCostCancun     = 320
	AddmodComputationCostCancun  = 360
	MulmodComputationCostCancun  = 700
	ExpComputationCostCancun     = 720
	Sha3ComputationCostCancun    = 560
	Mstore8ComputationCostCancun = 230

	SloadComputationCostCancun  = 2550
	SstoreComputationCostCancun = 2510
	Log1ComputationCostCancun   = 500
	Log2ComputationCostCancun   = 500
	Log3ComputationCostCancun   = 500
	Log4ComputationCostCancun   = 500

	Bls12381G1AddComputationCost          = 18750
	Bls12381G1MulComputationCost          = 600000
	Bls12381G2AddComputationCost          = 30000
	Bls12381G2MulComputationCost          = 1125000
	Bls12381PairingBaseComputationCost    = 1885000
	Bls12381PairingPerPairComputationCost = 1630000
	Bls12381MapG1ComputationCost          = 275000
	Bls12381MapG2ComputationCost          = 1190000

	OpcodeComputationCostLimit         = 100000000      // 100ms
	OpcodeComputationCostLimitCancun   = 150000000      // 150ms
	OpcodeComputationCostLimitInfinite = math.MaxUint64 // pass it to disable computation cost checks
)

// OpcodeComputationCostLimitOverride set by --opcode-computation-cost-limit.
// Overrides chain default settings above.
var OpcodeComputationCostLimitOverride = uint64(0)
