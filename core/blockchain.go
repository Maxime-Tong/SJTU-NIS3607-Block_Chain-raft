package core

import (
	"bytes"
	"nis3607/mylogger"
)

type Block struct {
	PrevBlock []byte
	Data      []byte
}

type BlockChain struct {
	gensisBlock        *Block
	lastCommittedBlock *Block
	Blocks             []*Block
	BlocksMap          map[string]*Block
	KeysMap            map[*Block]string
	logger             *mylogger.MyLogger
}

func InitBlockChain(id uint8) *BlockChain {
	blocks := make([]*Block, 1024)
	blocksMap := make(map[string]*Block)
	keysMap := make(map[*Block]string)
	//Generate gensis block
	prevBlock, _ := ComputeHash([]byte("NIS36107 blockchain"))
	data, _ := ComputeHash([]byte("NIS36107 consensus"))
	gensisBlock := &Block{
		PrevBlock: prevBlock,
		Data:      data,
	}
	blockChain := &BlockChain{
		gensisBlock:        gensisBlock,
		lastCommittedBlock: gensisBlock,
		Blocks:             blocks,
		BlocksMap:          blocksMap,
		KeysMap:            keysMap,
		logger:             mylogger.InitLogger("blockchain", id),
	}
	blockChain.AddBlockToChain(gensisBlock)
	return blockChain

}

func Block2Hash(block *Block) []byte {
	var buffer []byte
	buffer = append(buffer, block.PrevBlock...)
	buffer = append(buffer, block.Data...)
	hash, _ := ComputeHash(buffer)
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

func (bc *BlockChain) GetLastCommittedBlockHash() []byte {
	return Block2Hash(bc.lastCommittedBlock)
}

func (bc *BlockChain) AddBlockToChain(block *Block) {
	bc.Blocks = append(bc.Blocks, block)
	bc.KeysMap[block] = Block2Key(block)
	bc.BlocksMap[Block2Key(block)] = block
}

func (bc *BlockChain) GenerateBlock(data []byte) *Block {
	return &Block{
		PrevBlock: bc.GetLastCommittedBlockHash(),
		Data:      data,
	}
}

func (bc *BlockChain) CommitBlock(block *Block) bool {
	if !bytes.Equal(block.PrevBlock, bc.GetLastCommittedBlockHash()) {
		panic("The prev of block to commit is not the last committed block")
	}
	bc.logger.DPrintf("Committed Block:%v(%v)", Block2Key(block), Hash2Key(block.PrevBlock))
	bc.lastCommittedBlock = block
	if _, ok := bc.KeysMap[block]; !ok {
		bc.AddBlockToChain(block)
	}
	return true
}
