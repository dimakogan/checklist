package boosted

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"sync"
)

type PirClientUpdatable struct {
	waterfall *PirClientWaterfall

	servers [2]PirUpdatableServer

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

func NewPirClientUpdatable(source *rand.Rand, pirType PirType, servers [2]PirUpdatableServer) *PirClientUpdatable {
	return &PirClientUpdatable{
		waterfall: NewPirClientWaterfall(source, pirType),
		keyToPos:  make(map[uint32]int32),
		servers:   servers}
}

func (c *PirClientUpdatable) Init() error {
	err := c.Update()
	return err
}

func (c *PirClientUpdatable) Update() error {
	nextTimestamp := c.nextTimestamp()
	keyReq := KeyUpdatesReq{
		DefragTimestamp: int32(c.defragTimestamp),
		NextTimestamp:   int32(nextTimestamp),
	}
	var keyResp KeyUpdatesResp
	if err := c.servers[Left].KeyUpdates(keyReq, &keyResp); err != nil {
		return err
	}

	numNewRows, err := c.processKeyUpdate(&keyResp)
	if err != nil || numNewRows == 0 {
		return err
	}
	hintReq, err := c.waterfall.HintUpdateReq(numNewRows)
	if err != nil || hintReq == nil {
		return err
	}
	var hintResp HintResp
	if err := c.servers[Left].Hint(*hintReq, &hintResp); err != nil {
		return err
	}
	return c.waterfall.initHint(&hintResp)
}

func (c *PirClientUpdatable) Read(key uint32) (Row, error) {
	pos, ok := c.keyToPos[key]
	if !ok {
		return nil, fmt.Errorf("key not found")
	}
	queryReq, reconstructFunc := c.waterfall.Query(int(pos))
	if reconstructFunc == nil {
		return nil, fmt.Errorf("Failed to query: %x", key)
	}
	responses := make([]QueryResp, 2)
	errs := make([]error, 2)

	if c.CallAsync {
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			errs[Left] = c.servers[Left].Answer(queryReq[Left], &responses[Left])
		}()
		go func() {
			defer wg.Done()
			errs[Right] = c.servers[Right].Answer(queryReq[Right], &responses[Right])
		}()
		wg.Wait()
	} else {
		errs[Left] = c.servers[Left].Answer(queryReq[Left], &responses[Left])
		errs[Right] = c.servers[Right].Answer(queryReq[Right], &responses[Right])
	}
	if errs[Left] != nil {
		return nil, errs[Left]
	}
	if errs[Right] != nil {
		return nil, errs[Left]
	}

	return reconstructFunc(responses)
}

func (c *PirClientUpdatable) Keys() []uint32 {
	keys := make([]uint32, 0, len(c.keyToPos))
	for k := range c.keyToPos {
		keys = append(keys, k)
	}
	return keys
}

func (c *PirClientUpdatable) processKeyUpdate(keyResp *KeyUpdatesResp) (numNewRows int, err error) {
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

func (c *PirClientUpdatable) updatePositionMap(fromOpNumber int) {
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

func (c *PirClientUpdatable) keysSizeWithRice() (int, error) {
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

func (c *PirClientUpdatable) StorageNumBytes() int {
	numBytes := c.waterfall.StorageNumBytes()

	var keysBytes int
	var err error
	if c.waterfall.pirType != Punc {
		keysBytes, err = c.keysSizeWithRice()
	} else {
		keysBytes, err = SerializedSizeOf(compressPosMap(c.keyToPos, c.numRows))
		if c.totalKeyUpdateBytes < keysBytes {
			keysBytes = c.totalKeyUpdateBytes
		}
	}
	if err != nil {
		log.Fatalf("%s", err)
		return 0
	}
	numBytes += keysBytes

	return numBytes
}

func (c *PirClientUpdatable) nextTimestamp() int {
	return c.initialTimestamp + len(c.ops)
}
