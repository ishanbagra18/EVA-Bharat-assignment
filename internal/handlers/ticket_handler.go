package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"ticket-system/internal/auth"
	"ticket-system/internal/models"
	"ticket-system/internal/service"
)

type TicketHandler struct {
	ticketService service.TicketService
}

func NewTicketHandler(ticketService service.TicketService) *TicketHandler {
	return &TicketHandler{ticketService: ticketService}
}

func (h *TicketHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		RespondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ticket, err := h.ticketService.CreateTicket(r.Context(), userID, req)
	if err != nil {
		if errors.Is(err, service.ErrTitleRequired) {
			RespondError(w, http.StatusBadRequest, err.Error())
			return
		}
		RespondError(w, http.StatusInternalServerError, "failed to create ticket")
		return
	}

	RespondJSON(w, http.StatusCreated, ticket)
}

func (h *TicketHandler) GetTickets(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		RespondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	statusFilter := models.TicketStatus(r.URL.Query().Get("status"))

	tickets, err := h.ticketService.GetTickets(r.Context(), userID, statusFilter)
	if err != nil {
		if errors.Is(err, service.ErrInvalidStatus) {
			RespondError(w, http.StatusBadRequest, err.Error())
			return
		}
		RespondError(w, http.StatusInternalServerError, "failed to fetch tickets")
		return
	}

	RespondJSON(w, http.StatusOK, tickets)
}

func (h *TicketHandler) GetTicketByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		RespondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	ticketID := chi.URLParam(r, "id")
	if ticketID == "" {
		RespondError(w, http.StatusBadRequest, "ticket id is required")
		return
	}

	ticket, err := h.ticketService.GetTicketByID(r.Context(), userID, ticketID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTicketNotFound):
			RespondError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrForbiddenAccess):
			RespondError(w, http.StatusForbidden, err.Error())
		default:
			RespondError(w, http.StatusInternalServerError, "failed to fetch ticket")
		}
		return
	}

	RespondJSON(w, http.StatusOK, ticket)
}

func (h *TicketHandler) UpdateTicketStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		RespondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	ticketID := chi.URLParam(r, "id")
	if ticketID == "" {
		RespondError(w, http.StatusBadRequest, "ticket id is required")
		return
	}

	var req models.UpdateTicketStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ticket, err := h.ticketService.UpdateTicketStatus(r.Context(), userID, ticketID, req.Status)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidStatus):
			RespondError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrTicketNotFound):
			RespondError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrForbiddenAccess):
			RespondError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrClosedStatusTerminal):
			RespondError(w, http.StatusConflict, err.Error())
		default:
			RespondError(w, http.StatusInternalServerError, "failed to update ticket status")
		}
		return
	}

	RespondJSON(w, http.StatusOK, ticket)
}
