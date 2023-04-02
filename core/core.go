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

	//logger
	logger *mylogger.MyLogger

	//rpc network
	peers []*myrpc.ClientEnd
	//add your variants
}

func InitConsensus(config *Configuration) *Consensus {
	rand.Seed(time.Now().UnixNano())
	c := &Consensus{
		id:           config.Id,
		n:            config.N,
		port:         config.Port,
		proposalSize: config.ProposalSize,
		logger:       mylogger.InitLogger(config.Id),
		peers:        make([]*myrpc.ClientEnd, 0),
	}
	for _, peer := range config.Committee {
		clientEnd := &myrpc.ClientEnd{Port: uint64(peer)}
		c.peers = append(c.peers, clientEnd)
	}
	go c.serve()

	return c
}
func (c *Consensus) RpcExample(args *myrpc.ExampleArgs, reply *myrpc.ExampleReply) error {
	c.logger.DPrintf("Invoke RpcExample: received Proposal = %v from %v", string(args.Proposal[:16]), args.From)
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

func (c *Consensus) getProposal() []byte {
	//Generate a Proposal
	proposal := make([]byte, c.proposalSize)
	for i := uint64(0); i < c.proposalSize; i++ {
		proposal[i] = byte(rand.Intn(26) + 61)
	}
	c.logger.DPrintf("Invoke getProposal: generated proposal[%v] at %v", string(proposal[:16]), time.Now().Nanosecond())
	return proposal
}

func (c *Consensus) commitProposal(proposal []byte) {
	c.logger.DPrintf("Invoke commitProposal: committed proposal[%v] at %v", string(proposal[:16]), time.Now().Nanosecond())
}

func (c *Consensus) Run() {
	// wait for other node to start
	time.Sleep(time.Duration(2) * time.Second)
	for {
		proposal := c.getProposal()
		args := &myrpc.ExampleArgs{From: c.id, Proposal: proposal}
		reply := &myrpc.ExampleReply{}
		c.peers[c.id].Call("Consensus.RpcExample", args, reply)
		c.commitProposal(proposal)
		time.Sleep(time.Duration(1) * time.Second)
	}
}
