package buckets

import (
	"strings"
	"sync"
)

type Item interface {
	String() string
	Value() interface{}
}

type Items []Item

type Bucket struct {
	items   Items
	mutex   sync.RWMutex
	Channel chan int
}

func NewBucket() *Bucket {
	return &Bucket{
		items:   make(Items, 0),
		mutex:   sync.RWMutex{},
		Channel: make(chan int),
	}
}

func NewBucketFromItems(items Items) *Bucket {
	return &Bucket{
		items: items,
		mutex: sync.RWMutex{},
	}
}

func (b *Bucket) Merge(bucket *Bucket) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.items = append(b.items, *bucket.GetAll()...)
}

func (b *Bucket) Add(record Item) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	go func() {
		b.Channel <- 1
	}()

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

	b.Remove(newBucket.Count())

	return newBucket
}

func (b *Bucket) Remove(count int) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	newItems := make(Items, 0)

	b.items = append(newItems, b.items[count:]...)
}

func (b *Bucket) String() string {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	var strs []string

	for i := 0; i < len(b.items); i++ {
		strs = append(strs, b.items[i].String())
	}

	return strings.Join(strs, "\n")
}

func (b *Bucket) GetAll() *Items {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return &b.items
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
