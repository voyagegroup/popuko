package queue

import (
	"log"
	"testing"
)

// Should remove the active
func Test_AutoMergeQueue_RemoveAwaiting1(t *testing.T) {
	const number int = 1

	queue := AutoMergeQueue{}
	i1 := &AutoMergeQueueItem{
		PullRequest: number,
	}
	queue.SetActive(i1)

	if ok := queue.RemoveAwaiting(number); !ok {
		t.Fatalf("should be success to remove the awaiting")
	}

	if queue.HasActive() {
		t.Fatalf("queue.HasActive() should be false")
	}

	if queue.GetActive() != nil {
		t.Fatalf("queue.GetActive() should be nil")
	}
}

// Should remove the item in the queue.
func Test_AutoMergeQueue_RemoveAwaiting2(t *testing.T) {
	const number int = 1

	queue := AutoMergeQueue{}
	list := []*AutoMergeQueueItem{
		&AutoMergeQueueItem{
			PullRequest: number,
		},
		&AutoMergeQueueItem{
			PullRequest: number + 1,
		},
		&AutoMergeQueueItem{
			PullRequest: number + 2,
		},
	}
	for _, item := range list {
		if ok := queue.Push(item); !ok {
			t.Fail()
		}
	}

	if ok := queue.RemoveAwaiting(number + 1); !ok {
		t.Fatalf("should be success to remove the awaiting")
	}

	ok, next := queue.TakeNext()
	if !ok {
		t.Fatalf("queue.TakeNext() should be ok=true")
	}

	if next != list[0] {
		log.Printf("debug: queue is :%+v\n", queue)
		t.Fatalf("queue.TakeNext() should return the front of list, but %v\n", next)
	}
}
