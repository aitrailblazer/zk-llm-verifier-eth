# Quickstart: zk-llm-verifier-eth

Run the local provenance demo after cloning the repository.

---

## 0. Clone & Install Prereqs

```bash
git clone https://github.com/aitrailblazer/zk-llm-verifier-eth.git
cd zk-llm-verifier-eth

# Install Foundry toolchain
curl -L https://foundry.paradigm.xyz | bash
foundryup

# Install/verify Go 1.24.3 (adjust for your OS)
go version || brew install go@1.24
```

---

## 1. Start Local Ethereum Node

```bash
make localnet
```

Take note of one Anvil private key for later.

---

## 2. Deploy `InferenceSettlement`

```bash
export RPC_URL=http://localhost:8545
export PRIVATE_KEY=<anvil_private_key>
make onchain-deploy
export CONTRACT_ADDR=<deployed_address_from_output>
```

---

## 3. Build the Go CLI

```bash
make ds-build
```

---

## 4. Prepare Sample Artifacts

```bash
mkdir -p examples
echo '{"example":"delta"}' > examples/input.json
echo 'Risk summary text (AI output)...' > examples/output.txt
```

---

## 5. Submit an Attestation

```bash
make ds-attest \
  INPUT=examples/input.json \
  OUTPUT=examples/output.txt \
  CONTRACT=$CONTRACT_ADDR \
  TAG=DeltaSignal-v0.1
```

Result: prints model/input/output hashes and the attestation transaction hash. Inspect the `InferenceRecorded` event via Anvil logs or `cast receipt $TX_HASH`.

---

Next: read [`README.md`](README.md) for architecture details or explore the Solidity/Go sources under `onchain/` and `go/`.
