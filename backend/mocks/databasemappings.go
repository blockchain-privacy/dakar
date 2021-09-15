package mocks

import (
	"github.com/dgraph-io/dgo/v210/protos/api"
	"github.com/stretchr/testify/mock"
)

func MapUpsertAddresses(db *Database) {
	// only once because the input filter is very generic
	db.On("Mutate", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("*api.Request")).
		Return(nil, nil).Once()
}

func MapSetClassifying(db *Database) {
	// only once because the input filter is very generic
	db.On("Mutate", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("*api.Request")).
		Return(nil, nil).Once()
}

func MapGetClassifierStatus(db *Database) {
	var emptyMap map[string]string
	query := `{
				 q(func: type(ClassifierStatus)){
					uid
					isclassifying
					lastclassifiedid
				  }
				}`

	resp := api.Response{
		Json: []byte("{\"q\":[{\"uid\":\"0xcae3b25\",\"isclassifying\":true,\"lastclassifiedid\":1423346}]}"),
	}

	db.On("Query", mock.AnythingOfType("*context.timerCtx"), query, emptyMap).
		Return(&resp, nil)
}

func MapGetCrawlerStatus(db *Database) {
	var emptyMap map[string]string
	query := `{
				 q(func: type(CrawlerStatus)){
					uid
					iscrawling
					lastblockid
					lowestblockid
				  }
				}`

	resp := api.Response{
		Json: []byte("{\"q\":[{\"uid\":\"0x3d5\",\"iscrawling\":false,\"lastblockid\":1423346,\"lowestblockid\":1}]}"),
	}

	db.On("Query", mock.AnythingOfType("*context.timerCtx"), query, emptyMap).
		Return(&resp, nil)
}
