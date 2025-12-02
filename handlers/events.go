package handlers

import (
	"database/sql"
	//"encoding/json"
	"net/http"
	"strconv"
	"time"
	"ticketbooth-backend/models"
	"ticketbooth-backend/repositories"
	"github.com/go-chi/chi/v5"
)

type EventHandler struct {
	eventRepo       *repositories.EventRepository
	availabilityRepo *repositories.AvailabilityRepository
}

func NewEventHandler(eventRepo *repositories.EventRepository, availabilityRepo *repositories.AvailabilityRepository) *EventHandler {
	return &EventHandler{
		eventRepo:       eventRepo,
		availabilityRepo: availabilityRepo,
	}
}

// GetEvents handles GET /api/events
func (h *EventHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.eventRepo.GetAllEvents()
	if err != nil {
		InternalServerError(w, "Failed to fetch events")
		return
	}

	JSON(w, http.StatusOK, events)
}

// GetEventDate handles GET /api/event-dates/:id
func (h *EventHandler) GetEventDate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		BadRequest(w, "Invalid event date ID")
		return
	}

	eventDate, err := h.eventRepo.GetEventDateByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			NotFound(w, "Event date not found")
			return
		}
		InternalServerError(w, "Failed to fetch event date")
		return
	}

	// Convert to response format
	response := &models.EventDateResponse{
		ID:          eventDate.ID,
		SeatingMode: eventDate.SeatingMode,
	}

	if eventDate.Event != nil {
		response.Event = &models.EventInfo{
			ID:          eventDate.Event.ID,
			Title:       eventDate.Event.Title,
			Description: eventDate.Event.Description,
		}
	}

	if eventDate.Date != nil {
		response.Date = eventDate.Date.Format(time.RFC3339)
	}

	if eventDate.Venue != nil {
		response.Venue = &models.VenueInfo{
			ID:       eventDate.Venue.ID,
			Name:     eventDate.Venue.Name,
			Capacity: eventDate.Venue.Capacity,
		}
	}

	JSON(w, http.StatusOK, response)
}

// GetAvailability handles GET /api/event-dates/:id/availability
func (h *EventHandler) GetAvailability(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		BadRequest(w, "Invalid event date ID")
		return
	}

	// Get event date to determine seating mode
	eventDate, err := h.eventRepo.GetEventDateByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			NotFound(w, "Event date not found")
			return
		}
		InternalServerError(w, "Failed to fetch event date")
		return
	}

	if eventDate.SeatingMode == "GA" {
		// GA availability
		tiers, err := h.availabilityRepo.GetGAAvailability(id)
		if err != nil {
			InternalServerError(w, "Failed to fetch availability")
			return
		}

		response := &models.GAAvailabilityResponse{
			SeatingMode: "GA",
			Tiers:       tiers,
		}
		JSON(w, http.StatusOK, response)
	} else if eventDate.SeatingMode == "SEATED" {
		// Seated availability
		sections, err := h.availabilityRepo.GetSeatedAvailability(id)
		if err != nil {
			InternalServerError(w, "Failed to fetch availability")
			return
		}

		response := &models.SeatedAvailabilityResponse{
			SeatingMode: "SEATED",
			Sections:    sections,
		}
		JSON(w, http.StatusOK, response)
	} else {
		BadRequest(w, "Invalid seating mode")
	}
}

