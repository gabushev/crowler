package functional

import (
	"context"
	"crawler/internal/cfg"
	"crawler/internal/fetcher"
	"crawler/internal/parser"
	"crawler/internal/storage"
	"fmt"
	"github.com/stretchr/testify/suite"
	bolt "go.etcd.io/bbolt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

type ParsingTestSuite struct {
	suite.Suite

	testServer *http.Server

	db        *bolt.DB
	linkRepo  *storage.LinkRepository
	queueRepo *storage.QueueRepository
	crawler   *fetcher.Crawler
}

func (pts *ParsingTestSuite) Test_Crawl_Parsed_Ok() {
	pts.Run("4 files parsed good", func() {
		includeTestFiles([]string{"good_index.html", "included.js", "second_page.html", "main.css"})

		fs := http.FileServer(http.Dir("./staticTest"))
		pts.testServer = &http.Server{
			Addr:    "localhost:8888",
			Handler: fs,
		}

		doneChan := make(chan bool)
		go func() {
			time.Sleep(5 * time.Second)
			cancelCtx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			pts.testServer.Shutdown(cancelCtx)
			doneChan <- true
		}()
		go func() {
			_ = pts.testServer.ListenAndServe()
		}()

		pts.crawler.Crawl("http://localhost:8888/good_index.html", doneChan)

		cnt := 0
		pts.db.View(func(tx *bolt.Tx) error {
			_ = tx.Bucket([]byte("links")).ForEach(func(k, v []byte) error {
				cnt++
				return nil
			})
			return nil
		})

		d, err := pts.linkRepo.GetByKey("http://localhost:8888/good_index.html")
		pts.Assert().Nil(err)
		pts.Assert().NotNil(d)
		pts.Assert().Equal(4, cnt)
	})
}

func (pts *ParsingTestSuite) Test_Crawl_Parsed_With_Errors() {
	pts.Run("bad links parsed correct", func() {
		includeTestFiles([]string{"bad_index.html", "second_page.html", "included.js"})

		fs := http.FileServer(http.Dir("./staticTest"))
		pts.testServer = &http.Server{
			Addr:    "localhost:8888",
			Handler: fs,
		}

		doneChan := make(chan bool)
		go func() {
			time.Sleep(5 * time.Second)
			cancelCtx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			pts.testServer.Shutdown(cancelCtx)
			doneChan <- true
		}()
		go func() {
			_ = pts.testServer.ListenAndServe()
		}()

		pts.crawler.Crawl("http://localhost:8888/bad_index.html", doneChan)

		cnt := 0
		pts.db.View(func(tx *bolt.Tx) error {
			_ = tx.Bucket([]byte("links")).ForEach(func(k, v []byte) error {
				cnt++
				return nil
			})
			return nil
		})

		d, err := pts.linkRepo.GetByKey("http://localhost:8888/bad_index.html")
		pts.Assert().Nil(err)
		pts.Assert().NotNil(d)
		// it is a buggy normal behavior. the FileServer redirects if it has no proper file,
		// and http client does not provide the easy way for handling this situation
		pts.Assert().Equal(3, cnt)
	})

}

func includeTestFiles(filesList []string) {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	fmt.Print(path)

	if _, err := os.Stat("./staticTest/"); os.IsNotExist(err) {
		err := os.Mkdir("./staticTest/", os.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	for i := range filesList {
		sourceFile, err := os.ReadFile("./static/" + filesList[i])
		if err != nil {
			panic(err)
		}
		err = os.WriteFile("./staticTest/"+filesList[i], sourceFile, 0644)
	}
}

func TestParsingTestSuite(t *testing.T) {
	suite.Run(t, new(ParsingTestSuite))
}

func (pts *ParsingTestSuite) SetupTest() {
	var err error
	appCfg := cfg.Config{
		Parallelism: 10,
		AcceptableMimeTypes: []string{
			"text/html",
			"text/css",
			"application/javascript",
			"text/javascript",
		},
		DatabaseFile: "./test.db",
	}
	pts.db, err = bolt.Open(appCfg.DatabaseFile, 0600, nil)
	if err != nil {
		panic(err)
	}

	pts.linkRepo, err = storage.NewLinkRepository(pts.db)
	if err != nil {
		panic(err)
	}
	pts.queueRepo, err = storage.NewQueueRepository(pts.db)
	if err != nil {
		panic(err)
	}
	blacklist := storage.NewHashList()
	p := parser.NewParser()
	f := fetcher.WebFetcher{}
	l := log.New(os.Stdout, "crawler: ", log.LstdFlags|log.Lshortfile)
	pts.crawler = fetcher.NewCrawler(l, appCfg.Parallelism, p, &f, pts.linkRepo, pts.queueRepo, blacklist, "./downloadsTest")
}

func (pts *ParsingTestSuite) TearDownTest() {
	_ = pts.testServer.Shutdown(context.Background())
	pts.db.Close()
	err := os.RemoveAll("./staticTest/")
	if err != nil {
		panic(err)
	}
	err = os.Remove("./test.db")
	if err != nil {
		panic(err)
	}
	err = os.RemoveAll("./downloadsTest/")
	if err != nil {
		panic(err)
	}
}
