package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sync/atomic"

	. "checklist/driver"
	"checklist/pir"
	"checklist/updatable"
)

type userLoadGen struct {
	hintReqs   []*updatable.UpdatableHintReq
	keyUpdates []int
	numQueries int
	answerGen  *answerLoadGen

	requestRate                         int
	updatesDone, queriesDone, hintsDone uint64
}

func initUserLoadGen(config *Config, trace [][]int) *userLoadGen {
	waterfallClient := updatable.NewWaterfallClient(pir.RandSource(), config.PirType)
	config.NumRows = 0
	hintReqs := make([]*updatable.UpdatableHintReq, 0)
	numQueries := 0
	keyUpdates := make([]int, 0)
	for i, row := range trace {
		fmt.Fprintf(os.Stderr, "Generating requests: %4d/%-5d\r", i, len(trace))
		if row[ColumnAdds] > 0 {
			keyUpdates = append(keyUpdates, row[ColumnAdds])
			config.NumRows += row[ColumnAdds]
			hintReq, err := waterfallClient.HintUpdateReq(row[ColumnAdds], config.RowLen)
			if err != nil {
				log.Fatalf("Failed to generate HintReq for timestamp %d: %s", row[ColumnTimestamp], err)
			}
			if hintReq != nil {
				hintReqs = append(hintReqs, hintReq)
			}
		} else {
			numQueries++
			if config.PirType != pir.NonPrivate {
				numQueries++
			}
		}
	}
	fmt.Printf("Generated %d hint requests [OK]\n", len(hintReqs))
	reqRate := (trace[len(trace)-1][ColumnTimestamp] - trace[0][ColumnTimestamp]) / (len(keyUpdates) + numQueries)
	fmt.Printf("1 Req/Sec = %d Users\n", reqRate)
	return &userLoadGen{hintReqs, keyUpdates, numQueries, initAnswerLoadGen(config), reqRate, 0, 0, 0}
}

func (gen *userLoadGen) request(proxy *RpcProxy) error {
	r := rand.Intn(len(gen.keyUpdates) + gen.numQueries)
	if r < len(gen.keyUpdates) {
		updateSize := gen.keyUpdates[r]
		keyReq := updatable.KeyUpdatesReq{
			DefragTimestamp: math.MaxInt32,
			NextTimestamp:   int32(gen.answerGen.numRows - updateSize),
		}
		var keyResp updatable.KeyUpdatesResp
		err := proxy.KeyUpdates(keyReq, &keyResp)
		if err != nil {
			return fmt.Errorf("Failed to replay key update request %v, %s", keyReq, err)
		}
		// int(keyResp.KeysRice.NumEntries)+1 != updateSize
		if len(keyResp.Keys) != updateSize {
			panic(fmt.Sprintf("Invalid size of key update, expected: %d, got: %d", updateSize, len(keyResp.Keys)))
		}
		atomic.AddUint64(&gen.updatesDone, 1)
		if r < len(gen.hintReqs) {
			hintReq := gen.hintReqs[r]
			var hintResp pir.HintResp
			err := proxy.Hint(hintReq, &hintResp)
			if err != nil {
				return fmt.Errorf("Failed to replay hint request %v, %s", hintReq, err)
			}
			if hintResp.NumRows() != gen.hintReqs[r].NumRows {
				return fmt.Errorf("Failed to replay hint request %v , mismatching hint num rows, expected: %d, got: %d", hintReq, hintReq.NumRows, hintResp.NumRows())
			}
			atomic.AddUint64(&gen.hintsDone, 1)
		}
		return nil
	}
	err := gen.answerGen.request(proxy)
	if err == nil {
		atomic.AddUint64(&gen.queriesDone, 1)
	}
	return err
}

func (gen *userLoadGen) debugStr() string {
	return fmt.Sprintf("%d,%d,%d", gen.updatesDone, gen.hintsDone, gen.queriesDone)
}

func (gen *userLoadGen) reqRate() int {
	return gen.requestRate
}
