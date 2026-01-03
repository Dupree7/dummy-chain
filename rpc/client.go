package rpc

import (
	"dummy-chain/common/types"
	"net/rpc"

	ecommon "github.com/ethereum/go-ethereum/common"
)

type Client struct {
	client *rpc.Client
}

func NewClient(url string) (*Client, error) {
	client, err := rpc.Dial("tcp", url)
	if err != nil {
		return nil, err
	}
	return &Client{client}, nil
}

func (c *Client) GetAccountInfo(address ecommon.Address, account *types.Account) error {
	return c.client.Call("chain.GetAccountInfo", address, account)
}
