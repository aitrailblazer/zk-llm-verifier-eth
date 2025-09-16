package eth

import (
    "context"
    "crypto/ecdsa"
    "encoding/hex"
    "errors"
    "math/big"
    "strings"

    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
)

const settlementABI = `[{"inputs":[{"internalType":"bytes32","name":"modelId","type":"bytes32"},{"internalType":"bytes32","name":"inputHash","type":"bytes32"},{"internalType":"bytes32","name":"outputCommitment","type":"bytes32"}],"name":"verifyAndRecord","outputs":[{"internalType":"bytes32","name":"attestationId","type":"bytes32"}],"stateMutability":"nonpayable","type":"function"}]`

type SettlementClient struct {
    Client       *ethclient.Client
    ContractAddr common.Address
    ABI          abi.ABI
    ChainID      *big.Int
}

func NewSettlementClient(rpcURL, contractAddr string) (*SettlementClient, error) {
    c, err := ethclient.Dial(rpcURL)
    if err != nil {
        return nil, err
    }
    parsed, err := abi.JSON(strings.NewReader(settlementABI))
    if err != nil {
        return nil, err
    }
    chainID, err := c.ChainID(context.Background())
    if err != nil {
        return nil, err
    }
    return &SettlementClient{Client: c, ContractAddr: common.HexToAddress(contractAddr), ABI: parsed, ChainID: chainID}, nil
}

func (s *SettlementClient) SendAttestation(privKeyHex string, modelId, inputHash, outputCommitment [32]byte) (common.Hash, error) {
    privKey, err := crypto.HexToECDSA(strings.TrimPrefix(privKeyHex, "0x"))
    if err != nil { return common.Hash{}, err }
    pubKey, ok := privKey.Public().(*ecdsa.PublicKey)
    if !ok {
        return common.Hash{}, errors.New("invalid public key type")
    }
    fromAddr := crypto.PubkeyToAddress(*pubKey)

    nonce, err := s.Client.PendingNonceAt(context.Background(), fromAddr)
    if err != nil { return common.Hash{}, err }
    gasPrice, err := s.Client.SuggestGasPrice(context.Background())
    if err != nil { return common.Hash{}, err }

    calldata, err := s.ABI.Pack("verifyAndRecord", modelId, inputHash, outputCommitment)
    if err != nil { return common.Hash{}, err }

    tx := types.NewTx(&types.LegacyTx{
        Nonce:    nonce,
        To:       &s.ContractAddr,
        Value:    big.NewInt(0),
        Gas:      300000,
        GasPrice: gasPrice,
        Data:     calldata,
    })

    signed, err := types.SignTx(tx, types.LatestSignerForChainID(s.ChainID), privKey)
    if err != nil { return common.Hash{}, err }

    if err := s.Client.SendTransaction(context.Background(), signed); err != nil {
        return common.Hash{}, err
    }
    return signed.Hash(), nil
}

func Hex32(s string) ([32]byte, error) {
    var out [32]byte
    s = strings.TrimPrefix(s, "0x")
    b, err := hex.DecodeString(s)
    if err != nil { return out, err }
    if len(b) != 32 { return out, errors.New("expected 32 bytes") }
    copy(out[:], b)
    return out, nil
}
