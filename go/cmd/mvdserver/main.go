package main

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/sha3"
)

type PayReq struct {
	Scheme, Network, MaxAmountRequired, Resource, Description, MimeType, PayTo, Asset string
	MaxTimeoutSeconds                                                                 int
	Extra                                                                             interface{} `json:"extra,omitempty"`
}
type Pay402 struct {
	X402Version int      `json:"x402Version"`
	Accepts     []PayReq `json:"accepts"`
	Error       string   `json:"error,omitempty"`
}

func main() {
	// Receiving address is now optional:
	// If not set, defaults to the all-zeros dummy address.
	recv := getEnv("RECEIVING_ADDRESS", "0x0000000000000000000000000000000000000000")
	netw := getEnv("NETWORK", "base-sepolia")
	price := getEnv("MAX_AMOUNT", "100000") // atomic units; 100000 = 0.10 USDC
	asset := getEnv("ASSET_ADDRESS", "USDC")

	mux := http.NewServeMux()

	mux.HandleFunc("/insight/demo", func(w http.ResponseWriter, r *http.Request) {
		payHeader := r.Header.Get("X-PAYMENT")

		if payHeader == "" {
			req := PayReq{
				Scheme: "exact", Network: netw, MaxAmountRequired: price,
				Resource: r.URL.Path, Description: "DeltaSignal Demo Insight Card",
				MimeType: "application/json", PayTo: recv, MaxTimeoutSeconds: 120,
				Asset: asset, Extra: map[string]string{"name": "USDC", "version": "EIP-3009"},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPaymentRequired)
			json.NewEncoder(w).Encode(Pay402{X402Version: 1, Accepts: []PayReq{req}})
			return
		}

		paymentTx := fakeTx()
		inputHash := keccak([]byte(r.URL.Query().Encode()))
		card := map[string]any{"report_id": "demo-001", "company": "DemoCo", "signal": "Revenue restatement risk", "risk_tier": "Critical"}
		outBytes, _ := json.Marshal(card)
		outputHash := keccak(outBytes)
		modelID := keccak([]byte("DeltaSignal-v0.1"))
		attestTx := fakeTx()

		w.Header().Set("X-PAYMENT-RESPONSE", base64.StdEncoding.EncodeToString([]byte(paymentTx)))
		w.Header().Set("Content-Type", "application/json")
		card["paid_via"] = "x402"
		card["payment_tx"] = paymentTx
		card["payment_network"] = netw
		card["attest_tx"] = attestTx
		card["model_id"] = modelID
		card["input_hash"] = inputHash
		card["output_commitment"] = outputHash
		card["generated_at"] = time.Now().UTC().Format(time.RFC3339)
		json.NewEncoder(w).Encode(card)
	})

	log.Println("MVD x402 server on :4021 (mock-only)")
	log.Fatal(http.ListenAndServe(":4021", mux))
}

func keccak(b []byte) string {
	h := sha3.NewLegacyKeccak256()
	h.Write(b)
	return "0x" + hex.EncodeToString(h.Sum(nil))
}
func fakeTx() string {
	const x = "0123456789abcdef"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 64)
	for i := range b {
		b[i] = x[rand.Intn(16)]
	}
	return "0x" + string(b)
}
func getEnv(k, d string) string {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	return v
}
