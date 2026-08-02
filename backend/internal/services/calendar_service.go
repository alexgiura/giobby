package services

import (
	"context"
	"fmt"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"
)

// CalendarService handles calendars and events.
type CalendarService interface {
	ListCalendars(ctx context.Context, q models.CalendarListQuery) ([]models.Calendar, error)
	ListEvents(ctx context.Context, q models.CalendarEventListQuery) ([]models.CalendarEvent, error)
	GetEvent(ctx context.Context, id int32) (*models.CalendarEvent, error)
	CreateEvent(ctx context.Context, event models.CalendarEvent) (*models.CalendarEvent, error)
	UpdateEvent(ctx context.Context, id int32, event models.CalendarEvent) (*models.CalendarEvent, error)
	DeleteEvent(ctx context.Context, id int32) error
}

type calendarService struct {
	repo repository.CalendarRepository
}

func NewCalendarService(repo repository.CalendarRepository) CalendarService {
	return &calendarService{repo: repo}
}

func (s *calendarService) ListCalendars(ctx context.Context, q models.CalendarListQuery) ([]models.Calendar, error) {
	return s.repo.ListCalendars(ctx, q)
}

func (s *calendarService) ListEvents(ctx context.Context, q models.CalendarEventListQuery) ([]models.CalendarEvent, error) {
	return s.repo.ListEvents(ctx, q)
}

func (s *calendarService) GetEvent(ctx context.Context, id int32) (*models.CalendarEvent, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid event id")
	}
	return s.repo.GetEvent(ctx, id)
}

func (s *calendarService) CreateEvent(ctx context.Context, event models.CalendarEvent) (*models.CalendarEvent, error) {
	if strings.TrimSpace(event.Event) == "" {
		return nil, fmt.Errorf("event is required")
	}
	if event.IDCalendar <= 0 {
		return nil, fmt.Errorf("idCalendar is required")
	}
	return s.repo.CreateEvent(ctx, &event)
}

func (s *calendarService) UpdateEvent(ctx context.Context, id int32, event models.CalendarEvent) (*models.CalendarEvent, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid event id")
	}
	return s.repo.UpdateEvent(ctx, id, &event)
}

func (s *calendarService) DeleteEvent(ctx context.Context, id int32) error {
	if id <= 0 {
		return fmt.Errorf("invalid event id")
	}
	return s.repo.DeleteEvent(ctx, id)
}
