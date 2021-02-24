package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	b "github.com/dimakogan/boosted-pir"

	"log"
)

type requestTime struct {
	start, end time.Time
}

func main() {
	config := b.NewConfig().WithClientFlags()
	latenciesFile := config.FlagSet.String("latenciesFile", "", "Latencies output filename")
	config.Parse()

	latencies := make([]requestTime, 0)

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

	inShutdown := false

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		inShutdown = true
	}()

	for {
		if inShutdown {
			break
		}
		key := keys[rand.Intn(len(keys))]
		start := time.Now()
		_, err := client.Read(key)
		if err != nil {
			log.Fatalf("Failed to read key %d: %v", key, err)
		}
		latencies = append(latencies, requestTime{start, time.Now()})
	}
	if *latenciesFile != "" {
		log := "Time,Latency\n"
		for _, l := range latencies {
			log += fmt.Sprintf("%d,%d\n", l.start.Unix(), l.end.Sub(l.start).Milliseconds())
		}
		ioutil.WriteFile(*latenciesFile, []byte(log), 0644)
	}

	fmt.Printf("Completed %d queries\n", len(latencies))
}
