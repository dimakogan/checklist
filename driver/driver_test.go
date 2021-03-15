package driver

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"testing"

	"checklist/pir"
	"checklist/rpc"
	"checklist/safebrowsing"
	"checklist/updatable"

	"github.com/ugorji/go/codec"
	"gotest.tools/assert"
)

var config *Config

func TestMain(m *testing.M) {
	config = new(Config).AddPirFlags().AddClientFlags().Parse()
	os.Exit(m.Run())
}

func TestStatic(t *testing.T) {
	driverL, err := config.ServerDriver()
	assert.NilError(t, err)
	driverR, err := config.Server2Driver()
	assert.NilError(t, err)

	presetRow := make(pir.Row, config.RowLen)
	pir.RandSource().Read(presetRow)

	config.PresetRows = []RowIndexVal{{7, 0x1234, presetRow}}
	config.Updatable = false
	config.DataRandSeed = 13

	var none int
	err = driverL.Configure(config.TestConfig, &none)
	assert.NilError(t, err)
	err = driverR.Configure(config.TestConfig, &none)
	assert.NilError(t, err)

	client := pir.NewPIRReader(pir.RandSource(), pir.Server(driverL), pir.Server(driverR))

	err = client.Init(config.PirType)
	assert.NilError(t, err)

	val, err := client.Read(7)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, presetRow)
}

func TestUpdatable(t *testing.T) {
	driverL, err := config.ServerDriver()
	assert.NilError(t, err)
	driverR, err := config.Server2Driver()
	assert.NilError(t, err)

	presetRow := make(pir.Row, config.RowLen)
	pir.RandSource().Read(presetRow)

	config.PresetRows = []RowIndexVal{{7, 0x1234, presetRow}}
	config.Updatable = true
	config.DataRandSeed = 13

	var none int
	err = driverL.Configure(config.TestConfig, &none)
	assert.NilError(t, err)
	err = driverR.Configure(config.TestConfig, &none)
	assert.NilError(t, err)

	client := updatable.NewClient(pir.RandSource(), config.PirType, [2]updatable.UpdatableServer{driverL, driverR})
	err = client.Init()
	assert.NilError(t, err)

	val, err := client.Read(0x1234)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, presetRow)

	val, err = client.Read(0x1234)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, presetRow)
}

func TestSafeBrowsingList(t *testing.T) {
	blFile := "../safebrowsing/evil_urls.txt"
	file, err := os.Open(blFile)
	if err != nil {
		log.Fatalf("Failed to open block list file %s: %s", blFile, err)
	}
	partial, full, err := safebrowsing.ReadBlockedURLs(file)
	if err != nil || len(full) < 0 {
		log.Fatalf("Failed to read blocked urls from %s: %s", blFile, err)
	}
	if len(partial) != len(full) {
		log.Fatalf("Invalid number of partial %d and full %d", len(partial), len(full))
	}
	config.NumRows = len(partial)
	config.RowLen = len(full[0])
	for i := range partial {
		entry := RowIndexVal{
			Index: i,
			Key:   partial[i],
			Value: full[i],
		}
		config.PresetRows = append(config.PresetRows, entry)
	}
	none := 0

	driverL, err := config.ServerDriver()
	assert.NilError(t, err)
	assert.NilError(t, driverL.Configure(config.TestConfig, &none))
	driverR, err := config.Server2Driver()
	assert.NilError(t, err)
	assert.NilError(t, driverR.Configure(config.TestConfig, &none))

	client := updatable.NewClient(pir.RandSource(), config.PirType, [2]updatable.UpdatableServer{driverL, driverR})
	err = client.Init()
	assert.NilError(t, err)

	for i := range partial {
		val, err := client.Read(partial[i])
		assert.NilError(t, err)
		assert.DeepEqual(t, val, pir.Row(full[i]))
	}
}

func TestSerialization(t *testing.T) {
	h := rpc.CodecHandle(RegisteredTypes())

	w := new(bytes.Buffer)
	enc := codec.NewEncoder(w, h)
	dec := codec.NewDecoder(w, h)

	req := pir.NewPuncHintReq(pir.RandSource())
	var reqI pir.HintReq = req
	assert.NilError(t, enc.Encode(&reqI))

	var reqOut pir.HintReq
	assert.NilError(t, dec.Decode(&reqOut))
}

func _testMessageSizes(t *testing.T) {
	db := pir.MakeDB(3000000, 32)

	resp, err := pir.NewPuncHintReq(pir.RandSource()).Process(db)
	assert.NilError(t, err)
	client := resp.InitClient(pir.RandSource())

	qs, _ := client.Query(7)

	size, err := SerializedSizeOf(qs[0])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stdout, "query size: %d\n", size)
}

func _testMessageSizesUpdatable(t *testing.T) {
	numRows := 4000000
	initialSize := 3000000
	db := pir.MakeDB(numRows, 32)

	client := updatable.NewWaterfallClient(pir.RandSource(), pir.Punc)

	for i := 0; i < 50; i++ {
		updateSize := 0
		if i == 0 {
			updateSize = initialSize
		} else {
			updateSize = 200
		}
		req, err := client.HintUpdateReq(updateSize, 32)
		assert.NilError(t, err)
		if req != nil {
			resp, err := req.Process(db)
			assert.NilError(t, err)
			assert.NilError(t, client.InitHint(resp))
		}

		qs, _ := client.Query(7)
		size, err := SerializedSizeOf(qs[0])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(os.Stdout, "updatable request size: %d\n", size)
	}
}
