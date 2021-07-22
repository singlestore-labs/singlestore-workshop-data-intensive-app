package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"src"
)

func main() {
	// global configuration
	rand.Seed(time.Now().UnixNano())
	log.SetFlags(log.Ldate | log.Ltime)

	// handle command line flags
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to an optional config file")
	flag.Parse()

	// load configuration file if it exists
	var config *src.SimConfig
	if configPath != "" {
		conf, err := src.NewSimConfigFromFile(configPath)
		if err != nil {
			log.Fatal(err)
		}
		config = conf
	}

	sitemap, err := src.LoadSitemap(config.SitemapURL)
	if err != nil {
		log.Fatal(err)
	}

	simulators := make([]*src.Simulator, 0)
	stopAll := func() {
		for _, sim := range simulators {
			sim.Stop()
		}
	}

	// Trap SIGINT to trigger a clean shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signals
		log.Printf("received shutdown signal: %s", sig)
		stopAll()
	}()

	numWorkers := runtime.NumCPU()
	if config.NumWorkers != 0 {
		numWorkers = config.NumWorkers
	}

	log.Printf("starting simulation with %d workers", numWorkers)

	wg := sync.WaitGroup{}
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)

		simulator, err := src.NewSimulator(fmt.Sprintf("sim-%d", i), config, sitemap)
		if err != nil {
			log.Fatal(err)
		}

		simulators = append(simulators, simulator)

		go func(i int) {
			defer wg.Done()
			err := simulator.Run()
			if err != nil {
				log.Fatalf("sim-%d failed: %s", i, err)
			}
		}(i)
	}

	wg.Wait()
}
