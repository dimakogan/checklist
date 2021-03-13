package driver

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"testing"

	"checklist/pir"
	"checklist/rpc"
	"checklist/updatable"

	"github.com/ugorji/go/codec"
	"gotest.tools/assert"
)

var config *Config

func runServer() {
	driver, err := NewServerDriver()
	if err != nil {
		log.Fatalf("Failed to create server: %s", err)
	}

	server, err := rpc.NewServer(12345, true, RegisteredTypes())
	if err != nil {
		log.Fatalf("Failed to create server: %s", err)
	}
	if err := server.RegisterName("PirServerDriver", driver); err != nil {
		log.Fatalf("Failed to register PIRServer, %s", err)
	}

	var inShutdown bool
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		inShutdown = true
		server.Close()
	}()

	prof := NewProfiler(config.CpuProfile)
	defer prof.Close()

	go func() {
		err = server.Serve()
		if err != nil && !inShutdown {
			log.Fatalf("Failed to serve: %s", err)
		} else {
			fmt.Printf("Shutting down")
		}
	}()

}

func TestMain(m *testing.M) {
	config = new(Config).AddPirFlags().AddClientFlags().Parse()
	os.Exit(m.Run())
}

func TestPunc(t *testing.T) {
	//runServer()
	driver, err := config.ServerDriver()
	assert.NilError(t, err)

	presetRow := make(pir.Row, 32)
	pir.RandSource().Read(presetRow)

	assert.NilError(t, driver.Configure(TestConfig{
		NumRows:    100000,
		RowLen:     32,
		PresetRows: []RowIndexVal{{7, 0x7, presetRow}},
		PirType:    pir.Punc,
		Updatable:  false,
	}, nil))

	client := pir.NewPIRReader(pir.RandSource(), [2]pir.Server{driver, driver})

	err = client.Init(pir.Punc)
	assert.NilError(t, err)

	// runtime.GC()
	// if memProf, err := os.Create("mem.prof"); err != nil {
	// 	panic(err)
	// } else {
	// 	pprof.WriteHeapProfile(memProf)
	// 	memProf.Close()
	// }
	val, err := client.Read(0x7)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, presetRow)
}

func _TestMessageSizes(t *testing.T) {
	db := pir.MakeDB(3000000, 32)

	resp, err := pir.NewPuncHintReq().Process(db)
	assert.NilError(t, err)
	client := resp.InitClient(pir.RandSource())

	qs, _ := client.Query(7)

	size, err := SerializedSizeOf(qs[0])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stdout, "query size: %d\n", size)
}

func _TestMessageSizesUpdatable(t *testing.T) {
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
func TestPIRServerOverRPC(t *testing.T) {
	driver, err := config.ServerDriver()
	assert.NilError(t, err)

	assert.NilError(t, driver.Configure(TestConfig{
		NumRows:    100000,
		RowLen:     4,
		PresetRows: []RowIndexVal{{7, 0x1234, pir.Row{'C', 'o', 'o', 'l'}}},
		PirType:    pir.Punc,
		Updatable:  true,
	}, nil))

	client := updatable.NewClient(pir.RandSource(), pir.Punc, [2]updatable.UpdatableServer{driver, driver})

	err = client.Init()
	assert.NilError(t, err)

	val, err := client.Read(0x1234)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, pir.Row("Cool"))
}

func TestSerialization(t *testing.T) {
	h := codec.BincHandle{}
	h.StructToArray = true
	h.OptimumSize = true
	h.SetExt(reflect.TypeOf(pir.PuncHintReq{}), 0x12, codec.SelfExt)
	h.PreferPointerForStructOrArray = true
	w := new(bytes.Buffer)
	enc := codec.NewEncoder(w, &h)
	dec := codec.NewDecoder(w, &h)

	req := pir.NewPuncHintReq()
	var reqI pir.HintReq
	reqI = req
	assert.NilError(t, enc.Encode(&reqI))

	var reqOut pir.HintReq
	assert.NilError(t, dec.Decode(&reqOut))
	// var network bytes.Buffer         // Stand-in for a network connection
	// genc := gob.NewEncoder(&network) // Will write to network.
	// gdec := gob.NewDecoder(&network) // Will read from network.
	// assert.NilError(t, genc.Encode(reqI))

	// assert.NilError(t, gdec.Decode(reqOut))
}
