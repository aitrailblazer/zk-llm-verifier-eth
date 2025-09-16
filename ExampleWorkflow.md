# Example: Anchoring an AI Run with zk-llm-verifier-eth

1. **Produce an AI output (off-chain)**  
   Run your inference pipeline exactly as usual, then freeze the run. Save the byte-identical input, raw model output, and a human-readable model tag (for example `DeltaSignal-v0.1`). Do not reformat JSON, alter whitespace, or re-encode binaries; future verification depends on these bytes remaining unchanged.

2. **Compute cryptographic commitments (off-chain)**  
   Pass the frozen artifacts to the `dsverifier` CLI. It calculates three keccak256 fingerprints: `modelId` from the pipeline tag, `inputHash` from the exact input bytes, and `outputCommitment` from the exact output bytes. These hashes do not expose the contents but uniquely bind to them.

3. **Submit an attestation (on-chain)**  
   The CLI invokes `InferenceSettlement.verifyAndRecord` with the three commitments. The contract derives an `attestationId` that also captures the prover address and current block, then emits a single `InferenceRecorded` event with `modelId`, `inputHash`, `outputCommitment`, `prover`, and `attestationId`. Using an event keeps gas costs minimal while anchoring an immutable record to Ethereum.

4. **Retrieve the receipt (anytime)**  
   Anyone can pull the transaction receipt or stream contract events to view the `InferenceRecorded` log. The event is a permanent, auditable marker that these commitments were published at or before a specific block by a specific address.

5. **Verify independently (off-chain)**  
   A verifier recomputes keccak256 on the claimed input, output, and model tag. Matching hashes confirm that the published commitments correspond to those artifacts. If all three align, there is cryptographic evidence that the exact output came from the exact input under the declared pipeline tag at the recorded block height and by the recorded prover.

6. **What lives on-chain and what does not**  
   Only commitments and minimal context live on-chain: the three hashes, the prover address, and the derived `attestationId` in the event. Inputs, outputs, prompts, and decoding parameters remain under your control off-chain. This preserves privacy and keeps gas usage low while enabling strong verification.

7. **How it composes with Ethereum**  
   Pairing an `InferenceRecorded` receipt with identity proofs such as ERC-8004 links *who* produced a result with *what* they produced. Smart contracts can require a valid receipt before releasing payment, DAOs can enforce that proposals cite the correct `modelId` and `inputHash`, and marketplaces can filter for agents with consistent, verifiable provenance.

8. **How the guarantees strengthen over time**  
   The same workflow supports upgrades: add commitments for prompts or decoding settings, reference signed pipeline logs proven consistent in zero knowledge, enlist replication committees to co-attest closed-model runs, or integrate zk-inference so open-weight models can prove correctness on-chain.

9. **Practical notes for operation**  
   Emitting an event is inexpensive on L1 and even cheaper on L2; anchor frequent runs on L2 for cost efficiency. Hash-only postings keep sensitive artifacts private. The CLI can run post-inference in CI, as a pipeline sidecar, or as a daemon that watches a folder and anchors new runs automatically.

10. **The core idea in one sentence**  
    Every AI run becomes a signed, timestamped receipt on a neutral ledger by anchoring cryptographic commitments for the model tag, input, and output, shifting trust from screenshots to proofs anyone can verify forever.

---

## Azure AI Foundry Walkthrough (Closed Model)

Here is how to run the same provenance flow with a closed model served from Azure AI Foundry (for example, an Azure OpenAI deployment you label "GPT-5" when it becomes available). The goal stays the same: freeze the run off-chain, derive commitments, then anchor them on Ethereum with `zk-llm-verifier-eth`.

1. **Shape a deterministic Azure run**  
   Choose a single Azure AI Foundry deployment and make the run as reproducible as possible. Lock the deployment name, API version, model settings (temperature, top_p, max_tokens, seed if supported), and the full prompt/payload you send. Save everything needed to replay the call later, including Azure response headers such as `x-ms-request-id`.

   For a chat completion style request:
   ```bash
   export AZURE_OPENAI_ENDPOINT="https://<your-resource>.openai.azure.com"
   export AZURE_OPENAI_API_KEY="<your-azure-key>"
   export AZURE_OPENAI_DEPLOYMENT="gpt5"            # your deployment name
   export AZURE_OPENAI_API_VERSION="2024-XX-XX"     # the exact api-version you used

   mkdir -p runs/azure
   cat > runs/azure/request.json <<'JSON'
   {
     "messages": [
       {"role": "system", "content": "You are a precise financial analyst."},
       {"role": "user", "content": "Summarize liquidity and leverage from this 10-Q text: <...>"}
     ],
     "temperature": 0.0,
     "top_p": 1.0,
     "max_tokens": 800
   }
   JSON

   curl -sS -X POST \
     "$AZURE_OPENAI_ENDPOINT/openai/deployments/$AZURE_OPENAI_DEPLOYMENT/chat/completions?api-version=$AZURE_OPENAI_API_VERSION" \
     -H "Content-Type: application/json" \
     -H "api-key: $AZURE_OPENAI_API_KEY" \
     -d @runs/azure/request.json \
     -D runs/azure/response.headers.txt \
     | tee runs/azure/response.json >/dev/null
   ```

   You now have a frozen input (`request.json`), a raw output (`response.json`), and captured headers. Do not edit these files.

2. **Build a self-contained receipt**  
   Create a single JSON file that bundles everything verifiers need: endpoint identity, deployment, API version, request payload, raw output, and the Azure request id. This file becomes your canonical run receipt.
   ```bash
   jq -n \
     --slurpfile req runs/azure/request.json \
     --slurpfile res runs/azure/response.json \
     --arg endpoint "$AZURE_OPENAI_ENDPOINT" \
     --arg deployment "$AZURE_OPENAI_DEPLOYMENT" \
     --arg apiVersion "$AZURE_OPENAI_API_VERSION" \
     --arg reqId "$(grep -i '^x-ms-request-id:' runs/azure/response.headers.txt | awk '{print $2}' | tr -d '\r')" '
   {
     "provider": "azure-openai",
     "endpoint": $endpoint,
     "deployment": $deployment,
     "api_version": $apiVersion,
     "azure_request_id": $reqId,
     "request": $req[0],
     "response": $res[0]
   }' > runs/azure/receipt.json
   ```

   Optionally persist the original document (for example, the parsed 10-Q text) as a separate artifact if it is not already in `request.json`.

3. **Derive modelId, inputHash, outputCommitment**  
   - `modelId`: keccak256 of a stable tag describing the Azure deployment, such as `azure://<resource>/<deployment>@<api-version>` stored in `model.tag`.
   - `inputHash`: keccak256 of the exact input artifact (for example, `runs/azure/request.json` or your canonical `document.json`).
   - `outputCommitment`: keccak256 of a single file representing the output; hashing the entire `runs/azure/receipt.json` bakes response content and metadata together.

   ```bash
   printf "azure://%s/%s@%s" \
     "$AZURE_OPENAI_ENDPOINT" \
     "$AZURE_OPENAI_DEPLOYMENT" \
     "$AZURE_OPENAI_API_VERSION" \
     > runs/azure/model.tag

   # Optional previews
   cast keccak --from-hex $(xxd -p -c9999 runs/azure/model.tag)
   cast keccak runs/azure/request.json
   cast keccak runs/azure/receipt.json
   ```

4. **Anchor with `dsverifier`**  
   Use the deployed contract and the Azure artifacts:
   ```bash
   export RPC_URL=http://localhost:8545
   export PRIVATE_KEY=<anvil_or_test_key>
   export CONTRACT_ADDR=<deployed_inference_settlement_address>

   go/bin/dsverifier \
     -rpc $RPC_URL \
     -contract $CONTRACT_ADDR \
     -key $PRIVATE_KEY \
     -input runs/azure/request.json \
     -output runs/azure/receipt.json \
     -tag "$(cat runs/azure/model.tag)"
   ```

   The transaction emits `InferenceRecorded(modelId, inputHash, outputCommitment, prover, attestationId)`. Watch live if desired:
   ```bash
   go/bin/watch -rpc $RPC_URL -contract $CONTRACT_ADDR -from 0
   ```

5. **Verify later without Azure access**  
   A third party rehashes `model.tag`, your declared input artifact, and `receipt.json`, then compares to the on-chain event fields. Matching hashes confirm the Azure run produced the committed output from the committed input under the declared deployment and API version.

6. **Azure-specific guidance**  
   Keep temperature at zero when possible, fix `max_tokens` and `top_p`, and capture the exact `api-version` and deployment name. Preserve Azure request ids inside the receipt for enterprise audits. If prompts concatenate documents, store the canonical processed document separately and hash it consistently so `inputHash` remains meaningful. When the provider updates models, bump the tag string (for example, append `#2025-09-14`) and treat that as part of `model.tag`.

7. **When the model remains closed**  
   Even if weights and temperature semantics are private, provenance still blocks silent swaps because mismatched hashes will fail verification. Strengthen trust by folding extra metadata (prompt templates, decoding parameters, Azure request ids) into `receipt.json` and by engaging a replication committee to co-attest identical runs. Matching receipts from independent operators raise confidence without altering the on-chain interface.

You have now run a closed-model Azure inference, frozen the request and response plus provenance metadata in a single receipt file, derived commitments, and anchored them on Ethereum using the standard `zk-llm-verifier-eth` flow.
