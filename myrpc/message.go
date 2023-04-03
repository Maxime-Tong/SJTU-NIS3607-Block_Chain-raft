package myrpc

type ExampleArgs struct {
	From      uint8
	PrevBlock []byte
	Data      []byte
}

type ExampleReply struct {
}
