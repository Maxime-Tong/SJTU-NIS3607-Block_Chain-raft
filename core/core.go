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
	id   uint8
	n    uint8
	port uint64
	seq  uint64
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
		id:         config.Id,
		n:          config.N,
		port:       config.Port,
		seq:        0,
		blockChain: InitBlockChain(config.Id, config.BlockSize),
		logger:     mylogger.InitLogger("node", config.Id),
		peers:      make([]*myrpc.ClientEnd, 0),

		msgChan: make(chan *myrpc.ConsensusMsg, 1024),
	}
	for _, peer := range config.Committee {
		clientEnd := &myrpc.ClientEnd{Port: uint64(peer)}
		c.peers = append(c.peers, clientEnd)
	}
	go c.serve()

	return c
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

func (c *Consensus) OnReceiveMessage(args *myrpc.ConsensusMsg, reply *myrpc.ConsensusMsgReply) error {

	c.logger.DPrintf("Invoke RpcExample: receive message from %v at %v", args.From, time.Now().Nanosecond())
	c.msgChan <- args
	return nil
}

func (c *Consensus) broadcastMessage(msg *myrpc.ConsensusMsg) {
	reply := &myrpc.ConsensusMsgReply{}
	for id := range c.peers {
		c.peers[id].Call("Consensus.OnReceiveMessage", msg, reply)
	}
}

func (c *Consensus) handleMsgExample(msg *myrpc.ConsensusMsg) {
	block := &Block{
		Seq:  msg.Seq,
		Data: msg.Data,
	}
	c.blockChain.commitBlock(block)
}

func (c *Consensus) proposeLoop() {
	for {
		if c.id == 0 {
			block := c.blockChain.getBlock(c.seq)
			msg := &myrpc.ConsensusMsg{
				From: c.id,
				Seq:  block.Seq,
				Data: block.Data,
			}
			c.broadcastMessage(msg)
			c.seq++
		}
	}
}

func (c *Consensus) Run() {
	// wait for other node to start
	time.Sleep(time.Duration(1) * time.Second)
	//init rpc client
	for id := range c.peers {
		c.peers[id].Connect()
	}

	go c.proposeLoop()
	//handle received message
	for {

		msg := <-c.msgChan
		c.handleMsgExample(msg)
	}
}
