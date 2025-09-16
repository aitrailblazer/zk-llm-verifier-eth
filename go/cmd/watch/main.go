package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const settlementABI = `
[
  {
    "anonymous": false,
    "inputs": [
      { "indexed": true,  "internalType": "bytes32", "name": "modelId", "type": "bytes32" },
      { "indexed": true,  "internalType": "bytes32", "name": "inputHash", "type": "bytes32" },
      { "indexed": false, "internalType": "bytes32", "name": "outputCommitment", "type": "bytes32" },
      { "indexed": true,  "internalType": "address", "name": "prover", "type": "address" },
      { "indexed": false, "internalType": "bytes32", "name": "attestationId", "type": "bytes32" }
    ],
    "name": "InferenceRecorded",
    "type": "event"
  }
]
`

func main() {
	rpc := flag.String("rpc", "http://localhost:8545", "RPC URL (e.g., http://localhost:8545)")
	contractHex := flag.String("contract", "", "Contract address (0x...)")
	from := flag.Int64("from", -1, "Start block (default: latest block)")
	flag.Parse()

	if *contractHex == "" {
		log.Fatal("missing --contract (deployed contract address)")
	}
	contract := common.HexToAddress(*contractHex)

	client, err := ethclient.Dial(*rpc)
	if err != nil {
		log.Fatalf("dial rpc: %v", err)
	}
	parsed, err := abi.JSON(strings.NewReader(settlementABI))
	if err != nil {
		log.Fatalf("parse abi: %v", err)
	}

	ctx := context.Background()
	var start *big.Int
	if *from >= 0 {
		start = big.NewInt(*from)
	} else {
		h, err := client.HeaderByNumber(ctx, nil)
		if err != nil {
			log.Fatalf("get latest header: %v", err)
		}
		start = h.Number
	}

	topic0 := crypto.Keccak256Hash([]byte("InferenceRecorded(bytes32,bytes32,bytes32,address,bytes32)"))

	q := ethereum.FilterQuery{
		FromBlock: start,
		Addresses: []common.Address{contract},
		Topics:    [][]common.Hash{{topic0}},
	}

	fmt.Printf("Watching InferenceRecorded on %s from block %s (rpc=%s)\n", contract.Hex(), start.String(), *rpc)

	// historical
	logs, err := client.FilterLogs(ctx, q)
	if err != nil {
		log.Fatalf("filter logs: %v", err)
	}
	for _, lg := range logs {
		printLog(parsed, lg)
	}

	// live
	sink := make(chan types.Log, 64)
	sub, err := client.SubscribeFilterLogs(ctx, q, sink)
	if err != nil {
		log.Fatalf("subscribe: %v", err)
	}
	defer sub.Unsubscribe()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case err := <-sub.Err():
			log.Printf("subscription error: %v (reconnecting in 2s)", err)
			time.Sleep(2000 * time.Millisecond)
			sub, err = client.SubscribeFilterLogs(ctx, q, sink)
			if err != nil {
				log.Fatalf("resubscribe failed: %v", err)
			}
		case lg := <-sink:
			printLog(parsed, lg)
		case <-sigc:
			fmt.Println("\nShutting down watcher.")
			return
		}
	}
}

func printLog(parsed abi.ABI, lg types.Log) {
	modelId := lg.Topics[1].Hex()
	inputHash := lg.Topics[2].Hex()
	prover := common.BytesToAddress(lg.Topics[3].Bytes()).Hex()

	var data struct {
		OutputCommitment [32]byte
		AttestationId    [32]byte
	}
	if err := parsed.UnpackIntoInterface(&data, "InferenceRecorded", lg.Data); err != nil {
		log.Printf("unpack data: %v", err)
		return
	}

	fmt.Printf("\n[announcement] InferenceRecorded\n")
	fmt.Printf("  block        : %d\n", lg.BlockNumber)
	fmt.Printf("  tx           : %s\n", lg.TxHash.Hex())
	fmt.Printf("  modelId      : %s\n", modelId)
	fmt.Printf("  inputHash    : %s\n", inputHash)
	fmt.Printf("  prover       : %s\n", prover)
	fmt.Printf("  outputCommit : 0x%x\n", data.OutputCommitment)
	fmt.Printf("  attestation  : 0x%x\n", data.AttestationId)
}
