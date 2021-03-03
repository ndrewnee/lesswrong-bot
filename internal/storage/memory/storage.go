package memory

type Storage struct {
	cache map[string]string
}

func NewStorage() *Storage {
	return &Storage{
		cache: make(map[string]string),
	}
}

func (s *Storage) Get(key string) (string, error) {
	return s.cache[key], nil
}

func (s *Storage) Set(key, value string) error {
	s.cache[key] = value
	return nil
}
