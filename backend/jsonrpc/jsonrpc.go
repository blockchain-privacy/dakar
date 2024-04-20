package jsonrpc

import (
	"backend/cmd/cliutil"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"net/http"
	"sync"
)

type Client struct {
	httpClient *http.Client
	URI        string
	User       string
	Password   string
	// mutex controls access to id
	mutex sync.Mutex
	id    int
}

// NewClient creates a new rpc client, which uses the given user and password when making a request to host.
// If cert is not nil, a TLS connection is created, the passed certificate is not validated.
func NewClient(host string, user string, password string, cert []byte) *Client {
	httpProtocol := "http://"
	var tlsConfig *tls.Config
	if cert != nil {
		// set custom certificate without validation
		certs := x509.NewCertPool()
		certs.AppendCertsFromPEM(cert)
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec
			RootCAs:            certs,
		}
		httpProtocol = "https://"
	}

	return &Client{
		httpClient: &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}},
		URI:        httpProtocol + host,
		User:       user,
		Password:   password,
	}
}

type Request struct {
	Version string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []any  `json:"params,omitempty"`
	// ID should be unique per session
	ID *int `json:"id,omitempty"`
}

type Response struct {
	Result any `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// NewRequestID creates a new request ID which will be unique per client instance
func (j *Client) NewRequestID() *int {
	j.mutex.Lock()
	newID := j.id
	j.id++
	j.mutex.Unlock()
	return &newID
}

// Call makes a call to the provided method with the given parameters and stores the returned values into result
func (j *Client) Call(method string, params []any, result any) error {
	replyBuffer, err := json.Marshal(Request{
		Version: "1.0",
		Method:  method,
		Params:  params,
		ID:      j.NewRequestID(),
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

// Batch makes mutliple RPC calls in one request.
func (j *Client) Batch(requests []Request, results []Response) error {
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

// BlockchainClient implements several RPCs of Bitcoin and Dash
type BlockchainClient struct {
	rpc *Client
}

// NewBlockchainClient creates a new rpc client, which uses the given user and password when making a request to host.
// If cert is not nil, a TLS connection is created, the passed certificate is not validated.
func NewBlockchainClient(host string, user string, password string, cert []byte) *BlockchainClient {
	return &BlockchainClient{rpc: NewClient(host, user, password, cert)}
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
	Asm  string `json:"asm"`
	Hex  string `json:"hex,omitempty"`
	Type string `json:"type"`
	// Dash still uses Addresses instead of Address
	Addresses []string `json:"addresses,omitempty"`
	// Since Bitcoin Core v24 Address is used instead of Addresses is not set https://github.com/bitcoin/bitcoin/pull/20286
	Address string `json:"address,omitempty"`
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

func (d BlockchainClient) GetBlockCount() (int64, error) {
	var r int64
	err := d.rpc.Call("getblockcount", nil, &r)
	if err != nil {
		return 0, cliutil.NewStackError(err)
	}

	return r, nil
}

func (d BlockchainClient) GetBlockVerbose(blockHash string) (*GetBlockVerboseResult, error) {
	var r GetBlockVerboseResult
	err := d.rpc.Call("getblock", []any{blockHash, 1}, &r)
	if err != nil {
		return nil, cliutil.NewStackError(err)
	}

	return &r, nil
}

func (d BlockchainClient) GetBlockHash(blockHeight int64) (string, error) {
	var r string
	err := d.rpc.Call("getblockhash", []any{blockHeight}, &r)
	if err != nil {
		return "", cliutil.NewStackError(err)
	}

	return r, nil
}

func (d BlockchainClient) GetRawTransactionVerbose(txHash string) (*TxRawResult, error) {
	var r TxRawResult
	err := d.rpc.Call("getrawtransaction", []any{txHash, 1}, &r)
	if err != nil {
		return nil, cliutil.NewStackError(err)
	}

	return &r, nil
}

func (d BlockchainClient) GetRawTransactionVerboseBatch(txs []string) ([]*TxRawResult, error) {
	requests := make([]Request, len(txs))
	results := make([]Response, len(txs))
	request := Request{
		Version: "2.0",
		Method:  "getrawtransaction",
	}
	for i, tx := range txs {
		request.Params = []any{tx, 1}
		request.ID = d.rpc.NewRequestID()
		requests[i] = request
		results[i] = Response{Result: &TxRawResult{}}
	}

	err := d.rpc.Batch(requests, results)
	if err != nil {
		return nil, cliutil.NewStackError(err)
	}

	txResults := make([]*TxRawResult, len(results))
	for i, batchResult := range results {
		converted, ok := batchResult.Result.(*TxRawResult)
		if !ok {
			return nil, cliutil.NewStackErrorStr("conversion of rpc result to type failed")
		}
		txResults[i] = converted
	}

	return txResults, nil
}

// GenerateToAddress mines a new block and rewards the resulting coins to the given address
func (d BlockchainClient) GenerateToAddress(numBlocks int, address string) ([]string, error) {
	var blockHashes []string
	err := d.rpc.Call("generatetoaddress", []any{numBlocks, address}, &blockHashes)
	if err != nil {
		return nil, cliutil.NewStackError(err)
	}

	return blockHashes, nil
}

// GetNewAddress creates a new address in the current wallet. Fails if now wallet is loaded.
func (d BlockchainClient) GetNewAddress() (string, error) {
	var newAddress string
	err := d.rpc.Call("getnewaddress", []any{}, &newAddress)
	if err != nil {
		return "", cliutil.NewStackError(err)
	}

	return newAddress, nil
}

// CreateWallet creates a wallet with the given file name. Fails if the wallet already exists
func (d BlockchainClient) CreateWallet(name string) (string, error) {
	var newName string
	err := d.rpc.Call("createwallet", []any{name}, &newName)
	if err != nil {
		return "", cliutil.NewStackError(err)
	}

	return newName, nil
}

// LoadWallet loads a wallet with the given file name: Fails if the wallet is already loaded
func (d BlockchainClient) LoadWallet(fileName string) (string, error) {
	var newName string
	err := d.rpc.Call("loadwallet", []any{fileName}, &newName)
	if err != nil {
		return "", cliutil.NewStackError(err)
	}

	return newName, nil
}
