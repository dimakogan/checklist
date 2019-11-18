package boosted

// One database row.
//
// I think we want byte-slices here instead of strings, since 
// we're going to handle arbitrary binary data. Iterating over
// a row using `range` should give us bytes back, rather than
// the UTF-8 runes.
type Row []byte

//HintReq is a request for a hint from a client to a server.
type HintReq struct{
  Key *SetKey
  Deltas []int
}

//HintResp is a response to a hint request.
type HintResp struct{
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

// PIRServer is the interface that wraps the server methods.
type PIRServer interface {
	Hint(*HintReq, *HintResp) error
	Answer(*QueryReq, *QueryResp) error
}

// PIRClient is the interface that wraps the client methods.
type PIRClient interface {
	RequestHint() (*HintReq, error)
	InitHint(*HintResp) error
	Query(int) ([]*QueryReq, error)
	Reconstruct([]*QueryResp) (Row, error)
}
