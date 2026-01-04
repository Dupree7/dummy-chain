package rpc

import (
	"bytes"
	"dummy-chain/common/types"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	url    string
	client *http.Client
}

func NewClient(url string) (*Client, error) {
	return &Client{
		url:    url,
		client: http.DefaultClient,
	}, nil
}

func (c *Client) Call(method string, param interface{}, result interface{}) error {
	reqBody := rpcRequest{
		JsonRpc: "2.0",
		Method:  method,
		Params:  []interface{}{param},
		Id:      1,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := c.client.Post(c.url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var rpcResp rpcResponse
	if err = json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return err
	}
	if rpcResp.Error != nil {
		return fmt.Errorf("RPC Error: %v", rpcResp.Error)
	}

	return json.Unmarshal(rpcResp.Result, result)
}

func (c *Client) GetBlocksInterval(left, right uint64) (*types.BlockInfoList, error) {
	var list types.BlockInfoList
	err := c.Call("chain.GetBlocksInterval", BlockInterval{
		Left:  left,
		Right: right,
	}, &list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func (c *Client) GetAccountInfo(address string) (*types.AccountInfo, error) {
	var account types.AccountInfo
	err := c.Call("chain.GetAccountInfo", address, &account)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (c *Client) GetCurrenBlockHeight() (uint64, error) {
	var height uint64
	err := c.Call("chain.GetCurrenBlockHeight", nil, &height)
	if err != nil {
		return 0, err
	}
	return height, nil
}

func (c *Client) SendTransaction(base64Tx string) error {
	reply := false
	return c.Call("chain.SendTransaction", base64Tx, &reply)
}

type rpcRequest struct {
	JsonRpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	Id      int         `json:"id"`
}

type rpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  interface{}     `json:"error"`
	Id     int             `json:"id"`
}
