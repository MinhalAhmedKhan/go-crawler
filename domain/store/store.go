package store

type UrlStore struct {
	seenUrls map[string]struct{}
}

func NewUrlStore() *UrlStore {
	return &UrlStore{
		seenUrls: make(map[string]struct{}),
	}
}

func (s *UrlStore) Add(url string) error {
	s.seenUrls[url] = struct{}{}
	return nil
}

func (s *UrlStore) Seen(url string) (bool, error) {
	_, ok := s.seenUrls[url]
	return ok, nil
}
