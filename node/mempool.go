package node

import (
	"dummy-chain/common/types"
	"sync"

	ecommon "github.com/ethereum/go-ethereum/common"
)

type MemoryPool struct {
	memPool map[ecommon.Hash]*types.Transaction
	lock    sync.Mutex
}

func NewMemoryPool() *MemoryPool {
	return &MemoryPool{
		memPool: make(map[ecommon.Hash]*types.Transaction),
		lock:    sync.Mutex{},
	}
}

func (mp *MemoryPool) AddTransaction(tx *types.Transaction) {
	mp.lock.Lock()
	mp.memPool[tx.Hash] = tx
	mp.lock.Unlock()
}

func (mp *MemoryPool) GetMemPool() []*types.Transaction {
	mp.lock.Lock()
	defer mp.lock.Unlock()
	var txs []*types.Transaction
	for _, tx := range mp.memPool {
		txs = append(txs, tx)
	}
	return txs
}

func (mp *MemoryPool) RemoveTxs(tx []*types.Transaction) {
	mp.lock.Lock()
	for _, tx := range tx {
		delete(mp.memPool, tx.Hash)
	}
	mp.lock.Unlock()
}
