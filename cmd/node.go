package main

import (
	"flag"
	"fmt"
	"time"

	"nis3607/core"

	"github.com/schollz/progressbar/v3"
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
	bar := progressbar.Default(int64(*testTime*100), "正在运行:")
	for i := 0; i < *testTime*100; i++ {
		bar.Add(1)
		time.Sleep(time.Duration(10) * time.Millisecond)
	}
	fmt.Printf("Node %v finished test\n", *id)
	// time.Sleep(time.Duration(*testTime) * time.Second)
}
