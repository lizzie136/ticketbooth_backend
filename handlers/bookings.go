package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	"ticketbooth-backend/models"
	"ticketbooth-backend/repositories"
	"ticketbooth-backend/services"
	"time"
)

type BookingHandler struct {
	bookingService *services.BookingService
	bookingRepo    *repositories.BookingRepository
}

func NewBookingHandler(bookingService *services.BookingService, bookingRepo *repositories.BookingRepository) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
		bookingRepo:    bookingRepo,
	}
}

// CreateBooking handles POST /api/bookings
func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	var req models.BookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "Invalid request body")
		return
	}

	// Validate request
	if req.EventDateID == 0 {
		BadRequest(w, "eventDateId is required")
		return
	}
	if req.CustomerName == "" {
		BadRequest(w, "customerName is required")
		return
	}
	if req.PaymentSource == "" {
		BadRequest(w, "paymentSource is required")
		return
	}
	if req.UserID == 0 {
		BadRequest(w, "userId is required")
		return
	}

	// Determine if GA or seated based on request
	var response *models.BookingResponse
	var err error

	if len(req.Tiers) > 0 {
		// GA booking
		response, err = h.bookingService.BookGATickets(&req)
	} else if len(req.Seats) > 0 {
		// Seated booking
		response, err = h.bookingService.BookSeatedTickets(&req)
	} else {
		BadRequest(w, "Either tiers or seats must be provided")
		return
	}

	if err != nil {
		fmt.Println(err)
		if err == services.ErrInsufficientInventory {
			Conflict(w, "INSUFFICIENT_INVENTORY", "Not enough tickets left for one or more requested tiers.")
			return
		}
		if err == services.ErrSeatAlreadyTaken {
			Conflict(w, "SEAT_ALREADY_TAKEN", "One or more selected seats are no longer available.")
			return
		}
		if err == services.ErrNotFound {
			NotFound(w, "Event date not found")
			return
		}
		InternalServerError(w, "Failed to create booking")
		return
	}

	JSON(w, http.StatusCreated, response)
}

// GetOrder handles GET /api/orders/:id
func (h *BookingHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		BadRequest(w, "Invalid order ID")
		return
	}

	order, err := h.bookingRepo.GetOrderByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			NotFound(w, "Order not found")
			return
		}
		fmt.Println("get order error", err)
		InternalServerError(w, "Failed to fetch order")
		return
	}

	// Convert to response format
	response := &models.OrderResponse{
		ID:           order.ID,
		CustomerName: "", // Will be set from first ticket's to_name
		TotalAmount:  0,
		Tickets:      []*models.OrderTicketResponse{},
	}

	// Parse total amount
	if order.Amount != "" {
		if amount, err := strconv.ParseFloat(order.Amount, 64); err == nil {
			response.TotalAmount = amount
		}
	}

	// Format created at
	if order.CreatedAt != nil {
		response.CreatedAt = order.CreatedAt.Format(time.RFC3339)
	}

	// Convert tickets
	for _, ticket := range order.Tickets {
		ticketResp := &models.OrderTicketResponse{
			ID:         ticket.ID,
			TicketType: "",
		}

		if ticket.TicketType != nil {
			ticketResp.TicketType = ticket.TicketType.Name
		}

		if ticket.Seat != nil {
			seatLabel := ticket.Seat.Section + ticket.Seat.Row + ticket.Seat.Number
			ticketResp.SeatLabel = &seatLabel
		}

		// Get event info from ticket
		if ticket.Event != nil {
			ticketResp.EventTitle = ticket.Event.Title
		}

		if ticket.EventDate != nil && ticket.EventDate.Date != nil {
			ticketResp.EventDate = ticket.EventDate.Date.Format(time.RFC3339)
		}

		response.Tickets = append(response.Tickets, ticketResp)

		// Set customer name from first ticket
		if response.CustomerName == "" {
			response.CustomerName = ticket.ToName
		}
	}

	JSON(w, http.StatusOK, response)
}

// get all user orders
func GetAllUserOrdersHandler(bookingRepo repositories.BookingRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Assuming user ID comes from context (middleware sets it)
		userID, ok := r.Context().Value("userID").(string)
		if !ok || userID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		orders, err := bookingRepo.GetAllOrdersByUserID(userID)
		if err != nil {
			http.Error(w, "Failed to fetch orders", http.StatusInternalServerError)
			return
		}

		var response []*models.OrderResponse
		for _, order := range orders {
			orderResp := &models.OrderResponse{
				ID:           order.ID,
				CustomerName: "",
			}
			if order.CreatedAt != nil {
				orderResp.CreatedAt = order.CreatedAt.Format(time.RFC3339)
			}

			var tickets []*models.OrderTicketResponse
			for _, ticket := range order.Tickets {
				ticketResp := &models.OrderTicketResponse{
					ID:         ticket.ID,
					TicketType: "",
				}
				if ticket.TicketType != nil {
					ticketResp.TicketType = ticket.TicketType.Name
				}
				if ticket.Seat != nil {
					seatLabel := ticket.Seat.Section + ticket.Seat.Row + ticket.Seat.Number
					ticketResp.SeatLabel = &seatLabel
				}
				if ticket.Event != nil {
					ticketResp.EventTitle = ticket.Event.Title
				}
				if ticket.EventDate != nil && ticket.EventDate.Date != nil {
					ticketResp.EventDate = ticket.EventDate.Date.Format(time.RFC3339)
				}
				tickets = append(tickets, ticketResp)
				// Set customer name from first ticket
				if orderResp.CustomerName == "" {
					orderResp.CustomerName = ticket.ToName
				}
			}
			orderResp.Tickets = tickets

			response = append(response, orderResp)
		}

		JSON(w, http.StatusOK, response)
	}
}

// GetOrders handles GET /api/orders?userId={id}
func (h *BookingHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	if userID == "" {
		BadRequest(w, "userId query parameter is required")
		return
	}

	orders, err := h.bookingRepo.GetAllOrdersByUserID(userID)
	if err != nil {
		InternalServerError(w, "Failed to fetch orders")
		return
	}

	var response []*models.OrderResponse
	for _, order := range orders {
		resp := &models.OrderResponse{
			ID:           order.ID,
			CustomerName: "",
		}
		if order.CreatedAt != nil {
			resp.CreatedAt = order.CreatedAt.Format(time.RFC3339)
		}
		if order.Amount != "" {
			if amount, err := strconv.ParseFloat(order.Amount, 64); err == nil {
				resp.TotalAmount = amount
			}
		}

		var tickets []*models.OrderTicketResponse
		for _, ticket := range order.Tickets {
			ticketResp := &models.OrderTicketResponse{
				ID:         ticket.ID,
				TicketType: "",
			}
			if ticket.TicketType != nil {
				ticketResp.TicketType = ticket.TicketType.Name
			}
			if ticket.Seat != nil {
				seatLabel := ticket.Seat.Section + ticket.Seat.Row + ticket.Seat.Number
				ticketResp.SeatLabel = &seatLabel
			}
			if ticket.Event != nil {
				ticketResp.EventTitle = ticket.Event.Title
			}
			if ticket.EventDate != nil && ticket.EventDate.Date != nil {
				ticketResp.EventDate = ticket.EventDate.Date.Format(time.RFC3339)
			}
			if resp.CustomerName == "" {
				resp.CustomerName = ticket.ToName
			}
			tickets = append(tickets, ticketResp)
		}

		resp.Tickets = tickets
		response = append(response, resp)
	}

	JSON(w, http.StatusOK, response)
}
