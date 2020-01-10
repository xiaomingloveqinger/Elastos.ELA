package mempool

import (
	"math/rand"
	"testing"

	"github.com/elastos/Elastos.ELA/common"
	"github.com/elastos/Elastos.ELA/core/types"
	"github.com/elastos/Elastos.ELA/core/types/payload"
	"github.com/elastos/Elastos.ELA/elanet/pact"
	"github.com/stretchr/testify/assert"
)

func TestTxFeeOrderedList_AddTx(t *testing.T) {
	firedPopBack := false
	onPopBack := func(common.Uint256) {
		firedPopBack = true
	}

	protoTx := types.Transaction{
		TxType:  types.TransferAsset,
		Payload: &payload.TransferAsset{},
		Attributes: []*types.Attribute{
			{
				Usage: types.Nonce,
				Data:  randomNonceData(),
			},
		},
		Fee: 100,
	}
	protoTxSize := protoTx.GetSize()

	orderedList := newTxFeeOrderedList(onPopBack, uint64(protoTxSize*10))
	for i := 0; i < 10; i++ {
		tx := protoTx
		tx.Fee -= common.Fixed64(rand.Int63n(100))
		tx.Attributes = []*types.Attribute{
			{
				Usage: types.Nonce,
				Data:  randomNonceData(),
			},
		}

		assert.NoError(t, orderedList.AddTx(&tx))
		assert.False(t, firedPopBack)
	}
	assert.True(t, isListDescendingOrder(orderedList))
	assert.Equal(t, 10, orderedList.GetSize())

	protoTx.Fee = 0
	assert.True(t, orderedList.OverSize(uint64(protoTx.GetSize())))
	err := orderedList.AddTx(&protoTx)
	assert.True(t, err == addingTxExcluded)
	assert.False(t, firedPopBack)
	assert.True(t, isListDescendingOrder(orderedList))

	tx := protoTx
	tx.Fee = 1000
	tx.Attributes = []*types.Attribute{
		{
			Usage: types.Nonce,
			Data:  randomNonceData(),
		},
	}
	assert.NoError(t, orderedList.AddTx(&tx))
	assert.True(t, firedPopBack)
	assert.True(t, isListDescendingOrder(orderedList))

	tx = protoTx
	tx.Fee = 50 // set to the center
	tx.Attributes = []*types.Attribute{
		{
			Usage: types.Nonce,
			Data:  randomNonceData(),
		},
	}
	firedPopBack = false
	assert.NoError(t, orderedList.AddTx(&tx))
	assert.True(t, firedPopBack)
	assert.True(t, isListDescendingOrder(orderedList))
	assert.Equal(t, 10, orderedList.GetSize())
}

func TestTxFeeOrderedList_RemoveTx(t *testing.T) {
	orderedList := newTxFeeOrderedList(func(common.Uint256) {},
		pact.MaxTxPoolSize)

	protoTx := types.Transaction{
		TxType:  types.TransferAsset,
		Payload: &payload.TransferAsset{},
		Attributes: []*types.Attribute{
			{
				Usage: types.Nonce,
				Data:  randomNonceData(),
			},
		},
	}
	txSize := protoTx.GetSize()
	hashMap := make(map[common.Uint256]float64)
	for i := 0; i < 100; i++ {
		tx := protoTx
		tx.Attributes = []*types.Attribute{
			{
				Usage: types.Nonce,
				Data:  randomNonceData(),
			},
		}
		tx.Fee = common.Fixed64(rand.Int63n(1000) + 1)
		assert.NoError(t, orderedList.AddTx(&tx))
		hashMap[tx.Hash()] = float64(tx.Fee) / float64(txSize)

		assert.Equal(t, i+1, orderedList.GetSize())
	}

	i := 1
	for k, v := range hashMap {
		orderedList.RemoveTx(k, uint64(txSize), v)
		assert.Equal(t, 100-i, orderedList.GetSize())
		assert.True(t, isListDescendingOrder(orderedList))
		i++
	}
	assert.Equal(t, uint64(0), orderedList.totalSize)
}

func isListDescendingOrder(l *txFeeOrderedList) bool {
	for i := 0; i < len(l.list)-1; i++ {
		if l.list[i].FeeRate < l.list[i+1].FeeRate {
			return false
		}
	}
	return true
}

func randomNonceData() []byte {
	var data [20]byte
	rand.Read(data[:])
	return data[:]
}
