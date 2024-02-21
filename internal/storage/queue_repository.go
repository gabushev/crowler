package storage

import (
	"encoding/json"
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
	q := NewQueue([]string{})
	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(queueBucketName))
		if err != nil {
			return err
		}
		currentData := b.Get([]byte(queueData))
		if currentData != nil {
			var restoredData []string
			err = json.Unmarshal(currentData, &restoredData)
			if err != nil {
				return err
			}
			q = NewQueue(restoredData)
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
		qr.queue.Push([]byte(url))
		bucket := tx.Bucket([]byte(queueBucketName))

		return bucket.Put([]byte(url), []byte{})
	})
}

func (qr *QueueRepository) Pull() (string, error) {
	var url []byte
	err := qr.db.Update(func(tx *bolt.Tx) error {
		url = qr.queue.Pull()
		bucket := tx.Bucket([]byte(queueBucketName))
		return bucket.Delete(url)
	})
	return string(url), err
}

func (qr *QueueRepository) Size() int {
	return qr.queue.Size()
}

func (qr *QueueRepository) SaveState() {
	qr.mu.Lock()
	defer qr.mu.Unlock()
	qr.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(queueBucketName))
		data, err := json.Marshal(qr.queue.Data())
		if err != nil {
			return err
		}
		return bucket.Put([]byte(queueData), data)
	})
}
