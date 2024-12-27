package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/nutsdb/nutsdb"
)

const dbFile = "db/dblockchain_%s"
const blocksBucket = "blocks"
const genesisCoinbaseData = "Create block chain mannually according to Fuda MSE Project"
const genesisAddress = "1sVQW7fv6Gx6EqS1fvzzq9w56zFuhpJnK"

// Blockchain implements interactions with a DB
type Blockchain struct {
	tip []byte
	db  *nutsdb.DB
}

func GetDbName(nodeID string) string {
	return fmt.Sprintf(dbFile, nodeID)
}

func CreateGenesisIfNeeded(nodeID string) {
	if !dbExists(GetDbName(nodeID)) {
		bc := CreateBlockchain(nodeID)
		defer bc.db.Close()

		UTXOSet := UTXOSet{bc}
		UTXOSet.Reindex()

		fmt.Println("Done!")
	}
}

// CreateBlockchain creates a new blockchain DB
func CreateBlockchain(nodeID string) *Blockchain {
	dbFile := GetDbName(nodeID)
	if dbExists(dbFile) {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte

	cbtx := NewCoinbaseTX(genesisAddress, genesisCoinbaseData)
	genesis := NewGenesisBlock(cbtx)

	db, err := nutsdb.Open(nutsdb.DefaultOptions, nutsdb.WithDir(dbFile))
	if err != nil {
		log.Panic(err)
	}

	if err := db.Update(func(tx *nutsdb.Tx) error {
		// you should call Bucket with data structure and the name of bucket first
		return tx.NewBucket(nutsdb.DataStructureBTree, blocksBucket)
	}); err != nil {
		log.Fatal(err)
	}

	err = db.Update(func(tx *nutsdb.Tx) error {
		// err := tx.NewBucket(nutsdb.DataStructureBTree, blocksBucket)
		if err != nil {
			log.Panic(err)
		}

		err = tx.Put(blocksBucket, genesis.Hash, genesis.Serialize(), 0)
		if err != nil {
			log.Panic(err)
		}

		err = tx.Put(blocksBucket, []byte("l"), genesis.Hash, 0)
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

// NewBlockchain creates a new Blockchain with genesis Block
func NewBlockchain(nodeID string) *Blockchain {
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if !dbExists(dbFile) {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := nutsdb.Open(nutsdb.DefaultOptions, nutsdb.WithDir(dbFile))
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *nutsdb.Tx) error {
		tip, err = tx.Get(blocksBucket, []byte("l"))
		if err != nil {
			log.Panic(err)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

// AddBlock saves the block into the blockchain
func (bc *Blockchain) AddBlock(block *Block) {
	err := bc.db.Update(func(tx *nutsdb.Tx) error {
		blockInDb, _ := tx.Get(blocksBucket, block.Hash)

		if blockInDb != nil {
			return nil
		}

		blockData := block.Serialize()
		err := tx.Put(blocksBucket, block.Hash, blockData, 0)
		if err != nil {
			log.Panic(err)
		}

		lastHash, _ := tx.Get(blocksBucket, []byte("l"))
		lastBlockData, _ := tx.Get(blocksBucket, lastHash)
		lastBlock := DeserializeBlock(lastBlockData)

		if block.Height > lastBlock.Height {
			err = tx.Put(blocksBucket, []byte("l"), block.Hash, 0)
			if err != nil {
				log.Panic(err)
			}
			bc.tip = block.Hash
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

// FindTransaction finds a transaction by its ID
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, ID) {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction is not found")
}

// FindUTXO finds all unspent transaction outputs and returns transactions with spent outputs removed
func (bc *Blockchain) FindUTXO() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs)
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// Was the output spent?
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return UTXO
}

// Iterator returns a BlockchainIterat
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}

	return bci
}

// GetBestHeight returns the height of the latest block
func (bc *Blockchain) GetBestHeight() int {
	var lastBlock Block

	err := bc.db.View(func(tx *nutsdb.Tx) error {
		lastHash, _ := tx.Get(blocksBucket, []byte("l"))
		blockData, _ := tx.Get(blocksBucket, lastHash)
		lastBlock = *DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return lastBlock.Height
}

// GetBlock finds a block by its hash and returns it
func (bc *Blockchain) GetBlock(blockHash []byte) (Block, error) {
	var block Block

	err := bc.db.View(func(tx *nutsdb.Tx) error {
		blockData, _ := tx.Get(blocksBucket, blockHash)

		if blockData == nil {
			return errors.New("Block is not found")
		}

		block = *DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		return block, err
	}

	return block, nil
}

// GetBlockHashes returns a list of hashes of all the blocks in the chain
func (bc *Blockchain) GetBlockHashes() [][]byte {
	var blocks [][]byte
	bci := bc.Iterator()

	for {
		block := bci.Next()

		blocks = append(blocks, block.Hash)

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return blocks
}

// MineBlock mines a new block with the provided transactions
func (bc *Blockchain) MineBlock(transactions []*Transaction) *Block {
	var lastHash []byte
	var lastHeight int

	for _, tx := range transactions {
		// TODO: ignore transaction if it's not valid
		if !bc.VerifyTransaction(tx) {
			log.Panic("ERROR: Invalid transaction")
		}
	}

	err := bc.db.View(func(tx *nutsdb.Tx) error {
		lastHash, _ = tx.Get(blocksBucket, []byte("l"))

		blockData, _ := tx.Get(blocksBucket, lastHash)
		block := DeserializeBlock(blockData)

		lastHeight = block.Height

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, lastHash, lastHeight+1)

	err = bc.db.Update(func(tx *nutsdb.Tx) error {
		err := tx.Put(blocksBucket, newBlock.Hash, newBlock.Serialize(), 0)
		if err != nil {
			log.Panic(err)
		}

		err = tx.Put(blocksBucket, []byte("l"), newBlock.Hash, 0)
		if err != nil {
			log.Panic(err)
		}

		bc.tip = newBlock.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return newBlock
}

// SignTransaction signs inputs of a Transaction
func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

// VerifyTransaction verifies transaction input signatures
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

func dbExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}
