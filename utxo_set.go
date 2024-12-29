package main

import (
	"encoding/hex"
	"log"

	"github.com/nutsdb/nutsdb"
)

const utxoBucket = "chainstate"

// UTXOSet represents UTXO set
type UTXOSet struct {
	Blockchain *Blockchain
}

// FindSpendableOutputs finds and returns unspent outputs to reference in inputs
func (u UTXOSet) FindSpendableOutputs(pubkeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	db := u.Blockchain.db

	if err := db.View(func(tx *nutsdb.Tx) error {
		if keys, values, err := tx.GetAll(utxoBucket); err != nil {
			log.Panic(err)
		} else {
			for index, key := range keys {
				txID := hex.EncodeToString(key)
				outs := DeserializeOutputs(values[index])
				for outIdx, out := range outs.Outputs {
					if out.IsLockedWithKey(pubkeyHash) && accumulated < amount {
						accumulated += out.Value
						unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
					}
				}
			}
		}

		return nil
	}); err != nil {
		log.Panic(err)
	}

	return accumulated, unspentOutputs
}

// FindUTXO finds UTXO for a public key hash
func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	db := u.Blockchain.db

	if err := db.View(func(tx *nutsdb.Tx) error {
		if _, values, err := tx.GetAll(utxoBucket); err != nil {
			log.Panic(err)
		} else {
			for _, v := range values {
				outs := DeserializeOutputs(v)
				for _, out := range outs.Outputs {
					if out.IsLockedWithKey(pubKeyHash) {
						UTXOs = append(UTXOs, out)
					}
				}
			}
		}
		return nil
	}); err != nil {
		log.Panic(err)
	}

	return UTXOs
}

// CountTransactions returns the number of transactions in the UTXO set
func (u UTXOSet) CountTransactions() int {
	db := u.Blockchain.db
	counter := 0

	if err := db.View(func(tx *nutsdb.Tx) error {
		_, values, err := tx.GetAll(utxoBucket)
		if err != nil {
			log.Panic(err)
		}
		counter = len(values)

		return nil
	}); err != nil {
		log.Panic(err)
	}

	return counter
}

// Reindex rebuilds the UTXO set
func (u UTXOSet) Reindex() {
	db := u.Blockchain.db

	if err := db.Update(func(tx *nutsdb.Tx) error {
		if err := tx.DeleteBucket(nutsdb.DataStructureBTree, utxoBucket); err != nil && err != nutsdb.ErrBucketNotFound {
			log.Panic(err)
		}
		return nil
	}); err != nil {
		log.Panic(err)
	}

	if err := db.Update(func(tx *nutsdb.Tx) error {
		if err := tx.NewBucket(nutsdb.DataStructureBTree, utxoBucket); err != nil {
			log.Panic(err)
		}

		return nil
	}); err != nil {
		log.Panic(err)
	}

	UTXO := u.Blockchain.FindUTXO()

	_ = db.Update(func(tx *nutsdb.Tx) error {
		for txID, outs := range UTXO {
			if key, err := hex.DecodeString(txID); err != nil {
				log.Panic(err)
			} else {
				if err := tx.Put(utxoBucket, key, outs.Serialize(), 0); err != nil {
					log.Panic(err)
				}
			}
		}

		return nil
	})
}

// Update updates the UTXO set with transactions from the Block
// The Block is considered to be the tip of a blockchain
func (u UTXOSet) Update(block *Block) {
	db := u.Blockchain.db

	if err := db.Update(func(txn *nutsdb.Tx) error {
		for _, tx := range block.Transactions {
			if !tx.IsCoinbase() {
				for _, vin := range tx.Vin {
					updatedOuts := TXOutputs{}
					outsBytes, _ := txn.Get(utxoBucket, vin.Txid)
					outs := DeserializeOutputs(outsBytes)

					for outIdx, out := range outs.Outputs {
						if outIdx != vin.Vout {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					if len(updatedOuts.Outputs) == 0 {
						if err := txn.Delete(utxoBucket, vin.Txid); err != nil {
							log.Panic(err)
						}
					} else {
						if err := txn.Put(utxoBucket, vin.Txid, updatedOuts.Serialize(), 0); err != nil {
							log.Panic(err)
						}
					}

				}
			}

			newOutputs := TXOutputs{}
			// for _, out := range tx.Vout {
			// 	newOutputs.Outputs = append(newOutputs.Outputs, out)
			// }
			newOutputs.Outputs = append(newOutputs.Outputs, tx.Vout...)

			err := txn.Put(utxoBucket, tx.ID, newOutputs.Serialize(), 0)
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	}); err != nil {
		log.Panic(err)
	}
}
