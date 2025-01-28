package storage

import (
	"encoding/json"
	"os"
	"sync"

	"heat-logger/internal/models"
)

type JSONStorage struct {
	data     models.HeatingData
	dataLock sync.RWMutex
	filePath string
}

func NewJSONStorage(filePath string) *JSONStorage {
	return &JSONStorage{
		filePath: filePath,
		data: models.HeatingData{
			History: make([]models.Day, 0),
		},
	}
}

func (s *JSONStorage) Load() error {
	s.dataLock.Lock()
	defer s.dataLock.Unlock()

	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return s.save()
	}

	file, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(file, &s.data)
}

func (s *JSONStorage) save() error {
	file, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, file, 0644)
}

func (s *JSONStorage) GetHistory() []models.Day {
	s.dataLock.RLock()
	defer s.dataLock.RUnlock()
	return s.data.History
}

func (s *JSONStorage) AddEntry(day models.Day) error {
	s.dataLock.Lock()
	defer s.dataLock.Unlock()

	s.data.History = append(s.data.History, day)
	return s.save()
}

func (s *JSONStorage) GetRecentEntries(limit int) []models.Day {
	s.dataLock.RLock()
	defer s.dataLock.RUnlock()

	if len(s.data.History) <= limit {
		return s.data.History
	}
	return s.data.History[len(s.data.History)-limit:]
}
