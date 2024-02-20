package storage

import (
	bolt "go.etcd.io/bbolt"
	"sync"
)

type QueueRepository struct {
	db    *bolt.DB
	queue *Queue
	mu    sync.Mutex
}

const queueBucketName = "queue"
const queueData = "data"

func NewQueueRepository(db *bolt.DB) (*QueueRepository, error) {
	q := NewQueue([]interface{}{})
	err := db.Update(func(tx *bolt.Tx) error {

		b, err := tx.CreateBucketIfNotExists([]byte(queueBucketName))
		if err != nil {
			return err
		}
		currentData := b.Get([]byte(queueData))
		if currentData != nil {
			q = NewQueue([]interface{}{})
		}
		return err
	})
	if err != nil {
		return nil, err
	}

	return &QueueRepository{db: db, queue: q}, nil
}

func (qr *QueueRepository) Push(url string) error {
	return qr.db.Update(func(tx *bolt.Tx) error {
		qr.queue.Push(url)
		bucket := tx.Bucket([]byte(queueBucketName))

		return bucket.Put([]byte(url), []byte{})
	})
}

func (qr *QueueRepository) Pull() (string, error) {
	var url string
	err := qr.db.Update(func(tx *bolt.Tx) error {
		url = qr.queue.Pull().(string)
		bucket := tx.Bucket([]byte(queueBucketName))
		return bucket.Delete([]byte(url))
	})
	return url, err
}

func (qr *QueueRepository) Size() int {
	return qr.queue.Size()
}
