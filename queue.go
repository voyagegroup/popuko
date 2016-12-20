package main

import (
	"fmt"
	"log"
	"sync"
)

type autoMergeQRepo struct {
	mux sync.Mutex
	m   map[string]*autoMergeQueue
}

func newAutoMergeQRepo() *autoMergeQRepo {
	return &autoMergeQRepo{
		mux: sync.Mutex{},
		m:   make(map[string]*autoMergeQueue),
	}
}

func (s *autoMergeQRepo) Lock() {
	s.mux.Lock()
}

func (s *autoMergeQRepo) Unlock() {
	s.mux.Unlock()
}

func (s *autoMergeQRepo) Get(owner string, name string) *autoMergeQueue {
	k := owner + "/" + name
	item, ok := s.m[k]
	if !ok {
		n := &autoMergeQueue{}
		s.m[k] = n
		return n
	}

	return item
}

func (s *autoMergeQRepo) Remove(owner string, name string) {
	k := owner + "/" + name
	delete(s.m, k)
}

func (s *autoMergeQRepo) Save(owner string, name string, v *autoMergeQueue) {
	k := owner + "/" + name
	s.m[k] = v
}

type autoMergeQueue struct {
	isLocked bool
	mux      sync.Mutex
	q        []*autoMergeQueueItem
	current  *autoMergeQueueItem
}

func (s *autoMergeQueue) Lock() {
	s.mux.Lock()
}

func (s *autoMergeQueue) Unlock() {
	s.mux.Unlock()
}

func (s *autoMergeQueue) Push(item *autoMergeQueueItem) {
	s.q = append(s.q, item)
}

func (s *autoMergeQueue) GetNext() (ok bool, item *autoMergeQueueItem) {
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

func (s *autoMergeQueue) GetActive() *autoMergeQueueItem {
	return s.current
}

func (s *autoMergeQueue) SetActive(item *autoMergeQueueItem) error {
	if s.HasActive() {
		return fmt.Errorf("warn: active item has been already set!")
	}

	s.current = item
	return nil
}

func (s *autoMergeQueue) RemoveActive() {
	s.current = nil
}

func (s *autoMergeQueue) HasActive() bool {
	return s.current != nil
}

type autoMergeQueueItem struct {
	PullRequest int
	SHA         *string
}
