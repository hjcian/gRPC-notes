package service

import "sync"

type Rating struct {
	Count uint32
	Sum   float64
}

type RatingStore interface {
	Add(laptopID string, score float64) (*Rating, error)
}

type InMemoryRatingStore struct {
	mutex  sync.RWMutex
	rating map[string]*Rating
}

func NewInMemoryRatingStore() *InMemoryRatingStore {
	return &InMemoryRatingStore{
		rating: make(map[string]*Rating),
	}
}

func (s *InMemoryRatingStore) Add(laptopID string, score float64) (*Rating, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	rating := s.rating[laptopID]
	if rating == nil {
		rating = &Rating{
			Count: 1,
			Sum:   score,
		}
	} else {
		rating.Count++
		rating.Sum += score
	}

	s.rating[laptopID] = rating
	return rating, nil
}
