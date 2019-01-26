package epic

import (
	"testing"

	"github.com/JohnTitor/frau/queue"
)

func Test_queuePullReq1(t *testing.T) {
	const number int = 10
	const sha string = "qwerty"

	q := &queue.AutoMergeQueue{}
	item := &queue.AutoMergeQueueItem{
		PullRequest: number,
		PrHead:      sha,
	}
	active := &queue.AutoMergeQueueItem{
		PullRequest: number,
		PrHead:      sha,
	}
	q.SetActive(active)

	ok, mutated := queuePullReq(q, item)
	if !ok {
		t.Fail()
	}

	if mutated {
		t.Fail()
	}

	if !q.HasActive() {
		t.Fail()
	}

	if q.GetActive() != active {
		t.Fail()
	}
}

func Test_queuePullReq2(t *testing.T) {
	const number int = 10
	const sha string = "qwerty"

	q := &queue.AutoMergeQueue{}
	item := &queue.AutoMergeQueueItem{
		PullRequest: number,
		PrHead:      sha,
	}
	active := &queue.AutoMergeQueueItem{
		PullRequest: number,
		PrHead:      sha + "asdfg",
	}
	q.SetActive(active)

	ok, mutated := queuePullReq(q, item)
	if !ok {
		t.Fail()
	}

	if !mutated {
		t.Fail()
	}

	if q.HasActive() {
		t.Fail()
	}

	ok, next := q.TakeNext()
	if !ok {
		t.Fail()
	}

	if next != item {
		t.Fail()
	}
}

func Test_queuePullReq3(t *testing.T) {
	const number int = 10
	const sha string = "qwerty"

	q := &queue.AutoMergeQueue{}
	item := &queue.AutoMergeQueueItem{
		PullRequest: number,
		PrHead:      sha,
	}
	old := &queue.AutoMergeQueueItem{
		PullRequest: number,
		PrHead:      sha,
	}
	if ok := q.Push(old); !ok {
		t.Fail()
	}

	ok, mutated := queuePullReq(q, item)
	if !ok {
		t.Fail()
	}

	if mutated {
		t.Fail()
	}

	if q.HasActive() {
		t.Fail()
	}

	ok, next := q.TakeNext()
	if !ok {
		t.Fail()
	}

	if next != old {
		t.Fail()
	}
}

func Test_queuePullReq4(t *testing.T) {
	const number int = 10
	const sha string = "qwerty"

	q := &queue.AutoMergeQueue{}
	item := &queue.AutoMergeQueueItem{
		PullRequest: number,
		PrHead:      sha,
	}
	old := &queue.AutoMergeQueueItem{
		PullRequest: number,
		PrHead:      sha + "asdfg",
	}
	if ok := q.Push(old); !ok {
		t.Fail()
	}

	ok, mutated := queuePullReq(q, item)
	if !ok {
		t.Fail()
	}

	if !mutated {
		t.Fail()
	}

	if q.HasActive() {
		t.Fail()
	}

	ok, next := q.TakeNext()
	if !ok {
		t.Fail()
	}

	if next != item {
		t.Fail()
	}
}
