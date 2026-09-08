package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"ticket-system/internal/models"
	"ticket-system/internal/repository"
)

var (
	ErrTitleRequired        = errors.New("title is required")
	ErrInvalidStatus        = errors.New("invalid status value")
	ErrTicketNotFound       = errors.New("ticket not found")
	ErrForbiddenAccess      = errors.New("access denied: resource belongs to another user")
	ErrClosedStatusTerminal = errors.New("cannot change status of a closed ticket")
)

type TicketService interface {
	CreateTicket(ctx context.Context, userID string, req models.CreateTicketRequest) (*models.Ticket, error)
	GetTickets(ctx context.Context, userID string, status models.TicketStatus) ([]*models.Ticket, error)
	GetTicketByID(ctx context.Context, userID string, ticketID string) (*models.Ticket, error)
	UpdateTicketStatus(ctx context.Context, userID string, ticketID string, status models.TicketStatus) (*models.Ticket, error)
}

type ticketService struct {
	ticketRepo repository.TicketRepository
}

func NewTicketService(ticketRepo repository.TicketRepository) TicketService {
	return &ticketService{ticketRepo: ticketRepo}
}

func IsValidStatus(status models.TicketStatus) bool {
	switch status {
	case models.StatusOpen, models.StatusInProgress, models.StatusClosed:
		return true
	default:
		return false
	}
}

func CanTransition(from, to models.TicketStatus) bool {
	if !IsValidStatus(to) {
		return false
	}
	if from == models.StatusClosed {
		return to == models.StatusClosed
	}
	return true
}

func (s *ticketService) CreateTicket(ctx context.Context, userID string, req models.CreateTicketRequest) (*models.Ticket, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, ErrTitleRequired
	}

	now := time.Now().UTC()
	ticket := &models.Ticket{
		ID:          uuid.New().String(),
		UserID:      userID,
		Title:       title,
		Description: strings.TrimSpace(req.Description),
		Status:      models.StatusOpen,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.ticketRepo.CreateTicket(ctx, ticket); err != nil {
		return nil, err
	}

	return ticket, nil
}

func (s *ticketService) GetTickets(ctx context.Context, userID string, status models.TicketStatus) ([]*models.Ticket, error) {
	if status != "" && !IsValidStatus(status) {
		return nil, ErrInvalidStatus
	}

	tickets, err := s.ticketRepo.GetTicketsByUserID(ctx, userID, status)
	if err != nil {
		return nil, err
	}

	if tickets == nil {
		tickets = []*models.Ticket{}
	}

	return tickets, nil
}

func (s *ticketService) GetTicketByID(ctx context.Context, userID string, ticketID string) (*models.Ticket, error) {
	ticket, err := s.ticketRepo.GetTicketByID(ctx, ticketID)
	if err != nil {
		if errors.Is(err, repository.ErrTicketNotFound) {
			return nil, ErrTicketNotFound
		}
		return nil, err
	}

	if ticket.UserID != userID {
		return nil, ErrForbiddenAccess
	}

	return ticket, nil
}

func (s *ticketService) UpdateTicketStatus(ctx context.Context, userID string, ticketID string, newStatus models.TicketStatus) (*models.Ticket, error) {
	if !IsValidStatus(newStatus) {
		return nil, ErrInvalidStatus
	}

	ticket, err := s.ticketRepo.GetTicketByID(ctx, ticketID)
	if err != nil {
		if errors.Is(err, repository.ErrTicketNotFound) {
			return nil, ErrTicketNotFound
		}
		return nil, err
	}

	if ticket.UserID != userID {
		return nil, ErrForbiddenAccess
	}

	if ticket.Status == models.StatusClosed && newStatus != models.StatusClosed {
		return nil, ErrClosedStatusTerminal
	}

	if ticket.Status == newStatus {
		return ticket, nil
	}

	if !CanTransition(ticket.Status, newStatus) {
		return nil, ErrClosedStatusTerminal
	}

	updatedTicket, err := s.ticketRepo.UpdateTicketStatus(ctx, ticketID, newStatus)
	if err != nil {
		return nil, err
	}

	return updatedTicket, nil
}
