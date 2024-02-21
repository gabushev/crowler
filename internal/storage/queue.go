package storage

import "sync"

type Queue struct {
	items []string
	mutex sync.Mutex
}

func NewQueue(initData []string) *Queue {
	return &Queue{items: initData}
}

func (q *Queue) Push(item []byte) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.items = append(q.items, string(item))
}

func (q *Queue) Pull() []byte {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.items) == 0 {
		return nil
	}

	item := q.items[0]
	q.items = q.items[1:]
	return []byte(item)
}

func (q *Queue) Size() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return len(q.items)
}

func (q *Queue) Data() []string {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return q.items
}
