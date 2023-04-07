package core

import (
	"math/rand"
	"nis3607/mylogger"
	"time"
)

type Block struct {
	Seq  uint64
	Data []byte
}

type BlockChain struct {
	Id        uint8
	BlockSize uint64
	Blocks    []*Block
	BlocksMap map[string]*Block
	KeysMap   map[*Block]string
	logger    *mylogger.MyLogger
}

func InitBlockChain(id uint8, blocksize uint64) *BlockChain {
	blocks := make([]*Block, 1024)
	blocksMap := make(map[string]*Block)
	keysMap := make(map[*Block]string)
	//Generate gensis block
	blockChain := &BlockChain{
		Id:        id,
		BlockSize: blocksize,
		Blocks:    blocks,
		BlocksMap: blocksMap,
		KeysMap:   keysMap,
		logger:    mylogger.InitLogger("blockchain", id),
	}
	return blockChain
}

func Block2Hash(block *Block) []byte {
	hash, _ := ComputeHash(block.Data)
	return hash
}

func Hash2Key(hash []byte) string {
	var key []byte
	for i := 0; i < 20; i++ {
		key = append(key, uint8(97)+uint8(hash[i]%(26)))
	}
	return string(key)
}
func Block2Key(block *Block) string {
	return Hash2Key(Block2Hash(block))
}

func (bc *BlockChain) AddBlockToChain(block *Block) {
	bc.Blocks = append(bc.Blocks, block)
	bc.KeysMap[block] = Block2Key(block)
	bc.BlocksMap[Block2Key(block)] = block
}

// Generate a Block
func (bc *BlockChain) getBlock(seq uint64) *Block {
	data := make([]byte, bc.BlockSize)
	for i := uint64(0); i < bc.BlockSize; i++ {
		data[i] = byte(rand.Intn(256))
	}
	block := &Block{
		Seq:  seq,
		Data: data,
	}
	bc.logger.DPrintf("generate Block[%v] in seq %v at %v", Block2Key(block), block.Seq, time.Now().Nanosecond())
	return block
}

func (bc *BlockChain) commitBlock(block *Block) {
	bc.AddBlockToChain(block)
	bc.logger.DPrintf("commit Block[%v] in seq %v at %v", Block2Key(block), block.Seq, time.Now().Nanosecond())
}
