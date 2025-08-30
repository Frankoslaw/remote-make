package adapter

import (
	"fmt"
)

type WithID[K comparable] interface {
	ID() K
}

type Repository[K comparable, T WithID[K]] struct {
	inmem map[K]T
}

func NewRepository[K comparable, T WithID[K]]() *Repository[K, T] {
	return &Repository[K, T]{inmem: make(map[K]T)}
}

func (r *Repository[K, T]) Create(item T) error {
	r.inmem[item.ID()] = item
	return nil
}

func (r *Repository[K, T]) Get(id K) (*T, error) {
	item, exists := r.inmem[id]
	if !exists {
		return nil, fmt.Errorf("item not found")
	}
	return &item, nil
}

func (r *Repository[K, T]) List() ([]*T, error) {
	var items []*T
	for _, item := range r.inmem {
		items = append(items, &item)
	}
	return items, nil
}

func (r *Repository[K, T]) Delete(id K) error {
	delete(r.inmem, id)
	return nil
}
