package myrpc

import (
	"fmt"
	"net/rpc"
	"strconv"
)

type ClientEnd struct {
	Port      uint64
	rpcClient *rpc.Client
}

func (e *ClientEnd) Connect() {
	c, err := rpc.DialHTTP("tcp", "127.0.0.1:"+strconv.Itoa(int(e.Port)))
	if err == nil {
		e.rpcClient = c
	}
	// log.Fatal("dialing:", err)

}

func (e *ClientEnd) Call(svcMeth string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1:"+strconv.Itoa(int(e.Port)))
	// if err != nil {
	// 	log.Fatal("dialing:", err)
	// }
	// defer c.Close()
	if e.rpcClient == nil {
		return false
	}

	err := e.rpcClient.Call(svcMeth, args, reply)
	if err == nil {
		return true
	}

	//exception handling: may reconnect or retransmit
	fmt.Println(err)
	return false

}
