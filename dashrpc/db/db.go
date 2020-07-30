package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
	"time"
)

const timeout = time.Second * 30

func GetContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	return ctx
}

// drops ALL data from the database, schema included
func DropAll(c *dgo.Dgraph) error {
	return c.Alter(GetContext(), &api.Operation{
		DropOp: api.Operation_ALL,
	})
}

// create a new dgraph client connecting to the specified host and port
func CreateClient(host string, port uint) (*dgo.Dgraph, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), grpc.WithInsecure())

	if err != nil {
		return nil, conn, err
	}

	return dgo.NewDgraphClient(api.NewDgraphClient(conn)), conn, nil
}

// create a new dgraph client with default connection values
func CreateDefaultClient() (*dgo.Dgraph, *grpc.ClientConn, error) {
	return CreateClient("localhost", 9080)
}

// gets the number of instances of the given type in the database
func GetCount(c *dgo.Dgraph, dbType string) (count uint64, err error) {
	query := fmt.Sprintf(`{
				 q(func: type(%s)){
					count(uid)
				  }
				}
				`, dbType)

	resp, err := c.NewReadOnlyTxn().Query(GetContext(), query)

	if err != nil {
		return
	}

	var r struct {
		GetCount []struct {
			Count uint64 `json:"count"`
		} `json:"q"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return
	}

	if len(r.GetCount) != 1 {
		err = errors.New("wrong number of objects returned")
		return
	}

	count = r.GetCount[0].Count

	return
}
