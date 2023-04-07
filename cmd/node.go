package main

import (
	"flag"
	"fmt"
	"time"

	"nis3607/core"
)

func main() {
	//default: 7 nodes
	id := flag.Int("i", 0, "[node id]")
	testTime := flag.Int("t", 30, "[test time]")
	flag.Parse()
	config := core.GetConfig(*id)
	c := core.InitConsensus(config)
	//start to run node for testTime s
	go c.Run()

	time.Sleep(time.Duration(*testTime) * time.Second)
	fmt.Printf("Node %v finished test\n", *id)
}
