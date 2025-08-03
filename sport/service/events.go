package service

import (
	"github.com/shanehowearth/entain-master/sport/db"
	"github.com/shanehowearth/entain-master/sport/proto/sporting"

	"golang.org/x/net/context"
)

type Sporting interface {
	// GetEvent will return a single event.
	GetEvent(ctx context.Context, in *sporting.GetEventRequest) (*sporting.GetEventResponse, error)
	// ListEvents will return a collection of events.
	ListEvents(ctx context.Context, in *sporting.ListEventsRequest) (*sporting.ListEventsResponse, error)
}

// sportingService implements the Sporting interface.
type sportingService struct {
	eventsRepo db.EventsRepo
}

// NewSportingService instantiates and returns a new sportingService.
func NewSportingService(eventsRepo db.EventsRepo) Sporting {
	return &sportingService{eventsRepo}
}

func (s *sportingService) GetEvent(ctx context.Context, in *sporting.GetEventRequest) (*sporting.GetEventResponse, error) {
	event, err := s.eventsRepo.Get(in.GetEventId())
	if err != nil {
		return nil, err
	}

	return &sporting.GetEventResponse{Event: event}, nil
}
func (s *sportingService) ListEvents(ctx context.Context, in *sporting.ListEventsRequest) (*sporting.ListEventsResponse, error) {
	events, err := s.eventsRepo.List(in.Filter)
	if err != nil {
		return nil, err
	}

	return &sporting.ListEventsResponse{Events: events}, nil
}
