package registry

import (
	"fmt"
	"sync"
)

type MemoryRegistry struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewMemoryRegistry() *MemoryRegistry {
	return &MemoryRegistry{data: make(map[string][]byte)}
}

func (m *MemoryRegistry) SaveFunction(name string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[name] = data
	return nil
}

func (m *MemoryRegistry) GetFunction(name string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if val, ok := m.data[name]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("not found")
}

func (m *MemoryRegistry) ListFunctions() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	names := make([]string, 0, len(m.data))
	for k := range m.data {
		names = append(names, k)
	}
	return names, nil
}
