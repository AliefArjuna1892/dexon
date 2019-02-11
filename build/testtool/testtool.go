package main

import (
	"bytes"
	"context"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dexon-foundation/dexon"
	"github.com/dexon-foundation/dexon/accounts/abi"
	"github.com/dexon-foundation/dexon/cmd/zoo/monkey"
	vm "github.com/dexon-foundation/dexon/core/vm/evm"
	"github.com/dexon-foundation/dexon/crypto"
	"github.com/dexon-foundation/dexon/ethclient"
	"github.com/dexon-foundation/dexon/internal/build"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("need subcommand as first argument")
	}

	switch os.Args[1] {
	case "verifyGovCRS":
		doVerifyGovCRS(os.Args[2:])
	case "verifyGovMPK":
		doVerifyGovMPK(os.Args[2:])
	case "monkeyTest":
		doMonkeyTest(os.Args[2:])
	case "upload":
		doUpload(os.Args[2:])
	}
}

func doVerifyGovCRS(args []string) {
	if len(args) < 2 {
		log.Fatal("arg length is not enough")
	}

	client, err := ethclient.Dial(args[0])
	if err != nil {
		log.Fatalf("new ethclient fail: %v", err)
	}

	abiObject, err := abi.JSON(strings.NewReader(vm.GovernanceABIJSON))
	if err != nil {
		log.Fatalf("read abi fail: %v", err)
	}

	round, err := strconv.Atoi(args[1])
	if err != nil {
		log.Fatalf("pasre round from arg 2 fail: %v", err)
	}

	input, err := abiObject.Pack("crs", big.NewInt(int64(round)))
	if err != nil {
		log.Fatalf("pack input fail: %v", err)
	}

	result, err := client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &vm.GovernanceContractAddress,
		Data: input,
	}, nil)
	if err != nil {
		log.Fatalf("call contract fail: %v", err)
	}

	if bytes.Equal(make([]byte, 32), result) {
		log.Fatalf("round %s crs not found", args[1])
	}

	log.Printf("get round %s crs %x", args[1], result)
}

func doVerifyGovMPK(args []string) {
	if len(args) < 3 {
		log.Fatal("arg length is not enough")
	}

	client, err := ethclient.Dial(args[0])
	if err != nil {
		log.Fatalf("new ethclient fail: %v", err)
	}

	abiObject, err := abi.JSON(strings.NewReader(vm.GovernanceABIJSON))
	if err != nil {
		log.Fatalf("read abi fail: %v", err)
	}

	round, err := strconv.Atoi(args[1])
	if err != nil {
		log.Fatalf("pasre round from arg 2 fail: %v", err)
	}

	index, err := strconv.Atoi(args[2])
	if err != nil {
		log.Fatalf("pasre round from arg 2 fail: %v", err)
	}

	input, err := abiObject.Pack("dkgMasterPublicKeys", big.NewInt(int64(round)), big.NewInt(int64(index)))
	if err != nil {
		log.Fatalf("pack input fail: %v", err)
	}

	result, err := client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &vm.GovernanceContractAddress,
		Data: input,
	}, nil)
	if err != nil {
		log.Fatalf("call contract fail: %v", err)
	}

	if bytes.Equal(make([]byte, 0), result) {
		log.Fatalf("round %s index %s crs not found", args[1], args[2])
	}

	log.Printf("get round %s index %s master public key %x", args[1], args[2], result)
}

func doUpload(args []string) {
	auth := build.GCPOption{
		CredentialPath: os.Getenv("GCP_CREDENTIAL_PATH"),
	}

	if err := build.GCPFileUpload(args[0], args[1], filepath.Base(args[0]), auth); err != nil {
		log.Fatalf("upload fail: %v", err)
	}
}

func doMonkeyTest(args []string) {
	if len(args) < 1 {
		log.Fatal("arg length is not enough")
	}

	client, err := ethclient.Dial(args[0])
	if err != nil {
		log.Fatalf("new ethclient fail: %v", err)
	}

	monkey.Init(&monkey.MonkeyConfig{
		Key:      "test/keystore/monkey.key",
		Endpoint: args[0],
		N:        30,
		Sleep:    3000,
		Timeout:  60,
	})
	m, nonce := monkey.Exec()

	time.Sleep(10 * time.Second)

	for _, key := range m.Keys() {
		currentNonce, err := client.NonceAt(context.Background(), crypto.PubkeyToAddress(key.PublicKey), nil)
		if err != nil {
			log.Fatalf("get address nonce fail: %v", err)
		}

		if currentNonce != nonce+1 {
			log.Fatalf("expect nonce %v but %v", nonce, currentNonce)
		}
	}
}