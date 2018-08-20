package Blockchain

import (
	"github.com/btcsuite/btcd/btcjson"
	"github.com/romanornr/cyberchain/database"
	"encoding/gob"
	"bytes"
	"github.com/romanornr/cyberchain/blockdata"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

type Block struct {
	Block *btcjson.GetBlockVerboseResult
}

type BlockFinder interface { // might use to fetch
	FindBlock(hash *chainhash.Hash) (Block, error)
	//FindBlockByRPC(hash *chainhash.Hash) (Block)
}

type BlockList []Block
type BlockListCache []Block

var db = database.GetDatabaseInstance()

type BlockListProxy struct {
	Database *BlockList
	RPC *BlockList
	StackCache BlockListCache
	Stacksize int
	LastSearchUsedCache bool
}

// find block by looking into the database
// if the block is not in the database, check with an RPC call if it is.
// also add it in the databse if the RPC call has a result
func (b *BlockListProxy) FindBlock(hash *chainhash.Hash) (*btcjson.GetBlockVerboseResult, error) {

	//block := b.Database.FindBlock(hash.String())
	block := b.Database.FindBlock(hash.String())

	if block == nil {
		blockjson, _ := b.RPC.FindBlockByRPC(hash)
		b.AddBlockToDatabase(blockjson)
		return blockjson, nil
	}

	var blockjson *btcjson.GetBlockVerboseResult
	decoder := gob.NewDecoder(bytes.NewReader(block))
	decoder.Decode(&blockjson)

	return blockjson, nil

}

func (b *BlockList) FindBlockByRPC(hash *chainhash.Hash) (*btcjson.GetBlockVerboseResult, error){
	block, err := blockdata.GetBlock(hash)
	if err != nil {
		return nil, err
	}
	return block, nil
}

// find the block in the database by giving the blockhash
func (b *BlockList) FindBlock(hash string) []byte {
	return database.ViewBlock(hash)
}

func (b *BlockListProxy) AddBlockToDatabase(block *btcjson.GetBlockVerboseResult) {
	//b.Database.addBlock(block)
	//can do something like b.database.addTransaction(block)
	database.AddBlock(db, block.Hash, block)
}


//func (b *BlockList) addBlock(block *btcjson.GetBlockVerboseResult) {
//	database.AddBlock(db, block.Hash, block)
//}

// addBlockToStack takes the user argument and adds it to the stack in place.
// if the stack is full it removes the first element on it before adding.
func (b *BlockListProxy) addBlockToStack(block Block) {
	if len(b.StackCache) >=  b.Stacksize {
		b.StackCache = append(b.StackCache[1:], block)
	} else {
		b.StackCache.addBlockToCache(block)
	}
}

// add a new block to the end of the Block slice
func (b *BlockListCache) addBlockToCache(newBlock Block) {
	*b = append(*b, newBlock)
}