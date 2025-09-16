# Specification: zk-llm-verifier-eth

## Abstract

`zk-llm-verifier-eth` provides a minimal mechanism for **AI inference provenance on Ethereum**.  
It defines an `InferenceSettlement` contract and a companion Go client to anchor cryptographic commitments (`modelId`, `inputHash`, `outputCommitment`, `prover`) on-chain.  
Each inference results in an immutable `InferenceRecorded` event, enabling downstream agents, auditors, and counterparties to verify that outputs match declared inputs and models.  
This MVP establishes a trust-minimized foundation today, with a clear path toward zk-log consistency and zk-inference proofs.

---

## Motivation

Modern AI systems are opaque. Providers can:  
- silently swap models,  
- alter hyperparameters,  
- fabricate or filter outputs,  
- misrepresent which system produced a given result.  

Without cryptographic guarantees, users and institutions must **trust unverifiable provider claims**, creating systemic risks:  
- **Enterprise & Regulatory** — Boards and auditors cannot prove AI-driven reports match disclosed filings.  
- **Agent-to-Agent Coordination** — AI agents lack verifiable reputations; interactions are unverifiable.  
- **Research & Science** — Critical forecasts (climate, medical, legal) can be altered without traceability.  

Ethereum, as a neutral and censorship-resistant settlement layer, can anchor provenance records that persist beyond any single provider or platform.

---

## Specification

### On-Chain Contract — `InferenceSettlement`

- Function:
```solidity
function verifyAndRecord(
    bytes32 modelId,
    bytes32 inputHash,
    bytes32 outputCommitment
) external returns (bytes32 attestationId);
```

- Event:
```solidity
event InferenceRecorded(
    bytes32 indexed modelId,
    bytes32 indexed inputHash,
    bytes32 outputCommitment,
    address indexed prover,
    bytes32 attestationId
);
```

- `attestationId = keccak256(modelId, inputHash, outputCommitment, prover, block.number)`

| Field              | Purpose                                             |
|--------------------|-----------------------------------------------------|
| `modelId`          | Commitment to pipeline/model version (e.g. tag hash) |
| `inputHash`        | keccak256 hash of exact input payload                |
| `outputCommitment` | keccak256 hash of model output                       |
| `prover`           | Ethereum address submitting attestation              |
| `attestationId`    | Unique ID binding all fields + block metadata        |

### Off-Chain Client — Go CLI (`dsverifier`)

- Computes keccak256 hashes of input/output artifacts.  
- Derives `modelId` from a pipeline tag (e.g. `DeltaSignal-v0.1`).  
- Submits the `verifyAndRecord` transaction.  
- Returns tx hash and provenance metadata.

### Workflow

1. AI pipeline produces an output.  
2. CLI hashes input + output, derives `modelId`.  
3. Contract logs the attestation as `InferenceRecorded`.  
4. Anyone can replay, compare, or audit commitments.

---

## Rationale

- **Minimalism** — hash-based commitments are lightweight and usable today.  
- **Composability** — complements ERC-8004:  
  - ERC-8004 proves *who* an agent is.  
  - zk-llm-verifier-eth proves *what* they did.  
- **Extensibility** — upgrade path:  
  - zk-log consistency → bind commitments to signed pipeline runs.  
  - zk-inference → prove `y = f(x; W)` directly for open models.  
  - Replication committees → multi-signer attestations for closed models.

---

## Security Considerations

- **Hash Collisions** — mitigated by keccak256.  
- **Dishonest Provers** — may misrepresent models; mitigated by zk-log proofs and committees.  
- **Replay Attacks** — same attestation can be resubmitted; clients should check duplication.  
- **Censorship Resistance** — Ethereum ensures immutability of attestations.  
- **Scalability** — commitments are fixed-size (`bytes32`), minimizing gas use.

---

## Reference Implementation

- [`onchain/src/InferenceSettlement.sol`](onchain/src/InferenceSettlement.sol) — Solidity contract  
- [`go/cmd/dsverifier/main.go`](go/cmd/dsverifier/main.go) — Go CLI client  
- Tested with Anvil, deployable on Sepolia / Base Sepolia testnets.

---

## Example Use Case: DeltaSignal

DeltaSignal analyzes SEC filings and generates foresight signals.  
With zk-llm-verifier-eth:  
- Boards/investors can verify outputs correspond exactly to disclosed filings.  
- Tampering or model swaps become detectable.  
- Compliance workflows gain cryptographic assurance.

---

## Roadmap

- **MVP (current)** — Hash commitments, on-chain provenance.  
- **zk-log consistency (future)** — proofs binding commitments to signed pipeline runs.  
- **Parameter commitments (future)** — prompts, decoding settings, API receipts.  
- **Replication committees (future)** — multi-signer attestations for closed models.  
- **zk-inference (future)** — full inference proofs (Noir, RiscZero, ezkl).

---

## References

- [ERC-8004: AI Agent Identity Standard](https://eips.ethereum.org/)  
- [Ethereum Foundation dAI Team announcement](https://blog.ethereum.org/)  
- DeltaSignal: AI foresight pipeline for SEC filings (prototype)
