package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type Block struct {
	Index     int
	Timestamp string
	BPM       int
	Hash      string
	PrevHash  string
}

var Blockchain []Block

func calculateHash(block Block) string {
	record := fmt.Sprintf("%s%s%s%s", block.Index, block.Timestamp, block.BPM, block.PrevHash)
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

func generateBlock(prevBlock Block, bpm int) (Block, error) {
	var newBlock Block

	t := time.Now()

	newBlock.Index = prevBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.BPM = bpm
	newBlock.PrevHash = prevBlock.Hash
	newBlock.Hash = calculateHash(newBlock)

	return newBlock, nil
}

func isBlockValid(block Block, prevBlock Block) bool {
	wrongIndex := block.Index != prevBlock.Index+1
	wrongPrevHash := block.PrevHash != prevBlock.Hash
	wrongHash := block.Hash != calculateHash(block)

	if wrongIndex || wrongPrevHash || wrongHash {
		return false
	}

	return true
}

func replaceChain(newBlocks []Block) {
	if len(newBlocks) > len(Blockchain) {
		Blockchain = newBlocks
	}
}
