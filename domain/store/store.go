package store

import "sync"

type UrlStore struct {
	seenUrls map[string]struct{}
	rwMutex  sync.RWMutex
}

func NewUrlStore() *UrlStore {
	return &UrlStore{
		seenUrls: make(map[string]struct{}),
		rwMutex:  sync.RWMutex{},
	}
}

func (s *UrlStore) Add(url string) error {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()
	s.seenUrls[url] = struct{}{}
	return nil
}

func (s *UrlStore) Seen(url string) (bool, error) {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()
	_, ok := s.seenUrls[url]
	return ok, nil
}
