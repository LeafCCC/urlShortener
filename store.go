package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// UrlStore中chan缓存大小
const cacheLen = 1000

// 定义url映射的数据结构 使用锁确保读写不冲突
type UrlStore struct {
	urls  map[string]string
	mu    sync.RWMutex //这个锁比较特殊 可以使用RLock()允许多个读 但只能存在一个写
	file  *os.File
	cache chan record
}

// UrlStore的工厂函数
func NewUrlStore(filename string) *UrlStore {
	s := &UrlStore{urls: make(map[string]string)}
	f, err := os.Open(filename)
	defer f.Close()

	if err != nil {

		log.Fatal("Error opening UrlStore:", err)

	}

	s.file = f

	if err := s.load(); err != nil {
		log.Fatal("Error loading UrlStore:", err)
	}

	//创建一个长度
	s.cache = make(chan record, cacheLen)

	//运行从缓存中存储到磁盘里的函数
	go s.saveLoop(filename)

	return s
}

func (s *UrlStore) saveLoop(filename string) {
	f, err := os.Open(filename)

	if err != nil {
		log.Fatal("UrlStore:", err)
	}

	defer f.Close()

	e := json.NewEncoder(f)
	for {
		r := <-s.cache
		if err := e.Encode(r); err != nil {
			log.Println("UrlStore:", err)
		}
	}
}

// 定义获取url映射的函数
func (s *UrlStore) Get(query_url string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.urls[query_url]
}

// 存储新的url映射 返回是否成功
func (s *UrlStore) Set(key, val string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	//如果已经存在映射则返回失败
	_, present := s.urls[key]
	if present {
		return false
	}
	s.urls[key] = val
	return true
}

// 返回保存的url映射数量
func (s *UrlStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.urls)
}

// 生成对应的短url并使用Set放入映射
func (s *UrlStore) Put(url string) string {
	for {
		key := genKey(s.Count())

		if s.Set(key, url) {
			s.cache <- record{key, url}
			return key
		}
	}

	return ""
}

type record struct {
	Key, URL string
}

// 存入文件
func (s *UrlStore) save(key, url string) error {
	e := json.NewEncoder(s.file)
	return e.Encode(record{key, url})
}

// 读取文件
func (s *UrlStore) load() error {
	if _, err := s.file.Seek(0, 0); err != nil {
		return err
	}

	d := json.NewDecoder(s.file)

	var err error

	for err == nil {
		r := record{}

		//这里注意不要写成err:= debug半天才发现
		if err = d.Decode(&r); err == nil {
			fmt.Println("A new url map read,", r)
			s.Set(r.Key, r.URL)
		}
		time.Sleep(time.Second)
	}

	if err == io.EOF { //文件结束错误
		return nil
	}

	return err
}
