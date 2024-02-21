package main

import (
	"crawler/internal/apistats"
	"crawler/internal/cfg"
	"crawler/internal/fetcher"
	"crawler/internal/parser"
	"crawler/internal/storage"
	"fmt"
	bolt "go.etcd.io/bbolt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfgPath := "./configs/config.yaml"

	if len(os.Args) != 2 {
		fmt.Println("Usage: go run crawler.go <starting-url>")
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

	if _, err := os.Stat(appCfg.DownloadsDir); os.IsNotExist(err) {
		err := os.Mkdir(appCfg.DownloadsDir, os.ModePerm)
		if err != nil {
			log.Fatal("Could not create download location", err)
		}
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

	blacklist := storage.NewHashList()

	p := parser.NewParser()
	f := fetcher.NewWebFetcher(appCfg.AcceptableMimeTypes)
	stopChan := make(chan bool)
	crawler := fetcher.NewCrawler(logger, appCfg.Parallelism, p, f, linkRepo, queueRepo, blacklist, appCfg.DownloadsDir)

	apiStats := apistats.NewStatHandler(linkRepo, queueRepo, blacklist)

	http.HandleFunc("/", apiStats.Handler)
	go func() {
		err = http.ListenAndServe(appCfg.ApiAddr, nil)
		if err != nil {
			fmt.Println("Error starting the server:", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigCh
		logger.Printf("Received signal: %s", sig)
		stopChan <- true
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}()

	crawler.Crawl(webResource, stopChan)
}
