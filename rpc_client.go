package boosted

import (
	"fmt"
	"net/rpc"
)

type rpcPIRClient struct {
	remote *rpc.Client
	pir    *pirClientPunc
}

func NewRpcPirClient(remote *rpc.Client, pir *pirClientPunc) (*rpcPIRClient, error) {
	c := rpcPIRClient{
		remote: remote,
		pir:    pir,
	}
	hintReq, err := pir.RequestHint()
	if err != nil {
		return nil, fmt.Errorf("Client failed to RequestHint, %w", err)
	}
	var hintResp HintResp
	if err = remote.Call("PIRServer.Hint", hintReq, &hintResp); err != nil {
		return nil, fmt.Errorf("Remote PIRServer.Hint failed, %w ", err)
	}
	if err = pir.InitHint(&hintResp); err != nil {
		return nil, fmt.Errorf("Client failed to InitHint, %w", err)
	}
	return &c, nil
}

func (c *rpcPIRClient) Read(index int) (Row, error) {
	queryReq, err := c.pir.Query(index)
	if err != nil {
		return nil, fmt.Errorf("Client failed to Query, %w", err)
	}

	var queryResp QueryResp
	if err = c.remote.Call("PIRServer.Answer", queryReq[0], &queryResp); err != nil {
		return nil, fmt.Errorf("Remote Answer failed, %w", err)
	}

	val, err := c.pir.Reconstruct([]*QueryResp{&queryResp})
	if err != nil {
		return nil, fmt.Errorf("Client failed to Reconstruct, %w ", err)
	}

	return val, nil
}
