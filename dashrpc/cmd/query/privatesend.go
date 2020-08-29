package main

import (
	"dashrpc"
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
//func transactionSearch(dgraph *dgo.Dgraph, tx string, outputFile string) (err error, res map[string]*Result) {
//	recordFile, err := os.Create(outputFile)
//	if err != nil {
//		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
//		return
//	}
//
//	// Initialize the writer
//	writer := csv.NewWriter(recordFile)
//	res = search(dgraph, tx, writer)
//	if res == nil {
//		errMsg := "Result in NIL -- fix it."
//		// no results -> delete result file
//		if err := os.Remove(outputFile); err != nil {
//			errMsg += "\n" + err.Error()
//		}
//		err = errors.New(errMsg)
//		return
//	}
//
//	writer.Flush()       // Writes the buffered data to the writer
//	err = writer.Error() // Checks if any error occurred while writing
//	if err != nil {
//		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
//		return
//	}
//	err = recordFile.Close()
//	if err != nil {
//		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
//		return
//	}
//
//	// fmt.Printf("%v\n\n", res)
//	return
//}

// search initiates recursive search through all inputs to find all PrivacyOrigins
//func search(dgraph *dgo.Dgraph, txHash string, writer *csv.Writer) map[string]*Result {
//	tx, err := dbtx.GetTransaction(dgraph, txHash)
//	if err != nil {
//		log.Println(err)
//	}
//	// Sanity check
//	if !tx.IsPrivacyDestination() {
//		log.Printf("Error: TX is not a PrivateSend! %v", tx)
//		return nil
//	}
//	results := make(map[string]*Result)
//	var txCount int64
//	for _, input := range tx.Inputs {
//		var rounds int // sets to 0, initially
//		searchPrivacyOrigin(dgraph, input.Uid, rounds, &results, &txCount, writer)
//	}
//
//	return results
//}

//func searchPrivacyOrigin(dgraph *dgo.Dgraph, outputUid string, rounds int,
//	results *map[string]*Result, txCount *int64, writer *csv.Writer) {
//
//	if (*txCount % 100000) == 0 {
//		log.Printf("\n%v\n", *txCount)
//	}
//
//	op, err := dbop.GetVerboseOutputByUid(dgraph, outputUid)
//	if err != nil {
//		log.Println("Problem getting output", err)
//		return
//	}
//	hash := op.OutputTransaction
//
//	tx, err := dbtx.GetTransaction(dgraph, hash)
//	if err != nil {
//		// let's ignore it -- our DB does not have ALL TXs
//		return
//	}
//	*txCount++
//	// End Condition
//	if tx.IsPrivacyOrigin() {
//		o, err := dbop.GetVerboseOutputByUid(dgraph, tx.Inputs[0].Uid)
//		if err != nil {
//			log.Println("Problem getting output", err)
//			return
//		}
//
//		if _, ok := (*results)[o.Addresses[0]]; ok {
//			return
//		}
//		r := Result{}
//		r.hash = tx.Hash
//		// todo check if necessary
//		//r.tx = &tx
//		r.rounds = rounds
//		(*results)[o.Addresses[0]] = &r
//		rec := []string{
//			strconv.Itoa(len(*results)),
//			o.Addresses[0],
//			tx.Inputs[0].Amount,
//			r.hash,
//			//todo
//			//strconv.FormatInt(tx.Timestamp, 10),
//			strconv.Itoa(rounds),
//		}
//		err = writer.Write(rec)
//		if err != nil {
//			log.Println(err)
//		}
//		writer.Flush()
//		return
//	}
//	rounds++
//	// End Condition based on MAX rounds
//	if rounds > 16 {
//		return
//	}
//
//	for _, in2 := range tx.Inputs {
//		rounds2 := rounds
//		searchPrivacyOrigin(dgraph, in2.Uid, rounds2, results, txCount, writer)
//	}
//}
