package Miner

import (
	"fmt"
	"os"
	"time"

	block "blockChain/Block"
)

const (
	DEBUG_TIME = 10
	ADMIN_ID   = 0
)

// message sent to each node
type Message struct {
	sender   uint64
	newBlock block.Block
	isMal    bool
}

// message from administrator nide
type AdminMessage struct {
	Sender  uint64
	ifStop  bool
	newBits uint64
}

// miner defined as node
type Node struct {
	// the identifier of node
	id    uint64
	// the message channel to broadcast block
	peers       map[uint64]chan Message
	// channel for each node
	receiveChan chan Message
	// chain buffer
	Buffer      map[string]block.Block
	// receive message (like stop and change bits) info from admin
	adminChan   chan AdminMessage
	// current difficulty 
	Bits        uint64
}

// type MalNode struct {
// 	id          uint64
// 	peers       map[uint64]chan Message
// 	receiveChan chan Message
// 	Buffer      map[string]block.Block
// 	adminChan   chan AdminMessage
// 	Bits        uint64
// 	MalHeight   uint64
// }

func (n *Node) Init(id uint64, peers map[uint64]chan Message, adminChan chan AdminMessage, buffer map[string]block.Block, Bits uint64) {
	n.id = id
	n.peers = peers
	n.Buffer = make(map[string]block.Block)
	for blockID, block := range buffer {
		n.Buffer[blockID] = block
	}
	n.Bits = Bits
	n.receiveChan = make(chan Message)

	n.adminChan = adminChan
	n.peers = peers
	peers[n.id] = n.receiveChan
	fmt.Printf("Node Init: number %d\n", n.id)
}

//run start 1 node and start mining
func (n *Node) Run() {
	// for testing info
	// defer fmt.Printf("INFO: node %d exited \n", n.id)

	ticker := time.NewTicker(time.Second * DEBUG_TIME)
	defer ticker.Stop()

	// for testing growth rate
	prevChainHeight := 0
	currentChainHeight := 0
	// find the highest block in the chain and mine from that block
	fmt.Println("Node start : number ", n.id)
	startBlock := block.GetLastBlock(n.Buffer)
loop:
	for {
		select {
		case msg := <-n.receiveChan:
			{
				// simulation:skip all the malicious block
				if msg.isMal {
					fmt.Printf("INFO: node %d skipped malBlock\n", n.id)
					continue
				} else {
					fmt.Printf("INFO: node %d received INFO from %d\n", n.id, msg.sender)
					//test if the new block is valid
					if !msg.newBlock.IsValid() {
						fmt.Printf("Error: invalid new block received from %d Node to %d Node", msg.sender, n.id)
						break loop
					}
					// calculate if the height is bigger than current one
					// msg.newBlock.Height = int64(block.NO_HEIGHT)

					n.Buffer[msg.newBlock.Blockhash] = msg.newBlock
					// h := msg.newBlock.GetBlockHeight(n.Buffer)
					// msg.newBlock.Height = h
					// n.Buffer[msg.newBlock.Blockhash] = msg.newBlock
					if msg.newBlock.Height == int64(-1) {
						fmt.Printf("Error: invalid new block, no previous Block\n")
					} else {
						if msg.newBlock.Height > startBlock.Height {
							startBlock = msg.newBlock
							fmt.Printf("INFO: node %d change the startBlock\n", n.id)
						}
					}
				}
			}
		case adminMsg := <-n.adminChan:
			{	
				os.Exit(0)
				if adminMsg.ifStop {
					break loop
				} else {
					n.Bits = adminMsg.newBits
				}
			}
		case <-ticker.C:
			{
				currentChainHeight = int(block.GetLastBlock(n.Buffer).Height)
				chainGrowth := currentChainHeight - prevChainHeight
				// test chain growth rate
				if n.id == ADMIN_ID {
					fmt.Printf("GROWTH RATE INFO: %d blocks generated during last %d seconds\n", chainGrowth, DEBUG_TIME)
				}
				prevChainHeight = currentChainHeight
				n.PrintBuffer()
				break loop
			}
		default:
			startBlock = n.Mine(startBlock)
		}
	}
}

/*mining function for every single node
return the target for next mining
*/
func (n *Node) Mine(startBlock block.Block) block.Block {
	newBlock := block.Block{}
	newBlock.Init(startBlock.Blockhash, n.Bits)
	if newBlock.IsValid() {
		n.Buffer[newBlock.Blockhash] = newBlock
		newBlock.Height = newBlock.GetBlockHeight(n.Buffer)
		//broadcast msg to node
		fmt.Printf("INFO: node %d has generated new block with hash %.5s\n", n.id, newBlock.Blockhash)
		msg := Message{sender: n.id, newBlock: newBlock, isMal: false}
		n.Broadcast(msg)
		return newBlock
	} else {
		// fmt.Printf("INFO: node %d has calculated a wrong block with nonce %d \n", n.id, newBlock.Nonce)
		return startBlock
	}
}

// Malrun start 1 malicious node and start mining in a short chain
func (n *Node) MalRun() {
	defer fmt.Printf("INFO: MalNode %d exited \n", n.id)

	ticker := time.NewTicker(time.Second * DEBUG_TIME)
	defer ticker.Stop()

	// find the highest block in the chain and mine from that block
	fmt.Println("MalNode start : number ", n.id)
	startBlock := block.GetLastBlock(n.Buffer)
	// store the height of the benign block
	originalHeight := startBlock.Height
	benignHeight := int64(block.NO_HEIGHT)
loop:
	for {
		select {
		case msg := <-n.receiveChan:
			{
				// if the node is benign we skip the node and compare the length
				if msg.isMal {
					if !msg.newBlock.IsValid() {
						fmt.Printf("Error: invalid new block received from %d Node to %d Node", msg.sender, n.id)
						break loop
					}
					// calculate if the height is bigger than current one
					msg.newBlock.Height = int64(block.NO_HEIGHT)

					n.Buffer[msg.newBlock.Blockhash] = msg.newBlock
					h := msg.newBlock.GetBlockHeight(n.Buffer)
					msg.newBlock.Height = h
					n.Buffer[msg.newBlock.Blockhash] = msg.newBlock
					if msg.newBlock.Height == int64(-1) {
						fmt.Printf("Error: invalid new block, no previous Block")
					} else {
						if msg.newBlock.Height > startBlock.Height {
							startBlock = msg.newBlock
							fmt.Printf("INFO: Malnode %d change the startBlock\n", n.id)
						}
					}

				} else {
					benignHeight = int64(msg.newBlock.Height)
				}

			}
		case adminMsg := <-n.adminChan:
			{	
				if adminMsg.ifStop {
					break loop
				} else {
					n.Bits = adminMsg.newBits
				}
			}
		case <-ticker.C:
			{
				// debugging
				n.PrintBuffer()
				break loop
			}
		default:
			startBlock = n.MalMine(startBlock, benignHeight, originalHeight)
		}
	}
}

// Malmine is mining function for malcode, if the block height surpasses the benign height
// Malnode will broadcast to all the good nodes.
func (n *Node) MalMine(startBlock block.Block, benignHeight int64, originHeight int64) block.Block {
	newBlock := block.Block{}
	newBlock.Init(startBlock.Blockhash, n.Bits)
	if newBlock.IsValid() {
		n.Buffer[newBlock.Blockhash] = newBlock
		newBlock.Height = newBlock.GetBlockHeight(n.Buffer)
		//broadcast msg to Malnode
		fmt.Printf("INFO: Malnode %d has generated new block with hash %.5s\n", n.id, newBlock.Blockhash)
		msg := Message{sender: n.id, newBlock: newBlock, isMal: true}
		n.Broadcast(msg)
		//check if broadcast
		// if the benign chain has gone 2 blocks forward and the malchain goes before the benign chain, we will broadcast
		if newBlock.Height > benignHeight && benignHeight > originHeight {
			fmt.Printf("INFO:ATTACK SUCCESS! current benign height: %d, mal height %d\n", benignHeight, newBlock.Height)
			for _, block := range n.Buffer {
				if block.Height > originHeight {
					msg := Message{sender: n.id, newBlock: block, isMal: false}
					fmt.Printf("INFO: ATTACKER block sent\n")
					n.Broadcast(msg)
				}
			}
		}
		return newBlock

	} else {
		// fmt.Printf("INFO: node %d has calculated a wrong block with nonce %d \n", n.id, newBlock.Nonce)
		return startBlock
	}
}

func (n *Node) Broadcast(msg Message) {
	for id, ch := range n.peers {
		if id == n.id {
			continue
		}
		ch <- msg
	}
}

func (n *Node) AdminBroadcast(msg AdminMessage, adminPeers map[uint64]chan AdminMessage) {

	for id, ch := range adminPeers {
		
		if id == n.id {
			continue
		}
		ch <- msg
	}
}

func (n *Node) PrintBuffer() {
	fmt.Printf("INFO: current chain for node %d:\n,height: %d", n.id, block.GetLastBlock(n.Buffer).Height)
	for _, blk := range n.Buffer {
		block.PrintBlock(blk)
	}
}

func (n *Node) AdjustBits(bits uint64,adminPeers map[uint64]chan AdminMessage) {
	if (n.id!=0){
		return 
	}else{
		adminMsg:=AdminMessage{Sender:0, ifStop:false,newBits:bits}
		n.AdminBroadcast(adminMsg,adminPeers)
	}
}

func (n *Node)StopAll(adminPeers map[uint64]chan AdminMessage) {
	if (n.id!=0){
		fmt.Println("STOPSTOP")
		return 
	}else{

		adminMsg:=AdminMessage{Sender:0, ifStop:true,newBits:n.Bits}
		n.AdminBroadcast(adminMsg,adminPeers)
	}
}
