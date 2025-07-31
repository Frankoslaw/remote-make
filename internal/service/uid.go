package service

import "sync"

type UIDActions interface {
	Generate() int
}

type UIDService struct {
	UIDActions
	mu  sync.Mutex
	cur int
}

func NewUIDService() *UIDService {
	return &UIDService{cur: 1}
}

func (s *UIDService) Generate() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.cur
	s.cur++
	return id
}
