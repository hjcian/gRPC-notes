package service

import (
	"errors"
	"fmt"
	"sync"

	"github.com/jinzhu/copier"

	"github.com/hjcian/grpc-notes/pb"
)

// LaptopStore is an interface to store laptop
type LaptopStore interface {
	Save(laptop *pb.Laptop) error
	Find(id string) (*pb.Laptop, error)
}

// InMemoryLaptopStore is a InMemoryLaptopStore with RW lock
type InMemoryLaptopStore struct {
	mutex sync.RWMutex
	data  map[string]*pb.Laptop
}

// NewInMemoryLaptopStore returns a new InMemoryLaptopStore
func NewInMemoryLaptopStore() *InMemoryLaptopStore {
	return &InMemoryLaptopStore{
		data: make(map[string]*pb.Laptop),
	}
}

// ErrAlreadyExists is returned when a record with the same ID
var ErrAlreadyExists = errors.New("record already exists")

// Save saves the laptop to the store
func (store *InMemoryLaptopStore) Save(laptop *pb.Laptop) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if store.data[laptop.Id] != nil {
		return ErrAlreadyExists
	}

	other, err := deepCopy(laptop)
	if err != nil {
		return err
	}

	store.data[laptop.Id] = other
	return nil
}

func deepCopy(laptopFrom *pb.Laptop) (*pb.Laptop, error) {
	laptopTo := &pb.Laptop{}

	err := copier.Copy(laptopTo, laptopFrom)
	if err != nil {
		return nil, fmt.Errorf("cannot copy laptop data: %w", err)
	}

	return laptopTo, nil
}

// Find finds a laptop by ID
func (store *InMemoryLaptopStore) Find(id string) (*pb.Laptop, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	laptop := store.data[id]
	if laptop == nil {
		return nil, nil
	}

	return deepCopy(laptop)
}
