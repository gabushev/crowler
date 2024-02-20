package fetcher

import (
	"fmt"
	"log"
	"net/url"
	"time"
)

type Parser interface {
	ParseLinks(pageData []byte) ([]string, error)
}

type Fetcher interface {
	Download(urlString string) ([]byte, error)
}

type StorageRepository interface {
	SaveByKey(url string, data []byte) error
	GetByKey(url string) ([]byte, error)
	IsExists(url string) bool
}

type QueueInterface interface {
	Push(url string) error
	Put() (string, error)
	Size() int
}

type Crawler struct {
	logger      *log.Logger
	parallelism int
	parser      Parser
	fetcher     Fetcher
	linkRepo    StorageRepository
	queue       QueueInterface
}

func NewCrawler(
	log *log.Logger,
	parallelism int,
	p Parser,
	f Fetcher,
	l StorageRepository,
	q QueueInterface,
) *Crawler {
	return &Crawler{
		logger:      log,
		parallelism: parallelism,
		parser:      p,
		fetcher:     f,
		linkRepo:    l,
		queue:       q,
	}
}

func (c *Crawler) Stop() {}

type FetchTask struct {
	Link string
}

func (c *Crawler) JobProducer(linksChan chan *FetchTask) {
	go func() {
		for {
			if c.queue.Size() > 0 {
				item, err := c.queue.Put()
				if err != nil {
					c.logger.Println("error during the pulling the next item from the queue, err: ", err)
					return
				}
				linksChan <- &FetchTask{Link: item}
			} else {
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()
}

func (c *Crawler) ExecuteLink(urlString string) ([]string, []byte, error) {
	_, err := url.Parse(urlString)
	if err != nil {
		c.logger.Printf("Invalid URL, parsing error: %s", err)
		return nil, nil, fmt.Errorf("invalid URL, parsing error: %w", err)
	}

	body, err := c.fetcher.Download(urlString)
	if err != nil {
		return nil, nil, fmt.Errorf("download error for url %s, %err", urlString, err)
	}

	links, err := c.parser.ParseLinks(body)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse links from the page")
	}

	return c.filterLinks(urlString, links), body, nil
}

func (c *Crawler) filterLinks(originalLink string, links []string) []string {
	var filteredLinks []string
	original, _ := url.Parse(originalLink)

	for i := range links {
		l, err := url.Parse(links[i])
		if err != nil {
			continue
		}
		if "" == l.Hostname() {
			l.Host = original.Host
		} else {
			if l.Hostname() != original.Hostname() {
				continue
			}
		}
		if "" == l.Scheme {
			l.Scheme = original.Scheme
		} else {
			if l.Scheme != original.Scheme {
				continue
			}
		}

		filteredLinks = append(filteredLinks, l.String())
	}

	return filteredLinks
}

func (c *Crawler) Crawl(urlString string, exitChan chan bool) {
	u, err := url.Parse(urlString)
	if err != nil {
		c.logger.Printf("Invalid URL, parsing error: %s\n", err)
		return
	}
	domain := u.Hostname()

	linkBuf := make(chan *FetchTask, c.parallelism)

	err = c.queue.Push(urlString)
	if err != nil {
		c.logger.Println("Cannot push link to the queue, err: ", err)
	}
	c.logger.Println("start serving")
	go c.JobProducer(linkBuf)
	go func() {
		for {
			select {
			case link := <-linkBuf:
				c.logger.Println("recieved ling " + link.Link)
				if !c.isValidLink(domain, link.Link) {
					// mark page as bad source and dont use anymore
				}
				newLinks, pageData, err := c.ExecuteLink(link.Link)
				c.logger.Println(fmt.Sprintf("got new links, %d", len(newLinks)))
				if err != nil {
					// mark page as bad source and dont check it anymore
				}
				err = c.linkRepo.SaveByKey(link.Link, pageData)
				if err != nil {
					c.logger.Println("Cannot save link by key, err: ", err)
				}
				for i := range newLinks {
					c.logger.Println("Adding one more link ", newLinks[i])
					if c.linkRepo.IsExists(newLinks[i]) {
						continue
					}
					err = c.queue.Push(newLinks[i])
					if err != nil {
						c.logger.Println("Cannot push link to the queue, err: ", err)
					}
				}

			case <-exitChan:

			}
		}
	}()

	<-exitChan
}

func (c *Crawler) isValidLink(domain string, link string) bool {
	// add blacklist
	u, err := url.Parse(link)
	if err != nil {
		return false
	}
	if u.Hostname() != domain {
		return false
	}

	return true
}
