package updatable

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"sync"

	"checklist/pir"
	sb "checklist/safebrowsing"
)

type KeyUpdatesReq struct {
	DefragTimestamp int32
	NextTimestamp   int32
}

type KeyUpdatesResp struct {
	InitialTimestamp int32
	DefragTimestamp  int

	Keys     []uint32
	KeysRice *sb.RiceDeltaEncoding

	//Bit vector
	IsDeletion []byte
	RowLen     int

	ShouldDeleteHistory bool
}
type UpdatableServer interface {
	pir.Server
	KeyUpdates(req KeyUpdatesReq, resp *KeyUpdatesResp) error
}

type Client struct {
	waterfall *WaterfallClient

	servers [2]UpdatableServer

	initialTimestamp int
	defragTimestamp  int
	numRows          int
	rowLen           int
	ops              []dbOp
	keyToPos         map[uint32]int32

	// For testing
	CallAsync           bool
	totalKeyUpdateBytes int
}

func NewClient(source *rand.Rand, pirType pir.PirType, servers [2]UpdatableServer) *Client {
	return &Client{
		waterfall: NewWaterfallClient(source, pirType),
		keyToPos:  make(map[uint32]int32),
		servers:   servers}
}

func (c *Client) Init() error {
	err := c.Update()
	return err
}

func (c *Client) Update() error {
	nextTimestamp := c.nextTimestamp()
	keyReq := KeyUpdatesReq{
		DefragTimestamp: int32(c.defragTimestamp),
		NextTimestamp:   int32(nextTimestamp),
	}
	var keyResp KeyUpdatesResp
	if err := c.servers[pir.Left].KeyUpdates(keyReq, &keyResp); err != nil {
		return err
	}

	numNewRows, err := c.processKeyUpdate(&keyResp)
	if err != nil || numNewRows == 0 {
		return err
	}
	hintReq, err := c.waterfall.HintUpdateReq(numNewRows, keyResp.RowLen)
	if err != nil || hintReq == nil {
		return err
	}
	var hintResp pir.HintResp
	if err := c.servers[pir.Left].Hint(hintReq, &hintResp); err != nil {
		return err
	}
	return c.waterfall.InitHint(hintResp)
}

func (c *Client) Read(key uint32) (pir.Row, error) {
	pos, ok := c.keyToPos[key]
	if !ok {
		return nil, fmt.Errorf("key not found")
	}
	queryReq, reconstructFunc := c.waterfall.Query(int(pos))
	if reconstructFunc == nil {
		return nil, fmt.Errorf("Failed to query: %x", key)
	}
	responses := make([]interface{}, 2)
	errs := make([]error, 2)

	if c.CallAsync {
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			errs[pir.Left] = c.servers[pir.Left].Answer(queryReq[pir.Left], &responses[pir.Left])
		}()
		go func() {
			defer wg.Done()
			errs[pir.Right] = c.servers[pir.Right].Answer(queryReq[pir.Right], &responses[pir.Right])
		}()
		wg.Wait()
	} else {
		errs[pir.Left] = c.servers[pir.Left].Answer(queryReq[pir.Left], &responses[pir.Left])
		errs[pir.Right] = c.servers[pir.Right].Answer(queryReq[pir.Right], &responses[pir.Right])
	}
	if errs[pir.Left] != nil {
		return nil, errs[pir.Left]
	}
	if errs[pir.Right] != nil {
		return nil, errs[pir.Left]
	}

	return reconstructFunc(responses)
}

func (c *Client) Keys() []uint32 {
	keys := make([]uint32, 0, len(c.keyToPos))
	for k := range c.keyToPos {
		keys = append(keys, k)
	}
	return keys
}

func (c *Client) processKeyUpdate(keyResp *KeyUpdatesResp) (numNewRows int, err error) {
	c.rowLen = keyResp.RowLen
	if keyResp.ShouldDeleteHistory {
		c.ops = []dbOp{}
		c.numRows = 0
		c.initialTimestamp = int(keyResp.InitialTimestamp)
		c.defragTimestamp = keyResp.DefragTimestamp
		c.keyToPos = make(map[uint32]int32)
		c.totalKeyUpdateBytes = 0
		c.waterfall.reset()
	}
	var keys []uint32
	if keyResp.KeysRice != nil {
		c.totalKeyUpdateBytes += len(keyResp.KeysRice.EncodedData)
		var err error
		keys, err = DecodeRiceIntegers(keyResp.KeysRice)
		if err != nil {
			return 0, fmt.Errorf("Failed to Rice-decode key updates: %v", err)
		}
	} else {
		keys = keyResp.Keys
		c.totalKeyUpdateBytes += 4*len(keyResp.Keys) + len(keyResp.IsDeletion)
	}

	newOps := make([]dbOp, len(keys))
	for i := range keys {
		isDelete := len(keyResp.IsDeletion) > 0 && (keyResp.IsDeletion[i/8]&(1<<(i%8))) != 0
		newOps[i] = dbOp{
			Key:    keys[i],
			Delete: isDelete,
		}
	}

	if keyResp.DefragTimestamp > c.defragTimestamp {
		defraggedOps, _ := defrag(c.ops, keyResp.DefragTimestamp-c.initialTimestamp)
		c.defragTimestamp = keyResp.DefragTimestamp
		c.initialTimestamp += (len(c.ops) - len(defraggedOps))
		c.ops = append(defraggedOps, newOps...)
		c.numRows = 0
		c.keyToPos = make(map[uint32]int32)
		newOps = c.ops
		c.waterfall.reset()
	} else {
		c.ops = append(c.ops, newOps...)
	}
	numNewRows = 0
	for _, op := range newOps {
		if !op.Delete {
			numNewRows++
		}
	}

	c.updatePositionMap(len(c.ops) - len(newOps))

	return numNewRows, nil
}

func (c *Client) updatePositionMap(fromOpNumber int) {
	for i := fromOpNumber; i < len(c.ops); i++ {
		op := c.ops[i]
		if op.Delete {
			// propagate deletes backwards to previous layers
			delete(c.keyToPos, op.Key)
			continue
		}
		c.keyToPos[op.Key] = int32(c.numRows)
		c.numRows++
	}
}

type compressedMap struct {
	Present []byte
	Keys    []int32
}

func compressPosMap(k2v map[uint32]int32, valRange int) compressedMap {
	v2k := make([]int32, valRange)
	for k, v := range k2v {
		v2k[v] = int32(k)
	}
	present := make([]byte, valRange)
	keys := make([]int32, 0, len(k2v))
	for v, k := range v2k {
		keys = append(keys, k)
		present[v/8] |= (1 << (v % 8))
	}
	return compressedMap{present, keys}
}

func (c *Client) keysSizeWithRice() (int, error) {
	keys := make([]uint32, 0)
	for k := range c.keyToPos {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	rice, err := EncodeRiceIntegers(keys)
	if err != nil {
		return 0, err
	}
	return len(rice.EncodedData), nil
}

func (c *Client) StorageNumBytes(sizeFunc func(interface{}) (int, error)) int {
	bitsPerKey, fixedSize := c.waterfall.State()

	if c.waterfall.pirType != pir.Punc {
		numBytes, err := c.keysSizeWithRice()
		if err != nil {
			log.Fatalf("%s", err)
			return 0
		}
		return numBytes
	}

	// Add key size
	bitsPerKey += 32
	// Add one bit per row for 'is deleted' bitmap
	numBytes := (len(c.keyToPos)*bitsPerKey + c.numRows) / 8
	numBytes += fixedSize
	return numBytes
}

func (c *Client) nextTimestamp() int {
	return c.initialTimestamp + len(c.ops)
}
