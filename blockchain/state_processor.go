// Modifications Copyright 2024 The Kaia Authors
// Modifications Copyright 2018 The klaytn Authors
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
// This file is derived from core/state_processor.go (2018/06/04).
// Modified and improved for the klaytn development.
// Modified and improved for the Kaia development.

package blockchain

import (
	"time"

	"github.com/kaiachain/kaia/blockchain/state"
	"github.com/kaiachain/kaia/blockchain/types"
	"github.com/kaiachain/kaia/blockchain/vm"
	"github.com/kaiachain/kaia/common"
	"github.com/kaiachain/kaia/consensus"
	"github.com/kaiachain/kaia/params"
)

// StateProcessor is a basic Processor, which takes care of transitioning
// state from one point to another.
//
// StateProcessor implements Processor.
type StateProcessor struct {
	config *params.ChainConfig // Chain configuration options
	bc     *BlockChain         // Canonical block chain
	engine consensus.Engine    // Consensus engine used for block rewards
}

// ProcessStats includes the time statistics regarding StateProcessor.Process.
type ProcessStats struct {
	BeforeApplyTxs time.Time
	AfterApplyTxs  time.Time
	AfterFinalize  time.Time
}

// NewStateProcessor initialises a new StateProcessor.
func NewStateProcessor(config *params.ChainConfig, bc *BlockChain, engine consensus.Engine) *StateProcessor {
	return &StateProcessor{
		config: config,
		bc:     bc,
		engine: engine,
	}
}

// Process processes the state changes according to the Kaia rules by running
// the transaction messages using the statedb and applying any rewards to the processor.
//
// Process returns the receipts and logs accumulated during the process and
// returns the amount of gas that was used in the process. If any of the
// transactions failed to execute due to insufficient gas it will return an error.
func (p *StateProcessor) Process(block *types.Block, statedb *state.StateDB, cfg vm.Config) (types.Receipts, []*types.Log, uint64, []*vm.InternalTxTrace, ProcessStats, error) {
	var (
		receipts         types.Receipts
		usedGas          = new(uint64)
		header           = block.Header()
		allLogs          []*types.Log
		internalTxTraces []*vm.InternalTxTrace
		processStats     ProcessStats
	)

	p.engine.Initialize(p.bc, header, statedb)

	// Extract author from the header
	author, _ := p.bc.Engine().Author(header) // Ignore error, we're past header validation

	processStats.BeforeApplyTxs = time.Now()
	// Iterate over and process the individual transactions
	for i, tx := range block.Transactions() {
		statedb.SetTxContext(tx.Hash(), block.Hash(), i)
		receipt, internalTxTrace, err := p.bc.ApplyTransaction(p.config, &author, statedb, header, tx, usedGas, &cfg)
		if err != nil {
			return nil, nil, 0, nil, processStats, err
		}
		receipts = append(receipts, receipt)
		allLogs = append(allLogs, receipt.Logs...)
		internalTxTraces = append(internalTxTraces, internalTxTrace)
	}
	processStats.AfterApplyTxs = time.Now()

	// Finalize the block, applying any consensus engine specific extras (e.g. block rewards)
	if _, err := p.engine.Finalize(p.bc, header, statedb, block.Transactions(), receipts); err != nil {
		return nil, nil, 0, nil, processStats, err
	}
	processStats.AfterFinalize = time.Now()

	return receipts, allLogs, *usedGas, internalTxTraces, processStats, nil
}

// ProcessParentBlockHash stores the parent block hash in the history storage contract
// as per EIP-2935.
func ProcessParentBlockHash(header *types.Header, vmenv *vm.EVM, statedb vm.StateDB, rules params.Rules) error {
	var (
		from     = params.SystemAddress
		data     = header.ParentHash.Bytes()
		gasLimit = uint64(30_000_000)
	)

	intrinsicGas, err := types.IntrinsicGas(data, nil, nil, false, rules)
	if err != nil {
		return err
	}

	msg := types.NewMessage(
		from,
		&params.HistoryStorageAddress,
		0,
		common.Big0,
		gasLimit,
		common.Big0,
		nil,
		nil,
		data,
		false,
		intrinsicGas,
		nil,
		nil,
		nil,
	)

	vmenv.Reset(NewEVMTxContext(msg, header, vmenv.ChainConfig()), statedb)
	statedb.AddAddressToAccessList(params.HistoryStorageAddress)
	vmenv.Call(vm.AccountRef(from), *msg.To(), msg.Data(), gasLimit, common.Big0)
	statedb.Finalise(true, true)
	return nil
}
