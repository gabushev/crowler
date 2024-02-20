package storage

import "sync"

type Queue struct {
	items []interface{}
	mutex sync.Mutex
}

func NewQueue(initData []interface{}) *Queue {
	return &Queue{items: initData}
}

func (q *Queue) Push(item interface{}) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.items = append(q.items, item)
}

func (q *Queue) Put() interface{} {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.items) == 0 {
		return nil
	}

	item := q.items[0]
	q.items = q.items[1:]
	return item
}

func (q *Queue) Size() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return len(q.items)
}

func (q *Queue) Data() []interface{} {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return q.items
}
