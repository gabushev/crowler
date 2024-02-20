package storage

import "sync"

type Hashlist struct {
	urlList map[string]struct{}
	mu      *sync.Mutex
}

func NewHashList() *Hashlist {
	return &Hashlist{
		urlList: make(map[string]struct{}),
		mu:      &sync.Mutex{},
	}
}

func (b *Hashlist) AddToList(val string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.urlList[val] = struct{}{}
}

func (b *Hashlist) RemoveFromList(val string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.urlList, val)
}

func (b *Hashlist) DoesExist(url string) bool {
	if _, ok := b.urlList[url]; ok {
		return true
	}
	return false
}

func (b *Hashlist) Size() int {
	return len(b.urlList)
}
