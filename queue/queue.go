package queue

import (
	"fmt"
	"log"
	"sync"
)

type AutoMergeQRepo struct {
	mux sync.Mutex
	m   map[string]*AutoMergeQueue
}

func NewAutoMergeQRepo() *AutoMergeQRepo {
	return &AutoMergeQRepo{
		mux: sync.Mutex{},
		m:   make(map[string]*AutoMergeQueue),
	}
}

func (s *AutoMergeQRepo) Lock() {
	s.mux.Lock()
}

func (s *AutoMergeQRepo) Unlock() {
	s.mux.Unlock()
}

func (s *AutoMergeQRepo) Get(owner string, name string) *AutoMergeQueue {
	k := owner + "/" + name
	item, ok := s.m[k]
	if !ok {
		n := &AutoMergeQueue{}
		s.m[k] = n
		return n
	}

	return item
}

func (s *AutoMergeQRepo) Remove(owner string, name string) {
	k := owner + "/" + name
	delete(s.m, k)
}

func (s *AutoMergeQRepo) Save(owner string, name string, v *AutoMergeQueue) {
	k := owner + "/" + name
	s.m[k] = v
}

type AutoMergeQueue struct {
	isLocked bool
	mux      sync.Mutex
	q        []*AutoMergeQueueItem
	current  *AutoMergeQueueItem
}

func (s *AutoMergeQueue) Lock() {
	s.mux.Lock()
}

func (s *AutoMergeQueue) Unlock() {
	s.mux.Unlock()
}

func (s *AutoMergeQueue) Push(item *AutoMergeQueueItem) {
	s.q = append(s.q, item)
}

func (s *AutoMergeQueue) GetNext() (ok bool, item *AutoMergeQueueItem) {
	if len(s.q) == 0 {
		return true, nil
	}

	f, q := s.q[0], s.q[1:]
	s.q = q

	if f == nil {
		log.Println("error: the front of auto merge queue is nil")
		return
	}

	return true, f
}

func (s *AutoMergeQueue) GetActive() *AutoMergeQueueItem {
	return s.current
}

func (s *AutoMergeQueue) SetActive(item *AutoMergeQueueItem) error {
	if s.HasActive() {
		return fmt.Errorf("warn: active item has been already set!")
	}

	s.current = item
	return nil
}

func (s *AutoMergeQueue) RemoveActive() {
	s.current = nil
}

func (s *AutoMergeQueue) HasActive() bool {
	return s.current != nil
}

type AutoMergeQueueItem struct {
	PullRequest int
	SHA         *string
}
