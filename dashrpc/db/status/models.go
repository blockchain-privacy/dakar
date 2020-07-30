package status

import "fmt"

// Information about the state of the database
//type Status struct {
//	LastBlockId      uint64 `json:"lastblockid"`
//	StopBlockId      uint64 `json:"stopblockid"`
//	LastBlockHash    string `json:"lastblockhash"`
//	Status           string `json:"status"`
//	GlobalBlockCount uint64 `json:"globalblockcount"`
//	GlobalTxCount    uint64 `json:"globaltxcount"`
//}

const DType = "Status"

type Status struct {
	Uid        string   `json:"uid,omitempty"`
	IsCrawling *bool    `json:"iscrawling,omitempty"`
	DType      []string `json:"dgraph.type,omitempty"`
}

func (s Status) String() string {
	output := fmt.Sprintf("Uid: %s", s.Uid)

	if s.IsCrawling != nil {
		output += fmt.Sprintf(", IsCrawling: %t", *s.IsCrawling)
	}
	return output
}

func (s *Status) SetDType() {
	s.DType = []string{DType}
}

type statusQuery struct {
	Q []Status `json:"q"`
}
