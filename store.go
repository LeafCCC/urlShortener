package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

//定义url映射的数据结构 使用锁确保读写不冲突
type UrlStore struct {
	urls  map[string]string
	mu    sync.RWMutex //这个锁比较特殊 可以使用RLock()允许多个读 但只能存在一个写
	file  *os.File
	cache chan record
}

//UrlStore的工厂函数
func NewUrlStore(filename string) *UrlStore {
	s := &UrlStore{urls: make(map[string]string)}
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)

	if err != nil {

		log.Fatal("Error opening UrlStore:", err)

	}

	s.file = f

	if err := s.load(); err != nil {
		log.Fatal("Error loading UrlStore:", err)
	}

	//test

	return s
}

//定义获取url映射的函数
func (s *UrlStore) Get(query_url string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.urls[query_url]
}

//存储新的url映射 返回是否成功
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

//返回保存的url映射数量
func (s *UrlStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.urls)
}

//生成对应的短url并使用Set放入映射
func (s *UrlStore) Put(url string) string {
	for {
		key := genKey(s.Count())

		if s.Set(key, url) {
			if err := s.save(key, url); err != nil {
				log.Fatal("Error saving new url: ", err)
			}

			return key
		}
	}

	return ""
}

type record struct {
	Key, URL string
}

//存入文件
func (s *UrlStore) save(key, url string) error {
	e := gob.NewEncoder(s.file)
	return e.Encode(record{key, url})
}

//读取文件
func (s *UrlStore) load() error {
	if _, err := s.file.Seek(0, 0); err != nil {
		return err
	}

	d := gob.NewDecoder(s.file)

	var err error

	for err == nil {
		r := record{}

		//这里注意不要写成err:= debug半天才发现
		if err = d.Decode(&r); err == nil {
			fmt.Println(r)
			s.Set(r.Key, r.URL)
		}
		time.Sleep(time.Second)
	}

	if err == io.EOF { //文件结束错误
		return nil
	}

	return err
}
