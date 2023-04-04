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

	//the sequence numeber for current pbft instance
	seq uint64
	//the round numeber for current pbft instance
	round uint64
	//block to be committed in current seq num; the value is nil when a new pbft instance starts
	estimateBlock *Block

	//BlockChain
	blockChain *BlockChain
	//logger
	logger *mylogger.MyLogger
	//rpc network
	peers []*myrpc.ClientEnd

	//message channel exapmle
	msgChan chan *myrpc.ConsensusMsg
}

func InitConsensus(config *Configuration) *Consensus {
	rand.Seed(time.Now().UnixNano())
	c := &Consensus{
		id:            config.Id,
		n:             config.N,
		port:          config.Port,
		proposalSize:  config.ProposalSize,
		seq:           0,
		round:         0,
		estimateBlock: nil,
		blockChain:    InitBlockChain(config.Id),
		logger:        mylogger.InitLogger("node", config.Id),
		peers:         make([]*myrpc.ClientEnd, 0),

		msgChan: make(chan *myrpc.ConsensusMsg, 1024),
	}
	for _, peer := range config.Committee {
		clientEnd := &myrpc.ClientEnd{Port: uint64(peer)}
		c.peers = append(c.peers, clientEnd)
	}
	go c.serve()

	return c
}
func (c *Consensus) RpcExample(args *myrpc.ConsensusMsg, reply *myrpc.ConsensusMsgReply) error {

	c.logger.DPrintf("Invoke RpcExample: receive message %v from %v at %v", args.Type.String(), args.From, time.Now().Nanosecond())
	c.msgChan <- args
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

func (c *Consensus) broadcastMessage(msg *myrpc.ConsensusMsg) {
	reply := &myrpc.ConsensusMsgReply{}
	for id := range c.peers {
		c.peers[id].Call("Consensus.RpcExample", msg, reply)
	}
}

func (c *Consensus) getCurrentLeader() uint8 {
	return uint8(c.round) % c.n
}

func (c *Consensus) startNewInstance() {
	// init a new pbft instance
	c.seq = c.seq + 1
	c.round = 0
	c.estimateBlock = nil
	c.tryPropose()
}

func (c *Consensus) tryPropose() {
	if c.id != c.getCurrentLeader() {
		return
	}
	//slow down
	time.Sleep(time.Duration(10) * time.Millisecond)
	block := c.getBlock()
	preprepareMsg := &myrpc.ConsensusMsg{
		Type:      myrpc.PREPREPARE,
		From:      c.id,
		Seq:       c.seq,
		Round:     c.round,
		BlockHash: Block2Hash(block),
		PrevBlock: block.PrevBlock,
		Data:      block.Data,
	}
	c.broadcastMessage(preprepareMsg)

}

func (c *Consensus) handlePreprepare(msg *myrpc.ConsensusMsg) {
	block := &Block{
		PrevBlock: msg.PrevBlock,
		Data:      msg.Data,
	}
	c.estimateBlock = block

	c.commitBlock(block)
	c.startNewInstance()
}

func (c *Consensus) handlePrepare(msg *myrpc.ConsensusMsg) {

}

func (c *Consensus) handleCommit(msg *myrpc.ConsensusMsg) {

}
func (c *Consensus) Run() {
	// wait for other node to start
	time.Sleep(time.Duration(1) * time.Second)
	// init rpc client
	for id := range c.peers {
		c.peers[id].Connect()
	}

	//propose in seq 0
	c.tryPropose()
	for {
		msg := <-c.msgChan
		switch msg.Type {
		case myrpc.PREPREPARE:
			c.handlePreprepare(msg)
		case myrpc.PREPARE:
			c.handlePrepare(msg)
		case myrpc.COMMIT:
			c.handleCommit(msg)
		}
	}
}
