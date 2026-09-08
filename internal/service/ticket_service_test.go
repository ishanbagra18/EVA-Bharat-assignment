package service_test

import (
	"testing"

	"ticket-system/internal/models"
	"ticket-system/internal/service"
)

func TestStateTransitions(t *testing.T) {
	tests := []struct {
		name     string
		from     models.TicketStatus
		to       models.TicketStatus
		expected bool
	}{
		{"Open to InProgress", models.StatusOpen, models.StatusInProgress, true},
		{"Open to Closed", models.StatusOpen, models.StatusClosed, true},
		{"InProgress to Closed", models.StatusInProgress, models.StatusClosed, true},
		{"Open to Open (Idempotent)", models.StatusOpen, models.StatusOpen, true},
		{"InProgress to InProgress (Idempotent)", models.StatusInProgress, models.StatusInProgress, true},
		{"Closed to Closed (Idempotent)", models.StatusClosed, models.StatusClosed, true},
		{"Closed to Open (Forbidden)", models.StatusClosed, models.StatusOpen, false},
		{"Closed to InProgress (Forbidden)", models.StatusClosed, models.StatusInProgress, false},
		{"Open to Invalid", models.StatusOpen, models.TicketStatus("invalid_status"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CanTransition(tt.from, tt.to)
			if result != tt.expected {
				t.Errorf("CanTransition(%s, %s) = %v; want %v", tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

func TestIsValidStatus(t *testing.T) {
	if !service.IsValidStatus(models.StatusOpen) {
		t.Errorf("expected open to be valid")
	}
	if !service.IsValidStatus(models.StatusInProgress) {
		t.Errorf("expected in_progress to be valid")
	}
	if !service.IsValidStatus(models.StatusClosed) {
		t.Errorf("expected closed to be valid")
	}
	if service.IsValidStatus(models.TicketStatus("unknown")) {
		t.Errorf("expected unknown to be invalid")
	}
}
