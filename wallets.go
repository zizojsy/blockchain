package main

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"os"
)

const walletFile = "wallet_%s.dat"

const centerWalletsData = `217f0301010757616c6c65747301ff80000101010757616c6c65747301ff8400000028ff83040101176d61705b737472696e675d2a6d61696e2e57616c6c657401ff8400010c01ff8200000aff81050102ff860000002eff870301010a507269766174654b657901ff8800010201095075626c69634b657901ff8a0001014401ff8c0000002fff89030101095075626c69634b657901ff8a0001030105437572766501100001015801ff8c0001015901ff8c0000000aff8b050102ff8e000000fe0123ff80010122313837776952334a50366245426577546b6b5463535a4e6643526d7754797465597afff940ff8f0301010b5f507269766174654b657901ff9000010301014401ff8c00010a5075626c69634b65795801ff8c00010a5075626c69634b65795901ff8c0000000aff8b050102ff8e0000006cff90012102dc4994cc9de3281013dbee2081d18298d9096e8fb29bec79d629d0f3a83f8e5f01210228f46a4493c79fe6f0a9475de342464b7cc27eae9e259302d925c0f26daf9aa90121021e07735a8d36ade16c043cd22f1216a91508ecfa38e44a73ec38ba77a394b76e0028f46a4493c79fe6f0a9475de342464b7cc27eae9e259302d925c0f26daf9aa91e07735a8d36ade16c043cd22f1216a91508ecfa38e44a73ec38ba77a394b76e00`

// Wallets stores a collection of wallets
type Wallets struct {
	Wallets map[string]*Wallet
}

func GetCenterWallets() *Wallets {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	err := wallets.LoadFromHex(centerWalletsData)
	if err != nil {
		log.Panic(err)
	}
	return &wallets
}

func GetWallets(nodeID string) (*Wallets, error) {
	if nodeID == centerNodeId {
		return centerWallets, nil
	} else {
		return NewWallets(nodeID)
	}
}

// NewWallets creates Wallets and fills it from a file if it exists
func NewWallets(nodeID string) (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	err := wallets.LoadFromFile(nodeID)

	return &wallets, err
}

// CreateWallet adds a Wallet to Wallets
func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := string(wallet.GetAddress())

	ws.Wallets[address] = wallet

	return address
}

// GetAddresses returns an array of addresses stored in the wallet file
func (ws *Wallets) GetAddresses() []string {
	var addresses []string

	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

// GetWallet returns a Wallet by its address
func (ws Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

func (ws *Wallets) LoadFromHex(s string) error {
	data, err := hex.DecodeString(s)
	if err != nil {
		log.Panic(err)
	}

	var wallets Wallets
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(&wallets)
	if err != nil {
		log.Panic(err)
	}

	ws.Wallets = wallets.Wallets

	return nil
}

// LoadFromFile loads wallets from the file
func (ws *Wallets) LoadFromFile(nodeID string) error {
	walletFile := fmt.Sprintf(walletFile, nodeID)
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	fileContent, err := os.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}

	var wallets Wallets
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {
		log.Panic(err)
	}

	ws.Wallets = wallets.Wallets

	return nil
}

// SaveToFile saves wallets to a file
func (ws Wallets) SaveToFile(nodeID string) {
	var content bytes.Buffer
	walletFile := fmt.Sprintf(walletFile, nodeID)

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}

	err = os.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
