package boosted

import (
	"fmt"
	"log"
	"math"
	"math/rand"

	"github.com/klauspost/reedsolomon"
)

type pirClientErasure struct {
	// RS params
	chunkSize int
	allowLoss int

	rs         reedsolomon.Encoder
	puncClient *pirClientPunc
}

type pirServerErasure struct {
	pirServerPunc
}

var SEC_PARAM = 128

var DEFAULT_CHUNK_SIZE = 50

var NUM_HINTS_MULTIPLIER = 1

func prAtLeastChernoff(p float64, n int, k int) float64 {
	a := float64(k)/float64(n) - p
	return math.Pow(math.Pow(p/(p+a), p+a)*math.Pow((1-p)/(1-p-a), 1-p-a), float64(n))
}

func computeAllowedLoss(chunkSize int, numHintsMultiplier int) int {
	errBits := 0
	q := chunkSize
	// error prob
	p := math.Exp(-float64(numHintsMultiplier))
	// Starting overall block size for search
	B := int(float64(q) / (1 - p))
	for errBits < SEC_PARAM {
		B += 1
		// upper bound on the error: Pr[#Erasures >= B - q + 1]
		errBits = int(-math.Log2(prAtLeastChernoff(p, B, B-q+1)))
		if B > SEC_PARAM*q {
			log.Fatalf("Something wierd, block size B=lam*q is not enough B=%d,q=%d, errBits: %d", B, q, errBits)
		}
	}
	return B - q
}

func nEncodedRows(nRows, chunkSize, allowLoss int) int {
	// We split the length-n database into chunks of size `chunkSize`
	// and then we encode each chunk using a erasure encoding that
	// tolerates the loss of at most L rows.

	// XXX choose `chunkSize` and `allowLoss` such that if you flip `chunkSize`
	// coins that come up heads with probability 1/e, you end up with
	// fewer than `allowLoss` heads with very high probability.

	// Our encoding library requires CHUNK_SIZE + ALLOW_LOSS <= 256

	return (((nRows - 1) / chunkSize) + 1) * (chunkSize + allowLoss)
}

func encodeDatabase(data []Row, chunkSize int, allowLoss int) ([]Row, error) {

	enc, err := reedsolomon.New(chunkSize, allowLoss)
	if err != nil {
		return nil, fmt.Errorf("Could not create encoder: %s", err)
	}

	if len(data)%chunkSize != 0 {
		paddingLen := chunkSize - (len(data)-1)%chunkSize - 1
		data = append(data, data[0:paddingLen]...)
		//return nil, fmt.Errorf("DB length: %d is not multiple of CHUNK_SIZE: %d", len(data), CHUNK_SIZE)
	}

	encRows := nEncodedRows(len(data), chunkSize, allowLoss)
	encoded := make([]Row, encRows)
	rowLen := len(data[0])

	for i := 0; i < len(data)/chunkSize; i++ {
		toEnc := make([][]byte, chunkSize+allowLoss)

		// Data chunks
		for j := 0; j < chunkSize; j++ {
			toEnc[j] = data[i*chunkSize+j]
		}

		// Parity chunks
		for j := 0; j < allowLoss; j++ {
			toEnc[chunkSize+j] = make([]byte, rowLen)
		}

		err := enc.Encode(toEnc)
		if err != nil {
			panic("Encoding error")
		}

		for j := 0; j < chunkSize+allowLoss; j++ {
			// fmt.Printf("Copying %v\n", i*CHUNK_SIZE+j)
			encoded[i*(chunkSize+allowLoss)+j] = toEnc[j]
		}
	}

	return encoded, nil
}

func NewPirServerErasure(source *rand.Rand, data []Row, chunkSize int) (PirServer, error) {
	allowLoss := computeAllowedLoss(chunkSize, NUM_HINTS_MULTIPLIER)
	encdata, err := encodeDatabase(data, chunkSize, allowLoss)
	if err != nil {
		return nil, err
	}
	// fmt.Printf("LenIn = %v\n", len(data))
	// fmt.Printf("LenOut = %v\n", len(encdata))
	return pirServerErasure{NewPirServerPunc(source, encdata)}, nil
}

func NewPirClientErasure(source *rand.Rand, nRows int, chunkSize int, servers [2]PirServer) (*pirClientErasure, error) {
	allowLoss := computeAllowedLoss(chunkSize, NUM_HINTS_MULTIPLIER)
	nEnc := nEncodedRows(nRows, chunkSize, allowLoss)
	rs, err := reedsolomon.New(chunkSize, allowLoss)
	if err != nil {
		return nil, fmt.Errorf("Could not create RS encoder: %w", err)
	}

	puncClient := NewPirClientPunc(source, nEnc, servers)
	puncClient.nHints = int(math.Round(math.Pow(float64(puncClient.nRows), 0.5))) * NUM_HINTS_MULTIPLIER
	return &pirClientErasure{chunkSize: chunkSize, allowLoss: allowLoss, rs: rs, puncClient: puncClient}, nil
}

func (c *pirClientErasure) Init() error {
	return c.puncClient.Init()
}

func (c *pirClientErasure) Read(i int) (Row, error) {
	toReconstruct := make([][]byte, c.chunkSize+c.allowLoss)

	chunkNum := i / c.chunkSize
	goodChunks := 0
	encodedChunkSize := c.chunkSize + c.allowLoss
	toRead := make([]int, encodedChunkSize)
	for j := 0; j < encodedChunkSize; j++ {
		toRead[j] = chunkNum*encodedChunkSize + j
	}

	rows, errs := c.puncClient.ReadBatchAtLeast(toRead, c.chunkSize)

	for j := 0; j < encodedChunkSize; j++ {
		if errs[j] == nil {
			toReconstruct[j] = rows[j]
			goodChunks++
		}
	}
	if err := c.rs.Reconstruct(toReconstruct); err != nil {
		return nil, fmt.Errorf("Failed to reconstruct: CHUNK_SIZE: %d, ALLOW_LOSS: %d, goodChunks: %d, numCovered: %d, %w", c.chunkSize, c.allowLoss, goodChunks, c.puncClient.NumCovered(), err)
	}
	return toReconstruct[i%c.chunkSize], nil
}
