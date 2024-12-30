package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"

	"os"
)

// CLI responsible for processing command line arguments
type CLI struct{}

const indent = "  "

func (cli *CLI) createPrompt(cmd string, args []string, explains []string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%[1]s%[2]s\r\n", indent, cmd))

	for idx, arg := range args {
		sb.WriteString(fmt.Sprintf("%[1]s%[1]s%[2]s\r\n", indent, arg))
		sb.WriteString(fmt.Sprintf("%[1]s%[1]s%[1]s%[2]s\r\n", indent, explains[idx]))
	}

	return sb.String()
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println(cli.createPrompt("wallet",
		[]string{"-c",
			"-l",
			"-T -f A -t B -a AMOUNT [-m]"},
		[]string{"Create a new account in wallet",
			"List all accounts in wallet",
			"Transfer AMOUNT money from A to B, mine coin if -m flag is set"}))
	fmt.Println(cli.createPrompt("service",
		[]string{"-s [-m ADDRESS]",
			"-p",
			"-b ADDRESS"},
		[]string{"Start Servece, mine coin if ADDRESS is given",
			"Print all blocks in the blockchain",
			"Get balance of ADDRESS"}))
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

// Run parses command line arguments and processes commands
func (cli *CLI) Run() {
	cli.validateArgs()

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		fmt.Printf("NODE_ID env. var is not set!")
		os.Exit(1)
	}
	CreateGenesisIfNeeded(nodeID)

	walletCmd := flag.NewFlagSet("wallet", flag.ExitOnError)
	serviceCmd := flag.NewFlagSet("service", flag.ExitOnError)

	createWalletFlag := walletCmd.Bool("c", false, "Create a new account in wallet")
	listWalletFlag := walletCmd.Bool("l", false, "List all accounts in wallet")
	transferFlag := walletCmd.Bool("T", false, "Transfer AMOUNT money from A to B, mine coin if -m flag is set")

	startFlag := serviceCmd.Bool("s", false, "Start Servece, mine coin if ADDRESS is given")
	printFlag := serviceCmd.Bool("p", false, "Print all blocks in the blockchain")
	balanceFlag := serviceCmd.Bool("b", false, "Get balance of ADDRESS")

	fromAddr := walletCmd.String("f", "", "Source wallet address")
	toAddr := walletCmd.String("t", "", "Destination wallet address")
	transferAmount := walletCmd.Int("a", 0, "Amount to trainsfer")
	transferMine := walletCmd.Bool("m", false, "Mine immediately on the same node")
	balanceAddr := serviceCmd.String("a", "", "The address to get balance for")
	mineAddr := serviceCmd.String("m", "", "Enable mining mode and send reward to ADDRESS")

	switch os.Args[1] {
	case "wallet":
		err := walletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "service":
		err := serviceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if walletCmd.Parsed() {
		if *createWalletFlag {
			cli.createWallet(nodeID)
		}

		if *listWalletFlag {
			cli.listAddresses(nodeID)
		}

		if *transferFlag {
			if *fromAddr == "" || *toAddr == "" || *transferAmount <= 0 {
				walletCmd.Usage()
				os.Exit(1)
			}

			cli.send(*fromAddr, *toAddr, *transferAmount, nodeID, *transferMine)
		}
	}

	if serviceCmd.Parsed() {
		if *startFlag {
			cli.startNode(nodeID, *mineAddr)
		}

		if *printFlag {
			cli.printChain(nodeID)
		}

		if *balanceFlag {
			cli.getBalance(*balanceAddr, nodeID)
		}
	}
}

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

func (cli *CLI) getBalance(address, nodeID string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := NewBlockchain(nodeID)
	UTXOSet := UTXOSet{bc}
	defer bc.DB.Close()

	balance := 0
	pubKeyHash := Base58Decode([]byte(address))

	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) listAddresses(nodeID string) {

	wallets, err := GetWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.GetAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CLI) printChain(nodeID string) {
	bc := NewBlockchain(nodeID)
	defer bc.DB.Close()

	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("============ Block %x ============\n", block.Hash)
		fmt.Printf("Height: %d\n", block.Height)
		fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Printf("\n\n")

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) send(from, to string, amount int, nodeID string, mineNow bool) {
	if !ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := NewBlockchain(nodeID)
	UTXOSet := UTXOSet{bc}
	defer bc.DB.Close()

	wallets, err := GetWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)

	tx := NewUTXOTransaction(&wallet, to, amount, &UTXOSet)

	if mineNow {
		cbTx := NewCoinbaseTX(from, "")
		txs := []*Transaction{cbTx, tx}

		newBlock := bc.MineBlock(txs)
		UTXOSet.Update(newBlock)
	} else {
		sendTx(knownNodes[0], tx)
	}

	fmt.Println("Success!")
}

func (cli *CLI) startNode(nodeID, minerAddress string) {
	fmt.Printf("Starting node %s\n", nodeID)
	if len(minerAddress) > 0 {
		if ValidateAddress(minerAddress) {
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
		} else {
			log.Panic("Wrong miner address!")
		}
	}
	StartServer(nodeID, minerAddress)
}
