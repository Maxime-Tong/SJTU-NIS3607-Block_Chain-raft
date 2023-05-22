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

type State uint64

const (
	Follower State = iota

	Candidate

	Leader
)

const (
	DefaultHeartbeatInterval = 25 * time.Millisecond

	DefaultElectionTimeout = 150 * time.Millisecond
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
	// msgChan chan *myrpc.ConsensusMsg
	eventChan chan *myrpc.Event

	state       State
	currentTerm uint64
	blockTerms  []uint64
	nextIndex   []uint64
	matchIndex  []uint64
	reservedLog []*Block
	votedFor    uint8
	commitIndex uint64

	electionTimeout   time.Duration
	heartbeatInterval time.Duration
	quorumSize        uint64
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

		// msgChan: make(chan *myrpc.ConsensusMsg, 1024),
		eventChan: make(chan *myrpc.Event, 1024),

		state:       Follower,
		currentTerm: 0,
		votedFor:    config.N,
		nextIndex:   make([]uint64, config.N),
		matchIndex:  make([]uint64, config.N),
		blockTerms:  make([]uint64, 0, 1024),
		reservedLog: make([]*Block, 0, 1024),
		commitIndex: 0,

		electionTimeout:   DefaultElectionTimeout,
		heartbeatInterval: DefaultHeartbeatInterval,
		quorumSize:        uint64(config.N)/2 + 1,
	}

	c.appendBlocks(&Block{Seq: c.seq})

	for _, peer := range config.Committee {
		clientEnd := &myrpc.ClientEnd{Port: uint64(peer)}
		c.peers = append(c.peers, clientEnd)
	}
	go c.serve()

	return c
}

func (c *Consensus) followerLoop() {
	c.logger.DPrintf("become follower, terms: %v", c.currentTerm)
	since := time.Now()
	timeoutC := afterBetween(c.electionTimeout, c.electionTimeout*2)

	for c.state == Follower {
		update := false
		select {
		case event := <-c.eventChan:
			switch msg := event.Message.(type) {
			case *HeartbeatMsg:
				update = c.handleHeartbeatMsg(msg, event.Reply.(*HeartbeatReply))
			case *myrpc.RequestVoteMsg:
				update = c.handleRequestVoteMsg(msg, event.Reply.(*myrpc.RequestVoteReply))
			}
			// call back to event
			event.Err <- nil
		case <-timeoutC:
			elapsedTime := time.Now().Sub(since)
			c.logger.DPrintf("timeout, try to become candidate, %d", elapsedTime)
			c.state = Candidate
			// if c.id == 1 {
			// 	c.state = Candidate
			// } else {
			// 	update = true
			// }
		}

		if update {
			timeoutC = afterBetween(c.electionTimeout, c.electionTimeout*2)
			since = time.Now()
		}
	}
}

func (c *Consensus) candidateLoop() {
	c.logger.DPrintf("become candidate, terms: %v", c.currentTerm)
	var votesCnt uint64 = 1
	doVote := true

	var timeoutC <-chan time.Time
	var votesReplyC chan *myrpc.RequestVoteReply

	for c.state == Candidate {
		if doVote {
			c.currentTerm++
			c.votedFor = c.id
			votesCnt = 1
			timeoutC = afterBetween(c.electionTimeout, c.electionTimeout*2)

			votesReplyC = make(chan *myrpc.RequestVoteReply, c.n)
			msg := &myrpc.RequestVoteMsg{
				Term:         c.currentTerm,
				LastLogIndex: c.seq - 1,
				LastLogTerm:  c.blockTerms[c.seq-1],
				CandidateId:  c.id,
			}

			for id := range c.peers {
				if id == int(c.id) {
					continue
				}

				go func(p *myrpc.ClientEnd) {
					reply := &myrpc.RequestVoteReply{}
					p.Call("Consensus.OnReceiveRequestVoteMsg", msg, reply)
					votesReplyC <- reply
				}(c.peers[id])
			}

			doVote = false
		}

		select {
		case votesReply := <-votesReplyC:
			if votesReply.VoteGranted && votesReply.Term == c.currentTerm {
				votesCnt++
				if votesCnt == c.quorumSize {
					c.state = Leader
					return
				}
			}
			if votesReply.Term > c.currentTerm {
				c.state = Follower
				c.currentTerm = votesReply.Term
				c.votedFor = c.n
				return
			}
		case event := <-c.eventChan:
			switch msg := event.Message.(type) {
			case *HeartbeatMsg:
				_ = c.handleHeartbeatMsg(msg, event.Reply.(*HeartbeatReply))
			case *myrpc.RequestVoteMsg:
				_ = c.handleRequestVoteMsg(msg, event.Reply.(*myrpc.RequestVoteReply))
			}
			// call back to event
			event.Err <- nil
		case <-timeoutC:
			doVote = true
		}
	}
}

func (c *Consensus) leaderLoop() {
	c.logger.DPrintf("become leader, terms: %v", c.currentTerm)
	ticker := time.NewTicker(c.heartbeatInterval)
	defer ticker.Stop()
	heartbeatC := make(chan *HeartbeatReply, 1024)

	for id := range c.peers {
		c.nextIndex[id] = c.seq
		c.matchIndex[id] = 0
	}
	for c.state == Leader {
		// c.logger.DPrintf("len: [%v, %v, %v], commit: %v, seq: %v", len(c.blockChain.Blocks), len(c.blockTerms), len(c.reservedLog), c.commitIndex, c.seq)
		select {
		case <-ticker.C: // heartbeat
			c.logger.DPrintf("send heartbeat!")

			for id := range c.peers {
				if id == int(c.id) {
					continue
				}

				prevLogIndex := c.nextIndex[id] - 1
				msg := &HeartbeatMsg{
					Term:         c.currentTerm,
					PrevLogIndex: prevLogIndex,
					PrevLogTerm:  c.blockTerms[prevLogIndex],
					EntryTerms:   c.blockTerms[c.nextIndex[id]:],
					Entries:      c.reservedLog[c.nextIndex[id]:],
					CommitIndex:  c.commitIndex,
					Leaderid:     c.id,
				}
				go func(p *myrpc.ClientEnd) {
					reply := &HeartbeatReply{}
					p.Call("Consensus.OnReceiveHeartbeat", msg, reply)
					// c.logger.DPrintf("Heartbeat Reply %v", reply)
					heartbeatC <- reply
				}(c.peers[id])
			}

			// c.propose()
		case msg := <-heartbeatC:
			// c.logger.DPrintf("process heartbeatreply")
			c.handleHeartbeatReply(msg)
		case event := <-c.eventChan:
			// c.logger.DPrintf("process msgC")
			switch msg := event.Message.(type) {
			case *myrpc.RequestVoteMsg:
				_ = c.handleRequestVoteMsg(msg, event.Reply.(*myrpc.RequestVoteReply))
			case *HeartbeatMsg:
				_ = c.handleHeartbeatMsg(msg, event.Reply.(*HeartbeatReply))
			}
			// call back to event
			event.Err <- nil
		}

		c.propose()
	}
}

func (c *Consensus) loop() {
	state := c.state
	for {
		// c.logger.DPrintf("current state: %v, c.seq: %v, term: %v", state, c.seq, c.currentTerm)
		switch state {
		case Follower:
			c.followerLoop()
		case Candidate:
			c.candidateLoop()
		case Leader:
			c.leaderLoop()
		}
		state = c.state
	}
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

func (c *Consensus) hasVoted() bool {
	return c.votedFor != c.n
}

// we did not design mutex locks, so this function can only be used in main thread!
func (c *Consensus) handleRequestVoteMsg(msg *myrpc.RequestVoteMsg, reply *myrpc.RequestVoteReply) bool {

	reply.Term = c.currentTerm
	reply.VoteGranted = false

	if msg.Term < c.currentTerm {
		return false
	} else if msg.Term > c.currentTerm {
		// c.logger.DPrintf("request vote state: %v, terms: %v, msg: %v", c.state, c.currentTerm, msg)
		c.updateCurrentTerm(msg.Term)
	} else if c.hasVoted() && c.votedFor != msg.CandidateId {
		return false
	}

	lastLogIndex := c.seq - 1
	if lastLogIndex > msg.LastLogIndex || c.blockTerms[lastLogIndex] > msg.LastLogTerm {
		return false
	}

	reply.Term = c.currentTerm
	c.votedFor = msg.CandidateId
	reply.VoteGranted = true
	return true
}

func (c *Consensus) handleHeartbeatMsg(msg *HeartbeatMsg, reply *HeartbeatReply) bool {
	reply.Term = c.currentTerm
	reply.Success = false
	reply.LastLogIndex = Min(msg.PrevLogIndex, c.seq-1)
	reply.From = c.id

	if msg.Term < c.currentTerm {
		return false
	} else if msg.Term > c.currentTerm {
		c.updateCurrentTerm(msg.Term)
	} else {
		if c.state == Candidate {
			c.state = Follower
		}
	}

	if c.seq-1 < msg.PrevLogIndex || c.blockTerms[msg.PrevLogIndex] != msg.PrevLogTerm {
		return true
	}

	c.reservedLog = append(c.reservedLog[:msg.PrevLogIndex+1], msg.Entries...)
	c.seq = uint64(len(c.reservedLog))
	c.blockTerms = append(c.blockTerms[:msg.PrevLogIndex+1], msg.EntryTerms...)

	if msg.CommitIndex > c.commitIndex {
		for i := c.commitIndex + 1; i <= msg.CommitIndex; i++ {
			block := c.reservedLog[i]
			c.blockChain.commitBlock(block)
		}
		c.commitIndex = msg.CommitIndex
	}

	reply.Term = c.currentTerm
	reply.Success = true
	reply.LastLogIndex = c.seq - 1
	// c.logger.DPrintf("committed index: %v, len: [%v, %v]", c.commitIndex, len(c.reservedLog), len(c.blockTerms))
	return true
}

func (c *Consensus) handleHeartbeatReply(msg *HeartbeatReply) {
	if msg.Term > c.currentTerm {
		c.updateCurrentTerm(msg.Term)
		return
	}

	if !msg.Success && c.nextIndex[msg.From] > 0 {
		c.nextIndex[msg.From] = Min(c.nextIndex[msg.From]-1, msg.LastLogIndex+1)
		return
	}
	if msg.Success {
		c.nextIndex[msg.From] = msg.LastLogIndex + 1
		c.matchIndex[msg.From] = msg.LastLogIndex
		committedIndex := c.commitIndex
		for i := committedIndex + 1; i <= msg.LastLogIndex; i++ {
			matchCnt := uint64(1)
			for pid := range c.peers {
				if c.matchIndex[pid] >= i {
					matchCnt += 1
				}
			}
			if matchCnt >= c.quorumSize {
				block := c.reservedLog[i]
				c.blockChain.commitBlock(block)
				c.commitIndex = i
			}
		}
	}
}

func (c *Consensus) OnReceiveRequestVoteMsg(args *myrpc.RequestVoteMsg, reply *myrpc.RequestVoteReply) error {
	event := &myrpc.Event{
		Message: args,
		Reply:   reply,
		Err:     make(chan error, 1),
	}
	c.eventChan <- event

	// wait for call back
	err := <-event.Err
	return err
}

func (c *Consensus) OnReceiveHeartbeat(args *HeartbeatMsg, reply *HeartbeatReply) error {
	event := &myrpc.Event{
		Message: args,
		Reply:   reply,
		Err:     make(chan error, 1),
	}
	c.eventChan <- event

	// wait for call back
	err := <-event.Err
	return err
}

func (c *Consensus) updateCurrentTerm(newTerm uint64) {
	c.state = Follower
	c.currentTerm = newTerm
	c.votedFor = c.n
}

func (c *Consensus) appendBlocks(block *Block) {
	c.reservedLog = append(c.reservedLog, block)
	c.blockTerms = append(c.blockTerms, c.currentTerm)
	c.seq++
}

func (c *Consensus) propose() {
	block := c.blockChain.getBlock(c.seq)
	c.appendBlocks(block)
}

func (c *Consensus) Run() {
	// wait for other node to start
	time.Sleep(time.Duration(1) * time.Second)
	//init rpc client
	for id := range c.peers {
		c.peers[id].Connect()
	}

	c.loop()
}
