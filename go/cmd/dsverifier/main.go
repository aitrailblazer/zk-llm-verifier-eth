package main

import (
    "flag"
    "fmt"
    "io"
    "os"

    "github.com/ethereum/go-ethereum/common/hexutil"
    "github.com/ethereum/go-ethereum/crypto"
    "golang.org/x/crypto/sha3"

    seth "github.com/aitrailblazer/zk-llm-verifier-eth/go/pkg/eth"
)

func envOr(k, def string) string { if v := os.Getenv(k); v != "" { return v }; return def }

func keccakFile(path string) (string, error) {
    f, err := os.Open(path)
    if err != nil { return "", err }
    defer f.Close()
    h := sha3.NewLegacyKeccak256()
    if _, err := io.Copy(h, f); err != nil { return "", err }
    return "0x" + hexutil.Encode(h.Sum(nil))[2:], nil
}

func modelIDFromTag(tag string) string {
    k := crypto.Keccak256([]byte(tag))
    return "0x" + hexutil.Encode(k)[2:]
}

func mustHex32(s string) [32]byte {
    out, err := seth.Hex32(s)
    if err != nil { panic(err) }
    return out
}

func main() {
    rpc := flag.String("rpc", envOr("RPC_URL", "http://localhost:8545"), "RPC URL")
    contract := flag.String("contract", envOr("CONTRACT_ADDR", ""), "Contract address")
    priv := flag.String("key", envOr("PRIVATE_KEY", ""), "Private key (0x...)")
    input := flag.String("input", "", "Input file")
    output := flag.String("output", "", "Output file")
    tag := flag.String("tag", envOr("PIPELINE_TAG", "DeltaSignal-v0.1"), "Pipeline/model tag")
    flag.Parse()

    if *contract == "" || *priv == "" || *input == "" || *output == "" {
        fmt.Println("Usage: dsverifier -rpc <RPC> -contract 0x... -key 0x... -input data.json -output out.txt -tag DeltaSignal-v0.1")
        os.Exit(2)
    }

    inHex, err := keccakFile(*input); if err != nil { panic(err) }
    outHex, err := keccakFile(*output); if err != nil { panic(err) }
    modelHex := modelIDFromTag(*tag)

    fmt.Println("RPC              :", *rpc)
    fmt.Println("Contract         :", *contract)
    fmt.Println("Pipeline Tag     :", *tag)
    fmt.Println("Model ID         :", modelHex)
    fmt.Println("Input Hash       :", inHex)
    fmt.Println("Output Commitment:", outHex)

    client, err := seth.NewSettlementClient(*rpc, *contract); if err != nil { panic(err) }
    txHash, err := client.SendAttestation(*priv, mustHex32(modelHex), mustHex32(inHex), mustHex32(outHex))
    if err != nil { panic(err) }
    fmt.Println("TX Hash          :", txHash.Hex())
}
