package queue

import (
	"fmt"
	"log"
	"sync"
)

type AutoMergeQRepo struct {
	mux  sync.Mutex
	repo fileRepository
}

func NewAutoMergeQRepo(root string) *AutoMergeQRepo {
	repo := newFileRepository(root)
	if repo == nil {
		return nil
	}

	return &AutoMergeQRepo{
		mux:  sync.Mutex{},
		repo: *repo,
	}
}

func (s *AutoMergeQRepo) Lock() {
	s.mux.Lock()
}

func (s *AutoMergeQRepo) Unlock() {
	s.mux.Unlock()
}

func (s *AutoMergeQRepo) Get(owner string, name string) *AutoMergeQueue {
	ok, result := s.repo.load(owner, name)
	if !ok || result == nil {
		result = &AutoMergeQueue{
			owner:  owner,
			name:   name,
			parent: s,
		}
		s.Save(owner, name, result)
	}

	result.owner = owner
	result.name = name
	result.parent = s

	return result
}

func (s *AutoMergeQRepo) Save(owner string, name string, v *AutoMergeQueue) {
	if ok := s.repo.save(owner, name, v); !ok {
		log.Printf("error: cannot save the queue information for %v/%v\n", owner, name)
	}
}

type AutoMergeQueue struct {
	mux     sync.Mutex
	q       []*AutoMergeQueueItem
	current *AutoMergeQueueItem

	owner  string
	name   string
	parent *AutoMergeQRepo
}

func (s *AutoMergeQueue) Save() {
	s.parent.Save(s.owner, s.name, s)
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
	PullRequest int     `json:"pull_request"`
	SHA         *string `json:"sha"`
}
