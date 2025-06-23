package storage

import (
	"fmt"
	"sync"
	"time"

	"url-shortener/models"
)

type MemoryStorage struct {
	urls   map[string]*models.URL
	codes  map[string]string // shortCode -> id mapping
	mutex  sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		urls:  make(map[string]*models.URL),
		codes: make(map[string]string),
	}
}

func (s *MemoryStorage) Store(url *models.URL) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.urls[url.ID] = url
	s.codes[url.ShortCode] = url.ID
	return nil
}

func (s *MemoryStorage) GetByShortCode(shortCode string) (*models.URL, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	id, exists := s.codes[shortCode]
	if !exists {
		return nil, fmt.Errorf("short code not found")
	}
	
	url, exists := s.urls[id]
	if !exists {
		return nil, fmt.Errorf("URL not found")
	}
	
	return url, nil
}

func (s *MemoryStorage) IncrementAccess(shortCode string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	id, exists := s.codes[shortCode]
	if !exists {
		return fmt.Errorf("short code not found")
	}
	
	url, exists := s.urls[id]
	if !exists {
		return fmt.Errorf("URL not found")
	}
	
	url.AccessCount++
	url.LastAccess = time.Now()
	return nil
}

func (s *MemoryStorage) GetStats() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	return map[string]interface{}{
		"total_urls": len(s.urls),
		"total_codes": len(s.codes),
	}
}
