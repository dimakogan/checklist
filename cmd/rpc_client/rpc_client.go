package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	b "github.com/dimakogan/boosted-pir"

	"log"
)

func main() {
	numQueries := flag.Int("q", 10000, "Number of queries to do")
	latenciesFile := flag.String("latenciesFile", "", "Latencies output filename")
	server1Addr := flag.String("s1", "localhost:12345", "server address <HOSTNAME>:<PORT>")
	server2Addr := flag.String("s2", "localhost:12345", "server address <HOSTNAME>:<PORT>")

	flag.Parse()

	latencies := make([]int64, 0)

	fmt.Printf("Connecting to %s...", *server1Addr)
	proxyLeft, err := b.NewPirRpcProxy(*server1Addr)
	if err != nil {
		log.Fatal("Connection error: ", err)
	}

	fmt.Printf("Connecting to %s...", *server2Addr)
	proxyRight, err := b.NewPirRpcProxy(*server2Addr)
	if err != nil {
		log.Fatal("Connection error: ", err)
	}
	fmt.Printf("[OK]\n")

	fmt.Printf("Obtaining hint (this may take a while)...")
	client := b.NewPirClientUpdatable(b.RandSource(), [2]b.PirServer{proxyLeft, proxyRight})
	err = client.Init()
	if err != nil {
		log.Fatalf("Failed to Initialize client: %s\n", err)
	}
	fmt.Printf("[OK]\n")

	keys := client.Keys()
	fmt.Printf("Got %d keys from server\n", len(keys))

	for i := 0; i < *numQueries; i++ {
		key := keys[rand.Intn(len(keys))]
		start := time.Now()
		_, err := client.Read(key)
		if err != nil {
			log.Fatalf("Failed to read key %d: %v", key, err)
		}
		latencies = append(latencies, time.Since(start).Microseconds())
	}

	if len(*latenciesFile) > 0 {
		lOut, _ := os.Create(*latenciesFile)
		for _, l := range latencies {
			lOut.WriteString(fmt.Sprintf("%d\n", l))
		}
		lOut.Close()
	}

	fmt.Printf("Completed %d queries\n", len(latencies))
}
