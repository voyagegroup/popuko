package queue

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type fileRepository struct {
	rootPath string
	dict     map[string]*sync.RWMutex
}

const queueRepoDir = "/queue"

func newFileRepository(path string) *fileRepository {
	root, err := filepath.Abs(path + queueRepoDir)
	if err != nil {
		log.Printf("error: cannot get the path to the queue storage: %v\n", err)
		return nil
	}

	if !exists(root) {
		if err := os.MkdirAll(root, os.ModePerm); err != nil {
			log.Printf("error: cannot create the queue dir: %v\n", err)
			return nil
		}
	}

	return &fileRepository{
		rootPath: root,
		dict:     make(map[string]*sync.RWMutex),
	}
}

func (s *fileRepository) save(owner string, name string, queue *AutoMergeQueue) bool {
	k := owner + "/" + name
	mux, ok := s.dict[k]
	if !ok {
		v := new(sync.RWMutex)
		s.dict[k] = v
		mux = v
	}

	mux.Lock()
	defer mux.Unlock()

	c := autoMergeQFile{
		Version: fileFmtVersion,
		Queue:   queue.q,
		Current: queue.current,
	}

	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		fmt.Println("error: cannot marshal queue:", err)
		return false
	}

	dir, err := filepath.Abs(s.rootPath + "/" + owner)
	if !exists(dir) {
		if err := os.Mkdir(dir, 0775); err != nil {
			log.Println("error: cannot create the config home dir.")
			return false
		}
	}

	path, err := filepath.Abs(dir + "/" + name + ".json")
	if err != nil {
		log.Printf("error: cannot get the path to %v/%v\n", owner, name)
		return false
	}

	file, err := os.Create(path)
	if err != nil {
		log.Printf("error: cannot create the file to %v\n", path)
		return false
	}
	defer file.Close()

	n, err := file.Write(b)
	if err != nil {
		fmt.Printf("error: on writing the file to %v: %v\n", path, err)
		return false
	}

	if n != len(b) {
		fmt.Printf("error: `n != len(b)` on writing the file to %v: %v\n", path, err)
		return false
	}

	return true
}

func (s *fileRepository) load(owner string, name string) (bool, *AutoMergeQueue) {
	k := owner + "/" + name
	mux, ok := s.dict[k]
	if !ok {
		v := new(sync.RWMutex)
		s.dict[k] = v
		mux = v
	}

	mux.RLock()
	defer mux.RUnlock()

	path, err := filepath.Abs(s.rootPath + "/" + owner + "/" + name + ".json")
	if err != nil {
		log.Printf("error: cannot get the path to %v/%v\n", owner, name)
		return false, nil
	}

	if !exists(path) {
		return false, nil
	}

	b, err := ioutil.ReadFile(path)

	var result autoMergeQFile
	if err := json.Unmarshal(b, &result); err != nil {
		fmt.Println("error:", err)
		return true, nil
	}
	fmt.Printf("debug: %+v\n", result)

	q := AutoMergeQueue{
		q:       result.Queue,
		current: result.Current,
	}
	fmt.Printf("debug: %+v\n", q)

	return true, &q
}

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// XXX: Update this field when change the data struct.
const fileFmtVersion int32 = 0

type autoMergeQFile struct {
	Version int32                 `json:"version"`
	Queue   []*AutoMergeQueueItem `json:"queue"`
	Current *AutoMergeQueueItem   `json:"current_active"`
}
