package main

import (
	"dashrpc"
	dbop "dashrpc/db/output"
	dbtx "dashrpc/db/transaction"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"log"
	"os"
	"strconv"
)

//
// This is a proof-of-concept implementation of private-send traversal
//
//
//

// Result represents the data of final CreateDenominations TXs
type Result struct {
	hash   string
	tx     *dashrpc.TxDetails
	rounds int
}

// Writes the results of "search" to "outputFile"
func transactionSearch(dgraph *dgo.Dgraph, tx string, outputFile string) (err error, res map[string]*Result) {
	recordFile, err := os.Create(outputFile)
	if err != nil {
		errMsg := fmt.Sprintln("error while creating the file ::", err)
		err = errors.New(errMsg)
		return err, res
	}

	// Initialize the writer
	writer := csv.NewWriter(recordFile)
	res = search(dgraph, tx, writer)
	if res == nil {
		errMsg := "Result in NIL -- fix it."
		// no results -> delete result file
		if err := os.Remove(outputFile); err != nil {
			errMsg += "\n" + err.Error()
		}
		err = errors.New(errMsg)
		return err, res
	}

	writer.Flush()       // Writes the buffered data to the writer
	err = writer.Error() // Checks if any error occurred while writing
	if err != nil {
		errMsg := fmt.Sprintln("\"error while writing to the file ::\"", err)
		err = errors.New(errMsg)

		return err, res
	}
	err = recordFile.Close()
	if err != nil {
		errMsg := fmt.Sprintln("Error while closing the file ::", err)
		err = errors.New(errMsg)
		return err, res
	}

	// fmt.Printf("%v\n\n", res)
	return err, res
}

// search initiates recursive search through all inputs to find all CreateDenominations
func search(dgraph *dgo.Dgraph, txHash string, writer *csv.Writer) map[string]*Result {
	tx, err := dbtx.GetTransaction(dgraph, txHash)
	if err != nil {
		log.Println(err)
	}
	// Sanity check
	if !tx.IsPrivateSend() {
		log.Printf("Error: TX is not a PrivateSend! %v", tx)
		return nil
	}
	results := make(map[string]*Result)
	var txCount int64
	for _, input := range tx.Inputs {
		var rounds int // sets to 0, initially
		searchCreateDenominations(dgraph, input.Uid, rounds, &results, &txCount, writer)
	}

	return results
}

func searchCreateDenominations(dgraph *dgo.Dgraph, outputUid string, rounds int,
	results *map[string]*Result, txCount *int64, writer *csv.Writer) {

	if (*txCount % 100000) == 0 {
		log.Printf("\n%v\n", *txCount)
	}

	op, err := dbop.GetVerboseOutputByUid(dgraph, outputUid)
	if err != nil {
		log.Println("Problem getting output", err)
		return
	}
	hash := op.OutputTransaction

	tx, err := dbtx.GetTransaction(dgraph, hash)
	if err != nil {
		// let's ignore it -- our DB does not have ALL TXs
		return
	}
	*txCount++
	// End Condition
	if tx.IsCreateDenominations() {
		o, err := dbop.GetVerboseOutputByUid(dgraph, tx.Inputs[0].Uid)
		if err != nil {
			log.Println("Problem getting output", err)
			return
		}

		if _, ok := (*results)[o.Addresses[0]]; ok {
			return
		}
		r := Result{}
		r.hash = tx.Hash
		// todo check if necessary
		//r.tx = &tx
		r.rounds = rounds
		(*results)[o.Addresses[0]] = &r
		rec := []string{
			strconv.Itoa(len(*results)),
			o.Addresses[0],
			strconv.FormatFloat(*tx.Inputs[0].Amount, 'f', -1, 64),
			r.hash,
			//todo
			//strconv.FormatInt(tx.Timestamp, 10),
			strconv.Itoa(rounds),
		}
		err = writer.Write(rec)
		if err != nil {
			log.Println(err)
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
		searchCreateDenominations(dgraph, in2.Uid, rounds2, results, txCount, writer)
	}
}
