package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func transferEth(client *ethclient.Client, privKey *ecdsa.PrivateKey, to ecdsa.PublicKey, amount int64) error {
	publicKey := privKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return err
	}

	value := big.NewInt(amount)
	gasLimit := uint64(21000)
	gasPrice := big.NewInt(30000000000)

	toAddress := crypto.PubkeyToAddress(to)

	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)
	chainID := big.NewInt(1337)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privKey)
	if err != nil {
		return err
	}

	return client.SendTransaction(context.Background(), signedTx)
}

func getBalance(client *ethclient.Client, address string) {
	account := common.HexToAddress(address)
	balance, err := client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(balance)
}
