package storage

import bolt "go.etcd.io/bbolt"

type LinkRepository struct {
	db *bolt.DB
}

const linksBucketName = "links"

func NewLinkRepository(db *bolt.DB) (*LinkRepository, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(linksBucketName))
		return err
	})
	if err != nil {
		return nil, err
	}

	return &LinkRepository{db: db}, nil
}

func (lr *LinkRepository) SaveByKey(url string, data []byte) error {
	return lr.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(linksBucketName))

		return bucket.Put([]byte(url), data)
	})
}

func (lr *LinkRepository) GetByKey(url string) ([]byte, error) {
	var data []byte
	err := lr.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(linksBucketName))
		if bucket == nil {
			return nil
		}

		data = bucket.Get([]byte(url))
		return nil
	})
	return data, err
}

func (lr *LinkRepository) IsExists(url string) bool {
	var exists bool
	err := lr.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(linksBucketName))
		if bucket == nil {
			return nil
		}

		exists = bucket.Get([]byte(url)) != nil
		return nil
	})
	if err != nil {
		return false
	}
	return exists
}
