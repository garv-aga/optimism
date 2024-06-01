package derive

import (
    "bytes"
    "encoding/binary"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"

    "github.com/ethereum-optimism/optimism/op-node/eth"
)

const (
    InfrasoundFuncSignature = "report(uint64,uint64)"
    InfrasoundArguments     = 2
    InfrasoundLen           = 4 + 32*InfrasoundArguments
)

var (
    InfrasoundFuncBytes4 = crypto.Keccak256([]byte(InfrasoundFuncSignature))[:4]
    InfrasoundAddress    = common.HexToAddress("0xB565656a18bd287A28C6aBEBBcbBDB54DDe37eeb")
)

type InfrasoundInfo struct {
    Timestamp uint64
    Burn   uint64
}

func (info *InfrasoundInfo) MarshalBinary() ([]byte, error) {
    data := make([]byte, InfrasoundLen)
    offset := 0
    copy(data[offset:4], InfrasoundFuncBytes4)
    offset += 4
    binary.BigEndian.PutUint64(data[offset+24:offset+32], info.Timestamp)
    offset += 32
    binary.BigEndian.PutUint64(data[offset+24:offset+32], info.Burn)
    return data, nil
}

func (info *InfrasoundInfo) UnmarshalBinary(data []byte) error {
    if len(data) != L1InfoLen {
        return fmt.Errorf("data is unexpected length: %d", len(data))
    }
    var padding [24]byte
    offset := 4
    info.Timestamp = binary.BigEndian.Uint64(data[offset+24 : offset+32])
    if !bytes.Equal(data[offset:offset+24], padding[:]) {
        return fmt.Errorf("l1 burn tx timestamp exceeds uint64 bounds: %x", data[offset:offset+32])
    }
    offset += 32
    info.Burn = binary.BigEndian.Uint64(data[offset+24 : offset+32])
    if !bytes.Equal(data[offset:offset+24], padding[:]) {
        return fmt.Errorf("l1 burn tx burn exceeds uint64 bounds: %x", data[offset:offset+32])
    }
    return nil
}

func InfrasoundDepositTxData(data []byte) (InfrasoundInfo, error) {
    var info InfrasoundInfo
    err := info.UnmarshalBinary(data)
    return info, err
}

func InfrasoundDeposit(seqNumber uint64, block eth.BlockInfo, sysCfg eth.SystemConfig) (*types.DepositTx, error) {
    infoDat := InfrasoundInfo{
        Timestamp: block.Time(),
        Burn:   block.BaseFee().Uint64() * block.GasUsed() + block.BlobBaseFee(),
    }
    data, err := infoDat.MarshalBinary()
    if err != nil {
        return nil, err
    }
    source := L1InfoDepositSource{
        L1BlockHash: block.Hash(),
        SeqNumber:   seqNumber,
    }
    return &types.DepositTx{
        SourceHash:          source.SourceHash(),
        From:                L1InfoDepositerAddress,
        To:                  &InfrasoundAddress,
        Mint:                nil,
        Value:               big.NewInt(0),
        Gas:                 150_000_000,
        IsSystemTransaction: true,
        Data:                data,
    }, nil
}

func InfrasoundDepositBytes(seqNumber uint64, l1Info eth.BlockInfo, sysCfg eth.SystemConfig) ([]byte, error) {
    dep, err := InfrasoundDeposit(seqNumber, l1Info, sysCfg)
    if err != nil {
        return nil, fmt.Errorf("failed to create L1 burn tx: %w", err)
    }
    l1Tx := types.NewTx(dep)
    opaqueL1Tx, err := l1Tx.MarshalBinary()
    if err != nil {
        return nil, fmt.Errorf("failed to encode L1 burn tx: %w", err)
    }
    return opaqueL1Tx, nil
}