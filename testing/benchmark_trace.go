package testing

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"strings"
	"time"

	. "checklist/driver"
	"checklist/pir"
	"checklist/updatable"

	"golang.org/x/crypto/curve25519"
	"gotest.tools/assert"
)

func BenchmarkTrace(config *Config, traceReader io.Reader, stdOut, stdErr io.Writer) error {
	var ep ErrorPrinter
	var numUpdates int

	fmt.Fprintf(stdOut, "%15s%15s%15s%15s%15s%15s%15s%15s\n",
		"Timestamp",
		"NumAdds",
		"NumDeleted",
		"NumQueries",
		"ServerTime[us]", "ClientTime[us]", "CommBytes", "ClientStorage")

	var storageBytes int
	driver, err := config.ServerDriver()
	if err != nil {
		return fmt.Errorf("Failed to create driver: %s\n", err)
	}

	rand := pir.RandSource()

	var trace [][]int = LoadTrace(traceReader)
	config.NumRows = 0

	var none int
	if err := driver.Configure(config.TestConfig, &none); err != nil {
		return fmt.Errorf("Failed to configure driver: %s\n", err)
	}
	driver.ResetMetrics(0, &none)

	client := updatable.NewClient(pir.RandSource(), config.PirType, [2]updatable.UpdatableServer{driver, driver})

	var clientTime, serverTime time.Duration
	var numBytes int
	numUpdates = len(trace) - 1
	fmt.Fprintf(stdErr, "NumUpdates: %d\n", numUpdates)
	for i := 0; i < numUpdates; i++ {
		ts := trace[i][ColumnTimestamp]
		numAdds := trace[i][ColumnAdds]
		numDeletes := trace[i][ColumnDeletes]
		numQueries := trace[i][ColumnQueries]

		if numAdds+numDeletes > 0 {
			assert.NilError(ep, driver.AddRows(numAdds, &none))
			assert.NilError(ep, driver.DeleteRows(numDeletes, &none))

			driver.ResetMetrics(0, &none)

			start := time.Now()
			client.Update()
			clientTime = time.Since(start)

			assert.NilError(ep, driver.GetOfflineTimer(0, &serverTime))
			assert.NilError(ep, driver.GetOfflineBytes(0, &numBytes))

		}

		if numQueries > 0 {
			var rowIV RowIndexVal
			var numKeys int
			assert.NilError(ep, driver.NumKeys(none, &numKeys))
			assert.NilError(ep, driver.GetRow(rand.Intn(numKeys-200), &rowIV))

			driver.ResetMetrics(0, &none)

			start := time.Now()
			row, err := client.Read(rowIV.Key)
			clientTime = time.Since(start)
			assert.NilError(ep, err)
			assert.DeepEqual(ep, row, rowIV.Value)

			assert.NilError(ep, driver.GetOnlineTimer(0, &serverTime))
			assert.NilError(ep, driver.GetOnlineBytes(0, &numBytes))

		}

		if i%200 == 0 {
			storageBytes = client.StorageNumBytes(SerializedSizeOf)
		}

		fmt.Fprintf(stdOut, "%15d%15d%15d%15d%15d%15d%15d%15d\n",
			ts,
			numAdds,
			numDeletes,
			numQueries,
			serverTime.Microseconds(),
			(clientTime - serverTime).Microseconds(),
			numBytes,
			storageBytes)

		if config.Progress {
			fmt.Fprintf(stdErr, "%4d/%-5d\b\b\b\b\b\b\b\b\b\b", i, numUpdates)
		}
	}
	return nil
}

type Printer interface {
	Print(s string)
}

type printerToWriter struct {
	Printer
}

func (p *printerToWriter) Write(b []byte) (n int, err error) {
	p.Print(string(b))
	return len(b), nil
}

const (
	PirTypeDPF        int = int(pir.DPF)
	PirTypePunc       int = int(pir.Punc)
	PirTypeNonPrivate int = int(pir.NonPrivate)
)

func BenchmarkTraceForAndroid(trace string, pirType int, out, err Printer) bool {
	wOut := &printerToWriter{out}
	wErr := &printerToWriter{err}
	fmt.Fprintf(wErr, "Starting benchmark...")
	fmt.Fprintf(wErr, trace[0:30])
	config := new(Config).AddPirFlags().AddClientFlags().AddBenchmarkFlags().Parse()
	config.PirType = pir.PirType(pirType)

	errStr := BenchmarkTrace(config, strings.NewReader(trace), wOut, wErr)
	if errStr != nil {
		fmt.Fprintf(wErr, "Error: %s", errStr)
	}
	return errStr == nil
}

var benchmarkSink byte

func benchmarkTLSHandshake(stdout, stderr io.Writer) error {
	scalar := make([]byte, curve25519.ScalarSize)
	if _, err := rand.Read(scalar); err != nil {
		return fmt.Errorf("failed to read random scalar: %s", err)
	}
	point, err := curve25519.X25519(scalar, curve25519.Basepoint)
	if err != nil {
		return fmt.Errorf("failed to scalar multiply: %s", err)
	}
	if _, err := rand.Read(scalar); err != nil {
		return fmt.Errorf("failed to read random scalar: %s", err)
	}

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	msg := "hello, world"
	hash := sha256.Sum256([]byte(msg))

	sig, err := ecdsa.SignASN1(rand.Reader, privateKey, hash[:])
	if err != nil {
		panic(err)
	}

	start := time.Now()
	for i := 0; i < 1000; i++ {
		out, err := curve25519.X25519(scalar, point)
		if err != nil {
			return fmt.Errorf("failed to scalar multiply: %s", err)
		}
		benchmarkSink ^= out[0]
		out, err = curve25519.X25519(scalar, point)
		if err != nil {
			return fmt.Errorf("failed to scalar multiply: %s", err)
		}

		valid := ecdsa.VerifyASN1(&privateKey.PublicKey, hash[:], sig)
		if !valid {
			return fmt.Errorf("failed to verify signature")
		}

		benchmarkSink ^= out[0]
	}
	fmt.Fprintf(stdout, "Time of 1000 TLS handshakes: %d ms", time.Since(start).Milliseconds())
	return nil
}

func BenchmarkReference(out, err Printer) bool {
	wOut := &printerToWriter{out}
	wErr := &printerToWriter{err}
	fmt.Fprintf(wErr, "Starting benchmark reference...")
	if err := benchmarkTLSHandshake(wOut, wErr); err != nil {
		fmt.Fprintf(wErr, "Error in benchmark: %s", err)
		return false
	}
	return true
}
