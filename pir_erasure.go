package boosted

import (
	"fmt"
	"math/rand"

	"github.com/klauspost/reedsolomon"
)

type pirClientErasure struct {
	rs         reedsolomon.Encoder
	puncClient *pirClientPunc
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
		fmt.Errorf("Could not create encoder: %s", err)
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

func NewPirServerErasure(source *rand.Rand, data []Row) *pirServerPunc {
	encdata := encodeDatabase(data)
	// fmt.Printf("LenIn = %v\n", len(data))
	// fmt.Printf("LenOut = %v\n", len(encdata))
	return NewPirServerPunc(source, encdata)
}

func NewPirClientErasure(source *rand.Rand, nRows int, server PuncPirServer) (*pirClientErasure, error) {
	nEnc := nEncodedRows(nRows)
	rs, err := reedsolomon.New(CHUNK_SIZE, ALLOW_LOSS)
	if err != nil {
		return nil, fmt.Errorf("Could not create RS encoder: %w", err)
	}

	return &pirClientErasure{rs: rs, puncClient: NewPirClientPunc(source, nEnc, server)}, nil
}

func (c *pirClientErasure) Init() error {
	return c.puncClient.Init()
}

func (c *pirClientErasure) Read(i int) (Row, error) {
	toReconstruct := make([][]byte, CHUNK_SIZE+ALLOW_LOSS)

	chunkNum := i / CHUNK_SIZE
	goodChunks := 0
	encodedChunkSize := CHUNK_SIZE + ALLOW_LOSS
	toRead := make([]int, encodedChunkSize)
	for j := 0; j < encodedChunkSize; j++ {
		toRead[j] = chunkNum*encodedChunkSize + j
	}

	rows, errs := c.puncClient.ReadBatch(toRead)

	for j := 0; j < encodedChunkSize; j++ {
		if errs[j] == nil {
			toReconstruct[j] = rows[j]
			goodChunks++
		}
	}
	if err := c.rs.Reconstruct(toReconstruct); err != nil {
		return nil, fmt.Errorf("Failed to reconstruct: CHUNK_SIZE: %d, ALLOW_LOSS: %d, goodChunks: %d, numCovered: %d, %w", CHUNK_SIZE, ALLOW_LOSS, goodChunks, c.puncClient.NumCovered(), err)
	}
	return toReconstruct[i%CHUNK_SIZE], nil
}
