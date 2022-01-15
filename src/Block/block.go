package Block

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

const (
	EMPTY_BLOCK = "-1"

	NO_HEIGHT = -1
)

type Block struct {
	// hash of the previous block
	Previous string
	// solution: like bitcoin
	Nonce uint64
	//  difficulty : how many zeros required
	Bits uint64
	// timestamp
	Timestamp int64

	Blockhash string
	/*-------------data part----------------*/
	// using a transformation of Bits and Previous as data
	Data string
	//  using in judge which chain is longer
	Height int64
}

// only used in hash calculation
type BlockHead struct {
	Previous string `json:"prev"`
	Time     int64  `json:"time"`
	Bits     uint64 `json:"bits"`
	Nonce    uint64 `json:"nonce"`
}

func (block *Block) Init(prev string, bits uint64) {
	block.Previous = prev
	block.Bits = bits
	// using random value to init nonce
	rand.Seed(time.Now().UnixNano())
	block.Nonce = uint64(rand.Intn(100000000))
	block.Timestamp = time.Now().UnixNano()
	block.Data = fmt.Sprintf("%s%d", block.Previous, block.Bits)
	block.Blockhash = block.BlockHashCal()
	block.Height = NO_HEIGHT
}

func (block *Block) BlockHashCal() string {
	tmp := BlockHead{
		Previous: block.Previous,
		Time:     block.Timestamp,
		Bits:     block.Bits,
		Nonce:    block.Nonce,
	}
	text, err := json.Marshal(tmp)
	if err != nil {
		return block.Blockhash
	}
	return CalculateHash(text)
}

func CalculateHash(text []byte) string {
	h := sha256.New()
	h.Write([]byte(text))
	hash := hex.EncodeToString(h.Sum(nil))
	return hash
}

// use to print the whole chain for a node buffer
func PrintBlock(block Block) {
	fmt.Printf("BLOCK: prev %.5s, self %.5s, height %d \n", block.Previous, block.Blockhash, block.Height)
}

// func PrintChain(chain map[string]Block) {
// 	fmt.Printf("INFO: current chain:\n")
// 	for _, block := range chain {
// 		PrintBlock(block)
// 	}
// 	time.Sleep(time.Second)
// }

// check  whether the chain satisfies difficulty (bits)
func (block *Block) IsValid() bool {
	// for test only
	if block.Blockhash != block.BlockHashCal() {
		fmt.Printf("ERROR: Bad Block height %d, hash %s\n", block.Height, block.Blockhash)
		return false
	}
	data := fmt.Sprintf("%s%d", block.Data, block.Nonce)
	dataHash := CalculateHash([]byte(data))

	for i := uint64(0); i <= block.Bits; i++ {
		if string(dataHash[i]) != "0" {
			return false
		}
	}
	// fmt.Println(dataHash)
	return true
}

// Height calculation, find where the block is in the chain
func (block Block) GetBlockHeight(chain map[string]Block) int64 {
	// fmt.Printf("INFO: getting height of Block %s\n", block.Blockhash)
	if block.Previous == "FIRST" {
		return 0
	}
	if block.Height >= 0 {
		return block.Height
	}

	// find Previous Block in the map
	previousBlock, ifFind := chain[block.Previous]
	//debugging
	if !ifFind {
		if block.Previous == "FIRST" {
			// fmt.Println("INFO:in")
			return 0
		}
		fmt.Printf("ERROR: BLOCK %.5s NO prev block\n", block.Blockhash)
		// fmt.Println("warning！ no previous block find")
		return NO_HEIGHT
	} else {
		preHeight := previousBlock.GetBlockHeight(chain)
		if preHeight <= -1 {
			return NO_HEIGHT
		}
		// fmt.Println("INFO: prevHeight=", previousBlock.Height)
		block.Height = previousBlock.Height + 1
		// fmt.Println("INFO: height=", block.Height)
		chain[block.Blockhash] = block
		return chain[block.Blockhash].Height
	}

}

// to find the highest code in the chain
func GetLastBlock(chain map[string]Block) Block {
	maxHeight := int64(-1)
	maxBlock := Block{}
	for _, block := range chain {
		if block.Height > maxHeight {
			maxHeight = block.Height
			maxBlock = block
		}
	}
	// troubleShooting
	if maxHeight == -1 {
		fmt.Println("Warning: MaxHeight unrecognized")
		return Block{}
	} else {
		return maxBlock
	}

}

// used in changing difficulty
func (block *Block) setDifficulty(bits uint64) {
	block.Bits = bits
	block.Blockhash = block.BlockHashCal()
}
