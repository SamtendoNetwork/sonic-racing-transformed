package nex

import "sync"






type MutexSlice[V comparable] struct {
	*sync.RWMutex
	real []V
}


func (m *MutexSlice[V]) Add(value V) {
	m.Lock()
	defer m.Unlock()

	m.real = append(m.real, value)
}




func (m *MutexSlice[V]) Delete(value V) bool {
	m.Lock()
	defer m.Unlock()

	for i, v := range m.real {
		if v == value {
			m.real = append(m.real[:i], m.real[i+1:]...)
			return true
		}
	}

	return false
}




func (m *MutexSlice[V]) DeleteAll(value V) bool {
	m.Lock()
	defer m.Unlock()

	newSlice := make([]V, 0)
	oldLength := len(m.real)

	for _, v := range m.real {
		if v != value {
			newSlice = append(newSlice, v)
		}
	}

	m.real = newSlice

	return len(newSlice) < oldLength
}


func (m *MutexSlice[V]) Has(value V) bool {
	m.Lock()
	defer m.Unlock()

	for _, v := range m.real {
		if v == value {
			return true
		}
	}

	return false
}




func (m *MutexSlice[V]) GetIndex(value V) int {
	m.Lock()
	defer m.Unlock()

	for i, v := range m.real {
		if v == value {
			return i
		}
	}

	return -1
}




func (m *MutexSlice[V]) At(index int) (V, bool) {
	m.Lock()
	defer m.Unlock()

	if index >= len(m.real) {
		return *new(V), false
	}

	return m.real[index], true
}


func (m *MutexSlice[V]) Values() []V {
	m.Lock()
	defer m.Unlock()

	return m.real
}


func (m *MutexSlice[V]) Size() int {
	m.RLock()
	defer m.RUnlock()

	return len(m.real)
}






func (m *MutexSlice[V]) Each(callback func(index int, value V) bool) bool {
	m.RLock()
	defer m.RUnlock()

	for i, value := range m.real {
		if callback(i, value) {
			return true
		}
	}

	return false
}


func (m *MutexSlice[V]) Clear() {
	m.Lock()
	defer m.Unlock()

	m.real = make([]V, 0)
}


func NewMutexSlice[V comparable]() *MutexSlice[V] {
	return &MutexSlice[V]{
		RWMutex: &sync.RWMutex{},
		real:    make([]V, 0),
	}
}
