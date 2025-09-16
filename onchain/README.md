# onchain — zk-llm-verifier-eth

This directory contains the Foundry project for the `InferenceSettlement` contract. The contract exposes `verifyAndRecord(modelId, inputHash, outputCommitment)` and emits a tamper-evident `InferenceRecorded` event used throughout the provenance pipeline.

## Layout

- `src/InferenceSettlement.sol` – Solidity contract that hashes the attestation tuple and emits events.
- `script/Deploy.s.sol` – Forge script used by `make onchain-deploy` to broadcast deployments.
- `foundry.toml` – Shared compiler/config settings for the project.
- `broadcast/` & `cache/` – Auto-generated when you run scripts; hold transaction logs and secrets.

## Common Commands

From this directory or via the repository Makefile:

```bash
# compile the contract
forge build

# run the Foundry test suite (add tests under onchain/test)
forge test

# format Solidity sources
forge fmt

# deploy using environment variables
forge script script/Deploy.s.sol \
  --rpc-url ${RPC_URL:-http://localhost:8545} \
  --private-key ${PRIVATE_KEY:?set PRIVATE_KEY} \
  --broadcast
```

For convenience you can also run the top-level targets:

```bash
make onchain-build   # compiles via forge build
make onchain-deploy  # runs the deploy script with your RPC/private key
```

## Notes

- The deploy script prints the new contract address and stores a broadcast record under `broadcast/Deploy.s.sol/`.
- `forge snapshot`, `forge coverage`, and other Foundry utilities are available if you need gas reports or coverage metrics.
- Full Foundry documentation: https://book.getfoundry.sh/

Feel free to add tests in `onchain/test/` to cover new contract logic as the project evolves.
