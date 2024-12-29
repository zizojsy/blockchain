package main

import (
	"fmt"
)

func (cli *CLI) createBlockchain(nodeID string) {
	bc := CreateBlockchain(nodeID)
	defer bc.DB.Close()

	UTXOSet := UTXOSet{bc}
	UTXOSet.Reindex()

	fmt.Println("Done!")
}
