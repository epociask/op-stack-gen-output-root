package main

import (
	"context"
	"flag"
	"log"
	"math/big"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/ethereum/go-ethereum/ethclient/gethclient"
)

// blockInfo ... Wrapper for a block
// This is used to ensure compatibility with the rollup-node software
type blockInfo struct {
	*types.Block
}

// HeaderRLP ... Returns the RLP encoded header of a block
func (b blockInfo) HeaderRLP() ([]byte, error) {
	return rlp.EncodeToBytes(b.Header())
}

// blockToInfo ... Converts a block to a blockInfo
func blockToInfo(b *types.Block) blockInfo {
	return blockInfo{b}
}
func NewGethClient(rawURL string) (*gethclient.Client, error) {
	rpcClient, err := rpc.Dial(rawURL)
	if err != nil {
		return nil, err
	}

	gethClient := gethclient.New(rpcClient)
	return gethClient, nil
}
func main() {
	log.Printf("Ingesting CMD args")

	// Read & parse command line arguments
	l2RPC := flag.String("l2-rpc", "", "L2 Op-Geth RPC endpoint")
	l2BlockNum := flag.Int64("l2-block-num", 0, "L2 block number")

	l2tol1MessagePasserAddr := common.HexToAddress("0x4200000000000000000000000000000000000016")

	flag.Parse()

	log.Printf("Constructing L2 ETH clients (ethclient,gethclient)")
	// Construct L2 eth clients
	l2Client, err := ethclient.Dial(*l2RPC)
	if err != nil {
		panic(err)
	}

	l2Geth, err := NewGethClient(*l2RPC)
	if err != nil {
		panic(err)
	}

	bigInt := big.NewInt(*l2BlockNum)

	log.Printf("Fetching L2 block by number")
	outputBlock, err := l2Client.BlockByNumber(context.Background(), bigInt)
	if err != nil {
		panic(err)
	}

	log.Printf("Fetching withdrawal root for L2 block number")
	proofResp, err := l2Geth.GetProof(context.Background(),
		l2tol1MessagePasserAddr, []string{}, bigInt)
	if err != nil {
		panic(err)
	}

	log.Printf("Computing L2 state root")
	asInfo := blockToInfo(outputBlock)
	stateRoot, err := rollup.ComputeL2OutputRootV0(asInfo, proofResp.StorageHash)
	if err != nil {
		panic(err)
	}

	log.Printf("Successfully computed state root: %s", stateRoot)

}
