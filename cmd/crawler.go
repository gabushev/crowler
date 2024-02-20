package main

import (
	"crawler/internal/cfg"
	"crawler/internal/fetcher"
	"crawler/internal/parser"
	"crawler/internal/storage"
	"fmt"
	bolt "go.etcd.io/bbolt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfgPath := "./configs/config.yaml"

	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <starting-url> <parallelism>")
		return
	}

	webResource := os.Args[1]
	if len(os.Args) == 3 {
		cfgPath = os.Args[2]
	}

	logger := log.New(os.Stdout, "crawler: ", log.LstdFlags|log.Lshortfile)

	appCfg, err := cfg.NewConfig(cfgPath)
	if err != nil {
		log.Fatal("Error reading config:", err)
		return
	}

	db, err := bolt.Open(appCfg.DatabaseFile, 0600, nil)
	if err != nil {
		logger.Fatal("unable to open database:", err)
	}
	defer db.Close()

	linkRepo, err := storage.NewLinkRepository(db)
	if err != nil {
		logger.Fatal("unable to create link repository:", err)
	}
	queueRepo, err := storage.NewQueueRepository(db)
	if err != nil {
		logger.Fatal("unable to create queue repository:", err)
	}
	p := parser.NewParser()
	f := fetcher.WebFetcher{}
	stopChan := make(chan bool)
	crawler := fetcher.NewCrawler(logger, appCfg.Parallelism, p, &f, linkRepo, queueRepo)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigCh
		logger.Printf("Received signal: %s", sig)
		os.Exit(0)
	}()

	crawler.Crawl(webResource, stopChan)
}
