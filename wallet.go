package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"math/big"

	"golang.org/x/crypto/ripemd160"
)

const (
	version            = byte(0x00)
	addressChecksumLen = 4
)

// Wallet stores private and public keys
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// NewWallet creates and returns a Wallet
func NewWallet() *Wallet {
	priv, pub := newKeyPair()
	return &Wallet{priv, pub}
}

// GetAddress returns wallet address
func (w Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(w.PublicKey)

	versionedPayload := append([]byte{version}, pubKeyHash...)
	chksum := checksum(versionedPayload)

	fullPayload := append(versionedPayload, chksum...)

	return Base58Encode(fullPayload)
}

// HashPubKey hashes public key
func HashPubKey(pubKey []byte) []byte {
	pubSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(pubSHA256[:])
	if err != nil {
		log.Panic(err)
	}

	return RIPEMD160Hasher.Sum(nil)
}

// ValidateAddress check if address if valid
func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Equal(actualChecksum, targetChecksum)
}

// Checksum generates a checksum for a public key
func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	priv, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	pubKey := append(priv.PublicKey.X.Bytes(), priv.PublicKey.Y.Bytes()...)

	return *priv, pubKey
}

type _PrivateKey struct {
	D          *big.Int
	PublicKeyX *big.Int
	PublicKeyY *big.Int
}

func (w *Wallet) GobEncode() ([]byte, error) {
	privKey := &_PrivateKey{
		D:          w.PrivateKey.D,
		PublicKeyX: w.PrivateKey.PublicKey.X,
		PublicKeyY: w.PrivateKey.PublicKey.Y,
	}

	var buf bytes.Buffer

	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(privKey)
	if err != nil {
		return nil, err
	}

	_, err = buf.Write(w.PublicKey)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (w *Wallet) GobDecode(data []byte) error {
	buf := bytes.NewBuffer(data)
	var privKey _PrivateKey

	decoder := gob.NewDecoder(buf)
	err := decoder.Decode(&privKey)
	if err != nil {
		return err
	}

	w.PrivateKey = ecdsa.PrivateKey{
		D: privKey.D,
		PublicKey: ecdsa.PublicKey{
			X:     privKey.PublicKeyX,
			Y:     privKey.PublicKeyY,
			Curve: elliptic.P256(),
		},
	}
	w.PublicKey = make([]byte, buf.Len())
	_, err = buf.Read(w.PublicKey)
	if err != nil {
		return err
	}

	return nil
}
