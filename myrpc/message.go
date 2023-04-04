package myrpc

type MessageType uint8

const (
	PREPREPARE MessageType = iota
	PREPARE
	COMMIT
)

func (mt MessageType) String() string {
	switch mt {
	case PREPREPARE:
		return "PREPREPARE"
	case PREPARE:
		return "PREPARE"
	case COMMIT:
		return "COMMIT"
	}
	return "UNKNOW MESSAGETYPE"
}

type ConsensusMsg struct {
	Type      MessageType
	From      uint8
	Seq       uint64
	Round     uint64
	BlockHash []byte
	PrevBlock []byte //useful for preprepare message
	Data      []byte //useful for preprepare message
}

type ConsensusMsgReply struct {
}
