package main

//HintReq is a request for a hint from a client to a server.
type HintReq struct{}

//HintResp is a response to a hint request.
type HintResp struct{}

//QueryReq is a PIR query from a client to a server.
type QueryReq struct {
	// Add real stuff here.

	// Debug & testing.
	Index int
}

//QueryResp is a response to a PIR query.
type QueryResp struct {
	// Add real stuff here.

	// Debug & testing
	Val string
}

// PIRServer is the interface that wraps the server methods.
type PIRServer interface {
	Hint(*HintReq) (*HintResp, error)
	Answer(*QueryReq) (*QueryResp, error)
}

// PIRClient is the interface that wraps the client methods.
type PIRClient interface {
	RequestHint() (*HintReq, error)
	InitHint(*HintResp) error
	Query(int) ([]*QueryReq, error)
	Reconstruct([]*QueryResp) (string, error)
}
