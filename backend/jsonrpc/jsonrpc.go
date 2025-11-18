// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package jsonrpc

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"gitlab.com/blockchain-privacy/gomisc/serror"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	retryDuration = time.Second * 5
	maxRetries    = 5
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

// doRequestWithRetry calls the given function. If the function returns io.EOF, the function
// is called a few more times. Between each call retryDuration is waited.
func doRequestWithRetry(f func() (*http.Response, error)) (*http.Response, error) {
	var err error
	var resp *http.Response
	var encounteredError bool
	for range maxRetries {
		if encounteredError {
			// Retry the transaction if it was aborted
			time.Sleep(retryDuration)
		}

		if resp, err = f(); errors.Is(err, io.EOF) {
			encounteredError = true
			continue
		}

		break
	}

	return resp, err
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
		return serror.New(err)
	}

	call := func() (*http.Response, error) {
		request, err := http.NewRequest(http.MethodPost, j.URI, bytes.NewBuffer(replyBuffer))
		if err != nil {
			return nil, serror.New(err)
		}

		request.SetBasicAuth(j.User, j.Password)

		r, err := j.httpClient.Do(request) //nolint:bodyclose
		if err != nil {
			return nil, serror.New(err)
		}

		return r, nil
	}

	// sometimes the connection is reset (error EOF) by the RPC server for no apparent reason,
	// in this case we just retry the request
	r, err := doRequestWithRetry(call) //nolint:bodyclose
	if err != nil {
		return serror.New(err)
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)

	if r.StatusCode >= 400 {
		return serror.FromFormat("status code: %d", r.StatusCode)
	}

	rpcResult := Response{
		Result: &result,
	}
	err = json.NewDecoder(r.Body).Decode(&rpcResult)
	if err != nil {
		return serror.New(err)
	}

	if rpcResult.Error != nil {
		return serror.FromStr(rpcResult.Error.Message)
	}

	return nil
}

// Batch makes mutliple RPC calls in one request.
func (j *Client) Batch(requests []Request, results []Response) error {
	replyBuffer, err := json.Marshal(requests)
	if err != nil {
		return serror.New(err)
	}

	call := func() (*http.Response, error) {
		request, err := http.NewRequest(http.MethodPost, j.URI, bytes.NewBuffer(replyBuffer))
		if err != nil {
			return nil, serror.New(err)
		}

		request.SetBasicAuth(j.User, j.Password)

		r, err := j.httpClient.Do(request) //nolint:bodyclose
		if err != nil {
			return nil, serror.New(err)
		}
		return r, err
	}

	// sometimes the connection is reset (error EOF) by the RPC server for no apparent reason,
	// in this case we just retry the request
	r, err := doRequestWithRetry(call) //nolint:bodyclose
	if err != nil {
		return serror.New(err)
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)

	if r.StatusCode >= 400 {
		return serror.FromFormat("status code: %d", r.StatusCode)
	}

	err = json.NewDecoder(r.Body).Decode(&results)
	if err != nil {
		return serror.New(err)
	}

	for _, rpcResult := range results {
		if rpcResult.Error != nil {
			return serror.FromStr(rpcResult.Error.Message)
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
}

type Vin struct {
	Coinbase  string     `json:"coinbase"`
	Txid      string     `json:"txid"`
	Vout      int32      `json:"vout"`
	ScriptSig *ScriptSig `json:"scriptSig"`
}

// IsCoinBase returns a bool to show if a Vin is a Coinbase one or not.
func (v *Vin) IsCoinBase() bool {
	return len(v.Coinbase) > 0
}

type ScriptPubKeyResult struct {
	Asm     string `json:"asm"`
	Hex     string `json:"hex,omitempty"`
	Type    string `json:"type"`
	Address string `json:"address"`
}

type Vout struct {
	Value        float64            `json:"value"`
	N            int32              `json:"n"`
	ScriptPubKey ScriptPubKeyResult `json:"scriptPubKey"`
}

type TxRawResult struct {
	Txid string `json:"txid"`
	Hash string `json:"hash,omitempty"`
	Size int32  `json:"size,omitempty"`
	Vin  []Vin  `json:"vin"`
	Vout []Vout `json:"vout"`
}

type GetBlockVerboseResult struct {
	Hash         string        `json:"hash"`
	Tx           []string      `json:"tx,omitempty"`
	RawTx        []TxRawResult `json:"rawtx,omitempty"` // Note: this field is always empty when verbose != 2.
	Time         int64         `json:"time"`
	PreviousHash string        `json:"previousblockhash"`
	NextHash     string        `json:"nextblockhash,omitempty"`
}

func (d *BlockchainClient) GetBlockCount() (int64, error) {
	var r int64
	if err := d.rpc.Call("getblockcount", nil, &r); err != nil {
		return 0, err
	}

	return r, nil
}

// SetTimeout sets the request timeout of the rpc client
func (d *BlockchainClient) SetTimeout(timeout time.Duration) {
	d.rpc.httpClient.Timeout = timeout
}

func (d *BlockchainClient) GetBlockVerbose(blockHash string) (*GetBlockVerboseResult, error) {
	var r GetBlockVerboseResult
	if err := d.rpc.Call("getblock", []any{blockHash, 1}, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

func (d *BlockchainClient) GetBlockHash(blockHeight int64) (string, error) {
	var r string
	if err := d.rpc.Call("getblockhash", []any{blockHeight}, &r); err != nil {
		return "", err
	}

	return r, nil
}

func (d *BlockchainClient) GetRawTransactionVerbose(txHash string) (*TxRawResult, error) {
	var r TxRawResult
	if err := d.rpc.Call("getrawtransaction", []any{txHash, 1}, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

func (d *BlockchainClient) GetRawTransactionVerboseBatch(txs []string) ([]*TxRawResult, error) {
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

	if err := d.rpc.Batch(requests, results); err != nil {
		return nil, err
	}

	txResults := make([]*TxRawResult, len(results))
	for i, batchResult := range results {
		converted, ok := batchResult.Result.(*TxRawResult)
		if !ok {
			return nil, serror.FromStr("conversion of rpc result to type failed")
		}
		txResults[i] = converted
	}

	return txResults, nil
}

// GenerateToAddress mines a new block and rewards the resulting coins to the given address
func (d *BlockchainClient) GenerateToAddress(numBlocks int, address string) ([]string, error) {
	var blockHashes []string
	if err := d.rpc.Call("generatetoaddress", []any{numBlocks, address}, &blockHashes); err != nil {
		return nil, err
	}

	return blockHashes, nil
}

// GetNewAddress creates a new address in the current wallet. Fails if now wallet is loaded.
func (d *BlockchainClient) GetNewAddress() (string, error) {
	var newAddress string
	if err := d.rpc.Call("getnewaddress", []any{}, &newAddress); err != nil {
		return "", err
	}

	return newAddress, nil
}

// CreateWallet creates a wallet with the given file name. Fails if the wallet already exists
func (d *BlockchainClient) CreateWallet(name string) (string, error) {
	var newName string
	if err := d.rpc.Call("createwallet", []any{name}, &newName); err != nil {
		return "", err
	}

	return newName, nil
}

// LoadWallet loads a wallet with the given file name: Fails if the wallet is already loaded
func (d *BlockchainClient) LoadWallet(fileName string) (string, error) {
	var newName string
	if err := d.rpc.Call("loadwallet", []any{fileName}, &newName); err != nil {
		return "", err
	}

	return newName, nil
}
