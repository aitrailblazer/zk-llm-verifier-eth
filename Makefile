.PHONY: help onchain-build localnet onchain-deploy ds-build ds-attest watch-build watch

GOCACHE_DIR := $(CURDIR)/.gocache
GOMODCACHE_DIR := $(CURDIR)/.gomodcache

help:
	@echo "Available targets:"
	@echo "  make onchain-build    - Compile Solidity contracts with Foundry"
	@echo "  make localnet         - Launch Anvil on port 8545"
	@echo "  make onchain-deploy   - Deploy InferenceSettlement (requires RPC_URL, PRIVATE_KEY)"
	@echo "  make ds-build         - Tidy Go module and build dsverifier binary"
	@echo "  make ds-attest ...    - Hash artifacts and submit attestation"
	@echo "  make watch ...        - Build watcher and stream InferenceRecorded events"
	@echo ""
	@echo "Quick step-by-step:"
	@echo "  1) make onchain-build"
	@echo "  2) make ds-build"
	@echo "  3) make localnet (new terminal)"
	@echo "  4) make onchain-deploy"
	@echo "  5) make ds-attest INPUT=... OUTPUT=... CONTRACT=..."
	@echo "     (optional) make watch CONTRACT=..."
	@echo "\nTip: run 'make <target>' for the task you need."

DEFAULT_TARGET := help
.DEFAULT_GOAL := $(DEFAULT_TARGET)

onchain-build:
	cd onchain && forge build

localnet:
	anvil -p 8545

onchain-deploy:
	cd onchain && forge script script/Deploy.s.sol \
	  --rpc-url $${RPC_URL:-http://localhost:8545} \
	  --private-key $${PRIVATE_KEY:?set PRIVATE_KEY} \
	  --broadcast

ds-build:
	mkdir -p "$(GOCACHE_DIR)" "$(GOMODCACHE_DIR)" go/bin
	cd go && GOCACHE="$(GOCACHE_DIR)" GOMODCACHE="$(GOMODCACHE_DIR)" go mod tidy
	cd go && GOCACHE="$(GOCACHE_DIR)" GOMODCACHE="$(GOMODCACHE_DIR)" go build -o bin/dsverifier ./cmd/dsverifier

# Usage:
# make ds-attest INPUT=examples/input.json OUTPUT=examples/output.txt CONTRACT=0x... TAG=DeltaSignal-v0.1
ds-attest:
	@if [ -z "$(INPUT)" ]; then echo "INPUT not set (e.g., INPUT=examples/input.json)"; exit 2; fi
	@if [ -z "$(OUTPUT)" ]; then echo "OUTPUT not set (e.g., OUTPUT=examples/output.txt)"; exit 2; fi
	INPUT_PATH="$(abspath $(INPUT))"; \
	OUTPUT_PATH="$(abspath $(OUTPUT))"; \
	cd go && ./bin/dsverifier \
	  -rpc $${RPC_URL:-http://localhost:8545} \
	  -contract $${CONTRACT:?set CONTRACT} \
	  -key $${PRIVATE_KEY:?set PRIVATE_KEY} \
	  -input "$$INPUT_PATH" \
	  -output "$$OUTPUT_PATH" \
	  -tag $${TAG:-DeltaSignal-v0.1}

watch-build:
	mkdir -p "$(GOCACHE_DIR)" "$(GOMODCACHE_DIR)" go/bin
	cd go && GOCACHE="$(GOCACHE_DIR)" GOMODCACHE="$(GOMODCACHE_DIR)" go build -o bin/watch ./cmd/watch

# Usage: make watch CONTRACT=0x...
watch: watch-build
	cd go && ./bin/watch \
	  -rpc $${RPC_URL:-http://localhost:8545} \
	  -contract $${CONTRACT:?set CONTRACT} \
	  -from 0
