package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"checklist/driver"
	"checklist/pir"
	"checklist/updatable"

	"log"
)

type requestTime struct {
	start, end time.Time
}

func main() {
	config := new(driver.Config).AddPirFlags().AddClientFlags()
	latenciesFile := config.FlagSet.String("latenciesFile", "", "Latencies output filename")
	config.Parse()

	proxyLeft, err := config.ServerDriver()
	if err != nil {
		log.Fatal("Connection error: ", err)
	}

	proxyRight, err := config.Server2Driver()
	if err != nil {
		log.Fatal("Connection error: ", err)
	}

	fmt.Printf("Obtaining hint (this may take a while)...")
	client := updatable.NewClient(pir.RandSource(), config.PirType, [2]updatable.UpdatableServer{proxyLeft, proxyRight})
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

	latencies := make(chan (requestTime), 1000)

	go func() {
		cur := 0
		for {
			if inShutdown {
				close(latencies)
				break
			}
			key := keys[rand.Intn(len(keys))]
			start := time.Now()
			_, err := client.Read(key)
			if err != nil {
				fmt.Printf("Failed to read key %d: %v", key, err)
				continue
			}
			latencies <- requestTime{start, time.Now()}
			bar := []string{"|", "/", "-", "\\"}
			fmt.Fprintf(os.Stderr, "\rQuerying: "+bar[cur])
			cur = (cur + 1) % len(bar)
		}
	}()

	var f *os.File
	if *latenciesFile != "" {
		f, err = os.OpenFile(*latenciesFile, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("failed to create output file: %s", err)
		}
		fmt.Fprintf(f, "Seconds,Latency\n")
		defer f.Close()
	}

	for l := range latencies {
		latency := l.end.Sub(l.start).Milliseconds()
		if f != nil {
			fmt.Fprintf(f, "%d,%d\n", l.start.Unix(), latency)
		}
	}
}
