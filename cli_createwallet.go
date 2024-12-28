package main

import (
	"fmt"
	"os"
)

func (cli *CLI) createWallet(nodeID string) {
	if nodeID == centerNodeId {
		fmt.Println("Center Node NOT allowed to create wallet")
		os.Exit(1)
	}
	wallets, _ := NewWallets(nodeID)
	address := wallets.CreateWallet()
	wallets.SaveToFile(nodeID)

	fmt.Printf("Your new address: %s\n", address)
}
