package Test

import (
	block "blockChain/Block"
	node "blockChain/Miner"
	"fmt"
	"sync"
	"time"
)

const MaxChannelSize = 1000

func Test(benNodeNum int, malNodeNum int) {
	//init create FirstBlock and chain
	totalNodeNum := benNodeNum + malNodeNum
	firstBlock := block.Block{Previous: "FIRST", Nonce: 0, Bits: 10, Timestamp: time.Now().UnixNano(), Data: "This is the first!", Height: 0}
	firstBlock.Blockhash = firstBlock.BlockHashCal()
	chain := make(map[string]block.Block)
	chain[firstBlock.Blockhash] = firstBlock

	peers := make(map[uint64]chan node.Message)
	peers[0] = make(chan node.Message)

	adminPeers := make(map[uint64]chan node.AdminMessage)
	adminMsgChan := make(chan node.AdminMessage)
	adminPeers[0] = adminMsgChan
	adminNode := node.Node{}
	adminNode.Init(0, peers, adminMsgChan, chain, 3)

	// add Node

	nodeAdminChan := make([]chan node.AdminMessage, totalNodeNum-1)
	newNode := make([]node.Node, totalNodeNum-1)
	for i := 0; i < totalNodeNum-1; i++ {
		newNode[i] = node.Node{}
		newNode[i].Init(uint64(i+1), peers, nodeAdminChan[i], adminNode.Buffer, adminNode.Bits)
		adminPeers[uint64(i)+1] = nodeAdminChan[i]
	}

	// newNode.Init(1, peers, nodeAdminChan, adminNode.Buffer, adminNode.Bits)

	fmt.Println("Init Complete")
	// peers test
	// for k, _ := range adminPeers {
	// 	fmt.Println(k)
	// }
	var wg sync.WaitGroup
	wg.Add(totalNodeNum)
	// start all nodes
	go func(i0 int) {
		adminNode.Run()
		wg.Done()
	}(0)
	for j := 0; j < benNodeNum-1; j++ {
		go func(i0 int) {
			newNode[i0].Run()
			wg.Done()
		}(j)
	}
	if malNodeNum > 0 {
		for j := benNodeNum - 1; j < totalNodeNum-1; j++ {
			go func(i0 int) {
				newNode[i0].MalRun()
				wg.Done()
			}(j)
		}
	}

	
	wg.Wait()
}

