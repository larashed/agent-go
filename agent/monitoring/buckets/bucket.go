package buckets

import (
	"sync"
)

type Items []interface{}

type Bucket struct {
	items Items
	mutex sync.RWMutex
}

func NewBucket() *Bucket {
	return &Bucket{
		items: make(Items, 0),
		mutex: sync.RWMutex{},
	}
}

func NewBucketFromItems(items Items) *Bucket {
	return &Bucket{
		items: items,
		mutex: sync.RWMutex{},
	}
}

func (b *Bucket) Add(record string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.items = append(b.items, record)
}

func (b *Bucket) Items() Items {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.items
}

func (b *Bucket) Count() int {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return len(b.items)
}

func (b *Bucket) PullAndRemove(limit int) *Bucket {
	newBucket := b.GetMany(limit)
	count := newBucket.Count()
	b.Remove(count)

	return newBucket
}

func (b *Bucket) Remove(count int) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	newItems := make(Items, 0)

	b.items = append(newItems, b.items[count:]...)
}

func (b *Bucket) GetMany(limit int) *Bucket {
	count := b.Count()

	b.mutex.Lock()
	defer b.mutex.Unlock()

	if limit > count {
		limit = count
	}

	newItems := make(Items, 0)
	newItems = append(newItems, b.items[:limit]...)

	return NewBucketFromItems(newItems)
}
