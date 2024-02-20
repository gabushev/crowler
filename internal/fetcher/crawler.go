package fetcher

import (
	"fmt"
	"log"
	"net/url"
	"sync"
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
	Pull() (string, error)
	Size() int
}

type Blacklist struct {
	urlList map[string]struct{}
	mu      *sync.Mutex
}

func (b *Blacklist) AddToList(url string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.urlList[url] = struct{}{}
}

func (b *Blacklist) DoesExist(url string) bool {
	if _, ok := b.urlList[url]; ok {
		return true
	}
	return false
}

type Crawler struct {
	logger      *log.Logger
	parallelism int
	parser      Parser
	fetcher     Fetcher
	linkRepo    StorageRepository
	queue       QueueInterface
	blacklist   Blacklist
}

func NewCrawler(
	log *log.Logger,
	parallelism int,
	p Parser,
	f Fetcher,
	l StorageRepository,
	q QueueInterface,
) *Crawler {
	bl := Blacklist{
		urlList: make(map[string]struct{}),
		mu:      &sync.Mutex{},
	}
	return &Crawler{
		logger:      log,
		parallelism: parallelism,
		parser:      p,
		fetcher:     f,
		linkRepo:    l,
		queue:       q,
		blacklist:   bl,
	}
}

type FetchTask struct {
	Link string
}

func (c *Crawler) JobProducer(linksChan chan *FetchTask) {
	go func() {
		for {
			if c.queue.Size() > 0 {
				item, err := c.queue.Pull()
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
		return nil, nil, fmt.Errorf("unable to parse links from the page %s")
	}

	return c.filterLinks(urlString, links), body, nil
}

func (c *Crawler) filterLinks(originalLink string, links []string) []string {
	var filteredLinks []string
	original, _ := url.Parse(originalLink)

	for i := range links {
		l, err := url.Parse(links[i])
		if err != nil {
			c.blacklist.AddToList(links[i])
			continue
		}
		if "" == l.Hostname() {
			l.Host = original.Host
		} else {
			if l.Hostname() != original.Hostname() {
				c.blacklist.AddToList(l.String())
				continue
			}
		}
		if "" == l.Scheme {
			l.Scheme = original.Scheme
		} else {
			if l.Scheme != original.Scheme {
				c.blacklist.AddToList(l.String())
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

	// size = 0 means there is no postponed work and probably it is the first run
	if c.queue.Size() == 0 {
		err = c.queue.Push(urlString)
		if err != nil {
			c.logger.Println("Cannot push link to the queue, err: ", err)
		}
	}

	go c.JobProducer(linkBuf)
	go func() {
		for {
			select {
			case link := <-linkBuf:
				if !c.isValidLink(domain, link.Link) {
					c.blacklist.AddToList(link.Link)
					break
				}
				newLinks, pageData, err := c.ExecuteLink(link.Link)
				c.logger.Println(fmt.Sprintf("DEBUG: got new links, %d", len(newLinks)))
				if err != nil {
					c.blacklist.AddToList(link.Link)
				} else {
					err = c.linkRepo.SaveByKey(link.Link, pageData)
					if err != nil {
						c.logger.Println("Cannot save link by key, err: ", err)
					}
				}

				for i := range newLinks {
					if c.linkRepo.IsExists(newLinks[i]) || c.blacklist.DoesExist(newLinks[i]) {
						continue
					}
					err = c.queue.Push(newLinks[i])
					if err != nil {
						c.logger.Println("Cannot push link to the queue, err: ", err)
					}
				}

			case <-exitChan:
				// there might be any stop&exit procedures but we already have something persistent-like
			}
		}
	}()

	<-exitChan
}

func (c *Crawler) isValidLink(domain string, link string) bool {
	u, err := url.Parse(link)
	if err != nil {
		return false
	}
	if u.Hostname() != domain {
		return false
	}

	return true
}
