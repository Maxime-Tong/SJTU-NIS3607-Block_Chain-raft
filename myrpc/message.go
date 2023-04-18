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

type ConsensusMsg struct {
	// Type MessageType
	From uint8
	Seq  uint64
	Data []byte
}

type ConsensusMsgReply struct {
}
