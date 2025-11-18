// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import (
	"backend/jsonrpc"
	_ "embed"
	"log"
	"os"
	"testing"
)

type ContainerName string

const (
	EnvDBTests               = "DB_TESTS"
	EnvRPCTests              = "RPC_TESTS"
	EnvDBHostname            = "DB_HOSTNAME"
	EnvRPCHostname           = "RPC_HOSTNAME"
	EnvDBUser                = "DB_USER"
	EnvDBPassword            = "DB_PASSWORD"
	UseClassifierFile        = "classifier"
	UseBlockFile             = "block"
	UsePrivacyFile           = "privacy"
	UseBTCPrivacyFile        = "btc_privacy"
	ClassifierFileFirstBlock = 1557775
	ClassifierFileLastBlock  = 1557830
	BlockFileFirstBlock      = 60000
	BlockFileLastBlock       = 60020
)

// BlockFile contains Dash blocks from height 60000 to 60020.
// This file includes block, transaction, address and cluster data.
//
//go:embed testfiles/blocks_60000_60020.json
var BlockFile []byte

// ClassifierFile contains Dash blocks from height 1557775 to 1557780.
// This file includes block, transaction, and address data.
//
//go:embed testfiles/blocks_1557775_1557830.json
var ClassifierFile []byte

// PrivacyFile contains a small transaction graph created by traversing forward beginning with tx
// 452f795486980ef698fe652b56597eef3e7f6ad155cb0c9f1de21254d9bd9b0e
//
//go:embed testfiles/privacy_transactions.json
var PrivacyFile []byte

// BTCPrivacyFile contains a bitcoin blocks between 573945 574040
//
//go:embed testfiles/btc_privacy_transactions.json
var BTCPrivacyFile []byte

func DoDBTests() bool {
	_, ok := os.LookupEnv(EnvDBTests)
	return ok
}

func DoRPCTests() bool {
	_, ok := os.LookupEnv(EnvRPCTests)
	return ok
}

func GetDBName() (string, bool) {
	return os.LookupEnv(EnvDBHostname)
}

func GetRPCName() (string, bool) {
	return os.LookupEnv(EnvRPCHostname)
}

func GetDBUser() string {
	user, ok := os.LookupEnv(EnvDBUser)
	if !ok {
		user = "groot"
	}
	return user
}

func GetDBPassword() string {
	user, ok := os.LookupEnv(EnvDBPassword)
	if !ok {
		user = "password"
	}
	return user
}

func SkipIfNoRPC(t testing.TB) {
	if !DoRPCTests() {
		t.SkipNow()
	}
}

// RunRPCTests sets up the RPC client connection and runs all tests
func RunRPCTests(m *testing.M, client *jsonrpc.BlockchainClient) {
	if DoRPCTests() {
		rpcHostname, ok := GetRPCName()
		if !ok {
			log.Panic("environment variable " + EnvRPCHostname + " is not set")
			return
		}

		rpcClient := jsonrpc.NewBlockchainClient(rpcHostname+":8131", "rpc1user", "1234pass", nil)

		*client = *rpcClient

		// wallet might already exist -> ignore error
		_, _ = client.CreateWallet("testwallet")
		// wallet might already be loaded -> ignore error
		_, _ = client.LoadWallet("testwallet")

		generateToAddress, err := client.GetNewAddress()
		if err != nil {
			log.Panic(err)
		}

		_, err = client.GenerateToAddress(5, generateToAddress)
		if err != nil {
			log.Panic(err)
		}
	}

	m.Run()
}

func GetPointer[number any](n number) *number {
	return &n
}
