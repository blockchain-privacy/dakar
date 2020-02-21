package main

import (
	"dashrpc"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/dgraph-io/badger"
	"log"
	"os"
	"strconv"
)

// Result represents the data of final CreateDenominations TXs
type Result struct {
	hash string
	tx *dashrpc.TxDetails
	rounds int
}

// search initiates recursive search through all inputs to find all CreateDenominations
func search(db *badger.DB, txHash string, writer *csv.Writer) map[string]*Result {
	tx := dashrpc.TxDetails{}
	hash := txHash
	err := dashrpc.DbGetTxDetails(db, hash, &tx)
	if err != nil {
		fmt.Println("Error", err)
	}
	// Sanity check
	if !tx.IsPrivateSend() {
		fmt.Printf("Error: TX is not a PrivateSend! %v", tx)
		return nil
	}
	results := make(map[string]*Result)
	var txCount int64
	for _, input := range tx.Inputs {
		var rounds int // sets to 0, initially
		searchCreateDenominations(db, &input, rounds, &results, &txCount, writer)
	}

	return results
}


func searchCreateDenominations(db *badger.DB,
	in *dashrpc.TxOutput, rounds int,
	results *map[string]*Result, txCount *int64, writer *csv.Writer) {

	if (*txCount % 100000) == 0 {
		fmt.Printf("\n%v\n", *txCount)
	}
	tx := dashrpc.TxDetails{}
	hash := in.TxHash
	err := dashrpc.DbGetTxDetails(db, hash, &tx)
	if err != nil {
		// let's ignore it -- our DB does not have ALL TXs
		return
	}
	*txCount++
	// End Condition
	if tx.IsCreateDenominations() {
		if _, ok := (*results)[tx.Inputs[0].Addresses[0]]; ok {
			return
		}
		r := Result{}
		r.hash = tx.Hash
		r.tx = &tx
		r.rounds = rounds
		(*results)[tx.Inputs[0].Addresses[0]] = &r
		rec := []string{
			strconv.Itoa(len(*results)),
			tx.Inputs[0].Addresses[0],
			strconv.FormatFloat(tx.Inputs[0].Amount, 'f', -1, 64),
			r.hash,
			strconv.FormatInt(tx.Timestamp, 10),
			strconv.Itoa(rounds),
		}
		err := writer.Write(rec)
		if err != nil {
			fmt.Println(err)
		}
		writer.Flush()
		return
	}
	rounds++
	// End Condition based on MAX rounds
	if rounds > 16 {
		return
	}

	for _, in2 := range tx.Inputs {
		rounds2 := rounds
		searchCreateDenominations(db, &in2, rounds2, results, txCount, writer)
	}
}

//
// Simple utility to browse/lookup the TXs from the badger database
// Work in Progress. NOT WORKING YET.
//
func main() {
	badgerDir := flag.String("db", "/tmp/badger", "badger database location")
	txHash := flag.String("txhash", "", "Tx Hash")

	flag.Parse()
	// Open the Badger database located in the /tmp/badger directory.
	// It will be created if it doesn't exist.
	opts := badger.DefaultOptions(*badgerDir)
	// in badger 1.6.0 this is not needed to set explicit anymore
	// opts.Dir = *badgerDir
	// opts.ValueDir = *badgerDir
	opts.WithNumMemtables(50)
	opts.WithMaxTableSize(512 << 20)
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	recordFile, err := os.Create("./results.csv")
	if err != nil {
		fmt.Println("Error while creating the file ::", err)
		return
	}

	// Initialize the writer
	writer := csv.NewWriter(recordFile)
	res := search(db, *txHash, writer)
	if res == nil {
		fmt.Println("Result in NIL -- fix it.")
		return
	}

	writer.Flush() // Writes the buffered data to the writer
	err = writer.Error() // Checks if any error occurred while writing
	if err != nil {
		fmt.Println("Error while writing to the file ::", err)
		return
	}
	err = recordFile.Close()
	if err != nil {
		fmt.Println("Error while closing the file ::", err)
		return
	}

	// fmt.Printf("%v\n\n", res)
	fmt.Printf("Final map has %v elements\n", len(res))
	fmt.Printf("\n")
}
