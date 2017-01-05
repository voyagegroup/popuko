package queue

import (
	"fmt"
	"log"
	"sync"
)

type AutoMergeQRepo struct {
	mux     sync.Mutex
	repo    fileRepository
	qHandle map[string]*AutoMergeQueueHandle
}

func NewAutoMergeQRepo(root string) *AutoMergeQRepo {
	repo := newFileRepository(root)
	if repo == nil {
		return nil
	}

	return &AutoMergeQRepo{
		mux:     sync.Mutex{},
		repo:    *repo,
		qHandle: make(map[string]*AutoMergeQueueHandle),
	}
}

func (s *AutoMergeQRepo) Get(owner string, name string) *AutoMergeQueueHandle {
	s.mux.Lock()
	defer s.mux.Unlock()

	k := owner + "/" + name
	h, ok := s.qHandle[k]
	if !ok {
		h = &AutoMergeQueueHandle{
			owner:  owner,
			name:   name,
			parent: s,
		}
		s.qHandle[k] = h
	}

	return h
}

func (s *AutoMergeQRepo) save(owner string, name string, v *AutoMergeQueue) {
	if ok := s.repo.save(owner, name, v); !ok {
		log.Printf("error: cannot save the queue information for %v/%v\n", owner, name)
	}
}

type AutoMergeQueueHandle struct {
	sync.Mutex

	owner  string
	name   string
	parent *AutoMergeQRepo
}

func (s *AutoMergeQueueHandle) Load() *AutoMergeQueue {
	owner := s.owner
	name := s.name

	ok, result := s.parent.repo.load(owner, name)
	if !ok || result == nil {
		result = &AutoMergeQueue{}
		s.parent.save(owner, name, result)
	}

	result.ownerHandle = s

	return result
}

type AutoMergeQueue struct {
	ownerHandle *AutoMergeQueueHandle

	q       []*AutoMergeQueueItem
	current *AutoMergeQueueItem
}

func (s *AutoMergeQueue) Save() {
	s.ownerHandle.parent.save(s.ownerHandle.owner, s.ownerHandle.name, s)
}

func (s *AutoMergeQueue) Push(item *AutoMergeQueueItem) {
	s.q = append(s.q, item)
}

func (s *AutoMergeQueue) GetNext() (ok bool, item *AutoMergeQueueItem) {
	if len(s.q) == 0 {
		return true, nil
	}

	front, q := s.q[0], s.q[1:]
	s.q = q

	if front == nil {
		log.Println("error: the front of auto merge queue is nil")
		return
	}

	return true, front
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
