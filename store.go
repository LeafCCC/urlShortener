package main

import "sync"

//定义url映射的数据结构 使用锁确保读写不冲突
type UrlStore struct {
	urls map[string]string
	mu   sync.RWMutex //这个锁比较特殊 可以使用RLock()允许多个读 但只能存在一个写
}

//UrlStore的工厂函数
func NewUrlStore() *UrlStore {
	return &UrlStore{urls: make(map[string]string)}
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
			return key
		}
	}

	return ""
}
