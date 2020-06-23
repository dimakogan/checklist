package boosted

import (
	"math"
	"math/rand"

	"github.com/klauspost/reedsolomon"
)

type pirClientErasure struct {
	nEncodedRows int
	client       *pirClientPunc
}

type pirServerErasure struct {
	server *pirServerPunc
}

var CHUNK_SIZE = 128
var ALLOW_LOSS = 128

func nEncodedRows(nRows int) int {
	// We split the length-n database into chunks of size `chunkSize`
	// and then we encode each chunk using a erasure encoding that
	// tolerates the loss of at most L rows.

	// XXX choose `chunkSize` and `allowLoss` such that if you flip `chunkSize`
	// coins that come up heads with probability 1/e, you end up with
	// fewer than `allowLoss` heads with very high probability.

	// Our encoding library requires CHUNK_SIZE + ALLOW_LOSS <= 256

	return (nRows / CHUNK_SIZE) * (CHUNK_SIZE + ALLOW_LOSS)
}

func encodeDatabase(data []Row) []Row {

	enc, err := reedsolomon.New(CHUNK_SIZE, ALLOW_LOSS)
	if err != nil {
		panic("Could not create encoder.")
	}

	if len(data)%CHUNK_SIZE != 0 {
		panic("Haven't implemented this case")
	}

	encRows := nEncodedRows(len(data))
	encoded := make([]Row, encRows)
	rowLen := len(data[0])

	for i := 0; i < len(data)/CHUNK_SIZE; i++ {
		toEnc := make([][]byte, CHUNK_SIZE+ALLOW_LOSS)

		// Data chunks
		for j := 0; j < CHUNK_SIZE; j++ {
			toEnc[j] = data[i*CHUNK_SIZE+j]
		}

		// Parity chunks
		for j := 0; j < ALLOW_LOSS; j++ {
			toEnc[CHUNK_SIZE+j] = make([]byte, rowLen)
		}

		err := enc.Encode(toEnc)
		if err != nil {
			panic("Encoding error")
		}

		for j := 0; j < CHUNK_SIZE+ALLOW_LOSS; j++ {
			// fmt.Printf("Copying %v\n", i*CHUNK_SIZE+j)
			encoded[i*(CHUNK_SIZE+ALLOW_LOSS)+j] = toEnc[j]
		}
	}

	return encoded
}

func NewPirServerErasure(source *rand.Rand, data []Row) *pirServerErasure {
	encdata := encodeDatabase(data)
	// fmt.Printf("LenIn = %v\n", len(data))
	// fmt.Printf("LenOut = %v\n", len(encdata))
	server := NewPirServerPunc(source, encdata)

	return &pirServerErasure{
		server: server,
	}
}

func (s *pirServerErasure) Hint(req *HintReq, resp *HintResp) error {
	return s.server.Hint(req, resp)
}

func (s *pirServerErasure) Answer(q *QueryReq, resp *QueryResp) error {
	return s.server.Answer(q, resp)
}

func newPirClientErasure(source *rand.Rand, nRows int) *pirClientErasure {
	nEnc := nEncodedRows(nRows)
	nHints := int(math.Round(math.Pow(float64(nEnc), 0.5)))

	client := NewPirClientPunc(source, nEnc, nHints)

	return &pirClientErasure{
		nEncodedRows: nEnc,
		client:       client,
	}
}

func (c *pirClientErasure) RequestHint() (*HintReq, error) {
	return c.client.RequestHint()
}

func (c *pirClientErasure) InitHint(resp *HintResp) error {
	return c.client.InitHint(resp)
}

func (c *pirClientErasure) Query(i int) ([]*QueryReq, error) {
	return c.Query(i)
}

func (c *pirClientErasure) Reconstruct(resp []*QueryResp) (Row, error) {
	return c.Reconstruct(resp)
}
