package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/davecgh/go-spew/spew"

	"github.com/gorilla/mux"
)

// Block models the Data for a block in the Blockchain
type Block struct {
	Index     int
	Timestamp string
	BPM       int
	Hash      string
	PrevHash  string
}

// Message models the data for a request to create a new Block
type Message struct {
	BPM int
}

// Blockchain is a variable holding the chain of Blocks
var Blockchain []Block

func calculateHash(block Block) string {
	record := fmt.Sprintf("%d%s%d%s", block.Index, block.Timestamp, block.BPM, block.PrevHash)
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

func run() error {
	mux := makeMuxRouter()
	httpPort := os.Getenv("PORT")
	log.Println("Listening on port: ", httpPort)
	s := &http.Server{
		Addr:           fmt.Sprintf(":%s", httpPort),
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", handleGetBlockchain).Methods("GET")
	muxRouter.HandleFunc("/", handleWriteNewBlock).Methods("POST")
	return muxRouter
}

func handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(Blockchain, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	io.WriteString(w, string(bytes))
}

func handleWriteNewBlock(w http.ResponseWriter, r *http.Request) {
	var message Message

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&message); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}

	defer r.Body.Close()

	prevBlock := Blockchain[len(Blockchain)-1]

	newBlock, err := generateBlock(prevBlock, message.BPM)
	if err != nil {
		respondWithJSON(w, r, http.StatusInternalServerError, r.Body)
		return
	}

	if isBlockValid(newBlock, prevBlock) {
		newBlockchain := append(Blockchain, newBlock)
		replaceChain(newBlockchain)
		spew.Dump(Blockchain)
	}

	respondWithJSON(w, r, http.StatusCreated, newBlock)
}

func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}

	w.WriteHeader(code)
	w.Write(response)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	go func() {
		t := time.Now()
		genesisBlock := Block{0, t.String(), 0, "", ""}
		spew.Dump(generateBlock)
		Blockchain = append(Blockchain, genesisBlock)
	}()

	log.Fatal(run())
}
