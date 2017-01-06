package epic

import (
	"testing"

	"github.com/karen-irc/popuko/queue"
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

	ok, next := q.GetNext()
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
	q.Push(old)

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

	ok, next := q.GetNext()
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
	q.Push(old)

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

	ok, next := q.GetNext()
	if !ok {
		t.Fail()
	}

	if next != item {
		t.Fail()
	}
}
