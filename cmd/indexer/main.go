package main

import (
	"context"
	"fmt"

	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()
	flowClient, err := client.New("access.mainnet.nodes.onflow.org:9000", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	blockNumber := uint64(1000) // replace with your specific block number

	block, err := flowClient.GetBlockByHeight(ctx, blockNumber)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Block ID: %s\n", block.ID)
	for _, collection := range block.CollectionGuarantees {
		tx, err := flowClient.GetTransaction(ctx, collection.CollectionID)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Transaction ID: %s\n", tx.ID())
	}
}
