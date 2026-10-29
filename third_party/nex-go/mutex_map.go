package nex

import "sync"


type MutexMap[K comparable, V any] struct {
	*sync.RWMutex
	real map[K]V
}


func (m *MutexMap[K, V]) Set(key K, value V) {
	m.Lock()
	defer m.Unlock()

	m.real[key] = value
}


func (m *MutexMap[K, V]) Get(key K) (V, bool) {
	m.RLock()
	defer m.RUnlock()

	value, ok := m.real[key]

	return value, ok
}



func (m *MutexMap[K, V]) GetOrSetDefault(key K, createDefault func() V) V {
	m.Lock()
	defer m.Unlock()

	value, ok := m.real[key]

	if !ok {
		value = createDefault()
		m.real[key] = value
	}

	return value
}


func (m *MutexMap[K, V]) Has(key K) bool {
	m.RLock()
	defer m.RUnlock()

	_, ok := m.real[key]
	return ok
}


func (m *MutexMap[K, V]) Delete(key K) {
	m.Lock()
	defer m.Unlock()

	delete(m.real, key)
}



func (m *MutexMap[K, V]) DeleteIf(predicate func(key K, value V) bool) int {
	m.Lock()
	defer m.Unlock()

	amount := 0
	for key, value := range m.real {
		if predicate(key, value) {
			delete(m.real, key)
			amount++
		}
	}

	return amount
}


func (m *MutexMap[K, V]) RunAndDelete(key K, callback func(key K, value V)) {
	m.Lock()
	defer m.Unlock()

	if value, ok := m.real[key]; ok {
		callback(key, value)
		delete(m.real, key)
	}
}


func (m *MutexMap[K, V]) Size() int {
	m.RLock()
	defer m.RUnlock()

	return len(m.real)
}




func (m *MutexMap[K, V]) Each(callback func(key K, value V) bool) bool {
	m.RLock()
	defer m.RUnlock()

	for key, value := range m.real {
		if callback(key, value) {
			return true
		}
	}

	return false
}



func (m *MutexMap[K, V]) Clear(callback func(key K, value V)) {
	m.Lock()
	defer m.Unlock()

	for key, value := range m.real {
		if callback != nil {
			callback(key, value)
		}
		delete(m.real, key)
	}
}


func NewMutexMap[K comparable, V any]() *MutexMap[K, V] {
	return &MutexMap[K, V]{
		RWMutex: &sync.RWMutex{},
		real:    make(map[K]V),
	}
}
