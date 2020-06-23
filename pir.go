package boosted

// One database row.
type Row []byte

//HintReq is a request for a hint from a client to a server.
type HintReq struct {
	Sets          []*SetKey
	AuxRecordsSet *SetKey
}

//HintResp is a response to a hint request.
type HintResp struct {
	Hints      []Row
	AuxRecords map[int]Row
}

//QueryReq is a PIR query from a client to a server.
type QueryReq struct {
	PuncturedSet Set

	// Debug & testing.
	Index int
}

//QueryResp is a response to a PIR query.
type QueryResp struct {
	Answer Row

	// Debug & testing
	Val Row
}
