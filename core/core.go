package core

import (
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"time"

	"nis3607/mylogger"
	"nis3607/myrpc"
)

type Consensus struct {
	id           uint8
	n            uint8
	port         uint64
	proposalSize uint64

	//BlockChain
	blockChain *BlockChain
	//logger
	logger *mylogger.MyLogger

	//rpc network
	peers []*myrpc.ClientEnd

	//message channel exapmle
	msgChan chan *Block
}

func InitConsensus(config *Configuration) *Consensus {
	rand.Seed(time.Now().UnixNano())
	c := &Consensus{
		id:           config.Id,
		n:            config.N,
		port:         config.Port,
		proposalSize: config.ProposalSize,
		blockChain:   InitBlockChain(config.Id),
		logger:       mylogger.InitLogger("node", config.Id),
		peers:        make([]*myrpc.ClientEnd, 0),
		msgChan:      make(chan *Block, 1024),
	}
	for _, peer := range config.Committee {
		clientEnd := &myrpc.ClientEnd{Port: uint64(peer)}
		c.peers = append(c.peers, clientEnd)
	}
	go c.serve()

	return c
}
func (c *Consensus) RpcExample(args *myrpc.ExampleArgs, reply *myrpc.ExampleReply) error {
	block := &Block{
		PrevBlock: args.PrevBlock,
		Data:      args.Data,
	}
	c.msgChan <- block
	c.logger.DPrintf("Invoke RpcExample: receive Block[%v(%v)] at %v", Block2Key(block), Hash2Key(block.PrevBlock), time.Now().Nanosecond())
	return nil
}

func (c *Consensus) serve() {
	rpc.Register(c)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":"+strconv.Itoa(int(c.port)))
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

func (c *Consensus) getBlock() *Block {
	//Generate a Proposal
	proposal := make([]byte, c.proposalSize)
	for i := uint64(0); i < c.proposalSize; i++ {
		proposal[i] = byte(rand.Intn(26) + 61)
	}
	block := c.blockChain.GenerateBlock(proposal)
	c.logger.DPrintf("Invoke getProposal: generated Block[%v(%v)] at %v", Block2Key(block), Hash2Key(block.PrevBlock), time.Now().Nanosecond())
	return block
}

func (c *Consensus) commitBlock(block *Block) {
	if ok := c.blockChain.CommitBlock(block); ok {
		c.logger.DPrintf("Invoke commitProposal: committed Block[%v(%v)] at %v", Block2Key(block), Hash2Key(block.PrevBlock), time.Now().Nanosecond())
	}
}

func (c *Consensus) Run() {
	// wait for other node to start
	time.Sleep(time.Duration(1) * time.Second)
	for {
		block := c.getBlock()
		args := &myrpc.ExampleArgs{From: c.id, PrevBlock: block.PrevBlock, Data: block.Data}
		reply := &myrpc.ExampleReply{}
		c.peers[c.id].Call("Consensus.RpcExample", args, reply)
		receivedBlock := <-c.msgChan
		c.commitBlock(receivedBlock)
		time.Sleep(time.Duration(1) * time.Second)
	}
}
