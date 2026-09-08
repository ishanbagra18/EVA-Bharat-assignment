package repository

import (
	"context"
	"database/sql"
	"errors"

	"ticket-system/internal/models"
)

var (
	ErrTicketNotFound = errors.New("ticket not found")
)

type TicketRepository interface {
	CreateTicket(ctx context.Context, ticket *models.Ticket) error
	GetTicketsByUserID(ctx context.Context, userID string, status models.TicketStatus) ([]*models.Ticket, error)
	GetTicketByID(ctx context.Context, id string) (*models.Ticket, error)
	UpdateTicketStatus(ctx context.Context, id string, status models.TicketStatus) (*models.Ticket, error)
}

type SQLiteTicketRepository struct {
	db *sql.DB
}

func NewTicketRepository(db *sql.DB) TicketRepository {
	return &SQLiteTicketRepository{db: db}
}

func (r *SQLiteTicketRepository) CreateTicket(ctx context.Context, ticket *models.Ticket) error {
	query := `INSERT INTO tickets (id, user_id, title, description, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, ticket.ID, ticket.UserID, ticket.Title, ticket.Description, ticket.Status, ticket.CreatedAt, ticket.UpdatedAt)
	return err
}

func (r *SQLiteTicketRepository) GetTicketsByUserID(ctx context.Context, userID string, status models.TicketStatus) ([]*models.Ticket, error) {
	var query string
	var args []interface{}

	if status != "" {
		query = `SELECT id, user_id, title, description, status, created_at, updated_at FROM tickets WHERE user_id = ? AND status = ? ORDER BY created_at DESC`
		args = []interface{}{userID, string(status)}
	} else {
		query = `SELECT id, user_id, title, description, status, created_at, updated_at FROM tickets WHERE user_id = ? ORDER BY created_at DESC`
		args = []interface{}{userID}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tickets := []*models.Ticket{}
	for rows.Next() {
		t := &models.Ticket{}
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tickets = append(tickets, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tickets, nil
}

func (r *SQLiteTicketRepository) GetTicketByID(ctx context.Context, id string) (*models.Ticket, error) {
	query := `SELECT id, user_id, title, description, status, created_at, updated_at FROM tickets WHERE id = ?`
	t := &models.Ticket{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTicketNotFound
		}
		return nil, err
	}
	return t, nil
}

func (r *SQLiteTicketRepository) UpdateTicketStatus(ctx context.Context, id string, status models.TicketStatus) (*models.Ticket, error) {
	query := `UPDATE tickets SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? RETURNING id, user_id, title, description, status, created_at, updated_at`
	t := &models.Ticket{}
	err := r.db.QueryRowContext(ctx, query, string(status), id).Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTicketNotFound
		}
		return nil, err
	}
	return t, nil
}
