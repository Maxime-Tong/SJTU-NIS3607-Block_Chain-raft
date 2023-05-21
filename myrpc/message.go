package myrpc

// MessageType example
// type MessageType uint8

// const (
// 	PREPREPARE MessageType = iota
// 	PREPARE
// 	COMMIT
// )

// func (mt MessageType) String() string {
// 	switch mt {
// 	case PREPREPARE:
// 		return "PREPREPARE"
// 	case PREPARE:
// 		return "PREPARE"
// 	case COMMIT:
// 		return "COMMIT"
// 	}
// 	return "UNKNOW MESSAGETYPE"
// }

type Event struct {
	Message interface{}
	Reply   interface{}
	Err     chan error
}

type ConsensusMsg struct {
	// Type MessageType
	From uint8
	Seq  uint64
	Data []byte
}

type ConsensusMsgReply struct {
}

type RequestVoteMsg struct {
	Term         uint64
	LastLogIndex uint64
	LastLogTerm  uint64
	CandidateId  uint8
}

type RequestVoteReply struct {
	Term        uint64
	VoteGranted bool
}
