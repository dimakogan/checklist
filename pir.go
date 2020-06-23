package boosted

// One database row.
type Row []byte

//HintReq is a request for a hint from a client to a server.
type HintReq struct {
	Keys []*SetKey
}

//HintResp is a response to a hint request.
type HintResp struct {
	Hints []Row
}

//QueryReq is a PIR query from a client to a server.
type QueryReq struct {
	Key *PuncSetKey

	// Debug & testing.
	Index int
}

//QueryResp is a response to a PIR query.
type QueryResp struct {
	Answer Row

	// Debug & testing
	Val Row
}
