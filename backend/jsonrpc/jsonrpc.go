package jsonrpc

import (
	"backend/cmd/cliutil"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type Client struct {
	httpClient *http.Client
	URI        string
	User       string
	Password   string
}

func NewClient(host string, user string, password string) *Client {
	return &Client{
		httpClient: &http.Client{},
		URI:        "http://" + host,
		User:       user,
		Password:   password,
	}
}

type Request struct {
	Version string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
}

type Response struct {
	Result any `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (j Client) Call(method string, params []any, result any) error {
	replyBuffer, err := json.Marshal(Request{
		Version: "1.0",
		Method:  method,
		Params:  params,
	})
	if err != nil {
		return cliutil.NewStackError(err)
	}

	request, err := http.NewRequest(http.MethodPost, j.URI, bytes.NewBuffer(replyBuffer))
	if err != nil {
		return cliutil.NewStackError(err)
	}

	request.SetBasicAuth(j.User, j.Password)

	r, err := j.httpClient.Do(request) //nolint:bodyclose
	if err != nil {
		return cliutil.NewStackError(err)
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)

	if r.StatusCode >= 400 {
		return cliutil.NewStackErrorf("status code: %d", r.StatusCode)
	}

	rpcResult := Response{
		Result: &result,
	}
	err = json.NewDecoder(r.Body).Decode(&rpcResult)
	if err != nil {
		return cliutil.NewStackError(err)
	}

	if rpcResult.Error != nil {
		return cliutil.NewStackErrorStr(rpcResult.Error.Message)
	}

	return nil
}

func (j Client) Batch(requests []Request, results []Response) error {
	replyBuffer, err := json.Marshal(requests)
	if err != nil {
		return cliutil.NewStackError(err)
	}

	request, err := http.NewRequest(http.MethodPost, j.URI, bytes.NewBuffer(replyBuffer))
	if err != nil {
		return cliutil.NewStackError(err)
	}

	request.SetBasicAuth(j.User, j.Password)

	r, err := j.httpClient.Do(request) //nolint:bodyclose
	if err != nil {
		return cliutil.NewStackError(err)
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)

	if r.StatusCode >= 400 {
		return cliutil.NewStackErrorf("status code: %d", r.StatusCode)
	}

	err = json.NewDecoder(r.Body).Decode(&results)
	if err != nil {
		return cliutil.NewStackError(err)
	}

	for _, rpcResult := range results {
		if rpcResult.Error != nil {
			return cliutil.NewStackErrorStr(rpcResult.Error.Message)
		}
	}

	return nil
}

type DashClient struct {
	rpc *Client
}

func NewDashClient(host string, user string, password string) *DashClient {
	return &DashClient{
		rpc: NewClient(host, user, password),
	}
}

type ScriptSig struct {
	Asm string `json:"asm"`
	Hex string `json:"hex"`
}

type Vin struct {
	Coinbase  string     `json:"coinbase"`
	Txid      string     `json:"txid"`
	Vout      uint32     `json:"vout"`
	ScriptSig *ScriptSig `json:"scriptSig"`
	Sequence  uint32     `json:"sequence"`
	Witness   []string   `json:"txinwitness"`
}

// IsCoinBase returns a bool to show if a Vin is a Coinbase one or not.
func (v *Vin) IsCoinBase() bool {
	return len(v.Coinbase) > 0
}

type ScriptPubKeyResult struct {
	Asm       string   `json:"asm"`
	Hex       string   `json:"hex,omitempty"`
	ReqSigs   int32    `json:"reqSigs,omitempty"`
	Type      string   `json:"type"`
	Addresses []string `json:"addresses,omitempty"`
}

type Vout struct {
	Value        float64            `json:"value"`
	N            uint32             `json:"n"`
	ScriptPubKey ScriptPubKeyResult `json:"scriptPubKey"`
}

type TxRawResult struct {
	Hex           string `json:"hex"`
	Txid          string `json:"txid"`
	Hash          string `json:"hash,omitempty"`
	Size          int32  `json:"size,omitempty"`
	Vsize         int32  `json:"vsize,omitempty"`
	Weight        int32  `json:"weight,omitempty"`
	Version       int16  `json:"version"`
	LockTime      uint32 `json:"locktime"`
	Vin           []Vin  `json:"vin"`
	Vout          []Vout `json:"vout"`
	BlockHash     string `json:"blockhash,omitempty"`
	Confirmations uint64 `json:"confirmations,omitempty"`
	Time          int64  `json:"time,omitempty"`
	Blocktime     int64  `json:"blocktime,omitempty"`
}

type GetBlockVerboseResult struct {
	Hash          string        `json:"hash"`
	Confirmations int64         `json:"confirmations"`
	StrippedSize  int32         `json:"strippedsize"`
	Size          int32         `json:"size"`
	Weight        int32         `json:"weight"`
	Height        int64         `json:"height"`
	Version       int32         `json:"version"`
	VersionHex    string        `json:"versionHex"`
	MerkleRoot    string        `json:"merkleroot"`
	Tx            []string      `json:"tx,omitempty"`
	RawTx         []TxRawResult `json:"rawtx,omitempty"` // Note: this field is always empty when verbose != 2.
	Time          int64         `json:"time"`
	Nonce         uint32        `json:"nonce"`
	Bits          string        `json:"bits"`
	Difficulty    float64       `json:"difficulty"`
	PreviousHash  string        `json:"previousblockhash"`
	NextHash      string        `json:"nextblockhash,omitempty"`
}

func (d DashClient) GetBlockCount() (int64, error) {
	var r int64
	err := d.rpc.Call("getblockcount", nil, &r)
	if err != nil {
		return 0, cliutil.NewStackError(err)
	}

	return r, nil
}

func (d DashClient) GetBlockVerbose(blockHash string) (*GetBlockVerboseResult, error) {
	var r GetBlockVerboseResult
	err := d.rpc.Call("getblock", []any{blockHash, 1}, &r)
	if err != nil {
		return nil, cliutil.NewStackError(err)
	}

	return &r, nil
}

func (d DashClient) GetBlockHash(blockHeight int64) (string, error) {
	var r string
	err := d.rpc.Call("getblockhash", []any{blockHeight}, &r)
	if err != nil {
		return "", cliutil.NewStackError(err)
	}

	return r, nil
}

func (d DashClient) GetRawTransactionVerbose(txHash string) (*TxRawResult, error) {
	var r TxRawResult
	err := d.rpc.Call("getrawtransaction", []any{txHash, true}, &r)
	if err != nil {
		return nil, cliutil.NewStackError(err)
	}

	return &r, nil
}

func (d DashClient) GetRawTransactionVerboseBatch(txs []string) ([]*TxRawResult, error) {
	request := make([]Request, len(txs))
	results := make([]Response, len(txs))
	thisRequest := Request{
		Version: "2.0",
		Method:  "getrawtransaction",
	}
	for i, tx := range txs {
		thisRequest.Params = []any{tx, true}
		request[i] = thisRequest
		results[i] = Response{Result: &TxRawResult{}}
	}

	err := d.rpc.Batch(request, results)
	if err != nil {
		return nil, cliutil.NewStackError(err)
	}

	txResults := make([]*TxRawResult, len(results))
	for i, batchResult := range results {
		converted, ok := batchResult.Result.(*TxRawResult)
		if !ok {
			return nil, cliutil.NewStackErrorStr("not able to convert rpc result to type")
		}
		txResults[i] = converted
	}

	return txResults, nil
}
