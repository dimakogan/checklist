package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	b "github.com/dimakogan/boosted-pir"

	"log"
)

func main() {
	config := b.NewConfig().WithClientFlags()
	numQueries := config.FlagSet.Int("q", 10000, "Number of queries to do")
	latenciesFile := config.FlagSet.String("latenciesFile", "", "Latencies output filename")
	config.Parse()

	latencies := make([]int64, 0)

	proxyLeft, err := config.ServerDriver()
	if err != nil {
		log.Fatal("Connection error: ", err)
	}

	proxyRight, err := config.Server2Driver()
	if err != nil {
		log.Fatal("Connection error: ", err)
	}

	fmt.Printf("Obtaining hint (this may take a while)...")
	client := b.NewPirClientUpdatable(b.RandSource(), config.PirType, [2]b.PirUpdatableServer{proxyLeft, proxyRight})
	client.CallAsync = true
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
