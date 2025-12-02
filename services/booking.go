package services

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"ticketbooth-backend/db"
	"ticketbooth-backend/models"
	"ticketbooth-backend/repositories"
	"github.com/jmoiron/sqlx"
)

var (
	ErrInsufficientInventory = errors.New("INSUFFICIENT_INVENTORY")
	ErrSeatAlreadyTaken      = errors.New("SEAT_ALREADY_TAKEN")
	ErrNotFound              = errors.New("NOT_FOUND")
)

type BookingService struct {
	db              *db.DB
	bookingRepo     *repositories.BookingRepository
	inventoryRepo   *repositories.InventoryRepository
	eventRepo       *repositories.EventRepository
	ticketTypeRepo  *repositories.TicketTypeRepository
	seatRepo        *repositories.SeatRepository
}

func NewBookingService(
	db *db.DB,
	bookingRepo *repositories.BookingRepository,
	inventoryRepo *repositories.InventoryRepository,
	eventRepo *repositories.EventRepository,
	ticketTypeRepo *repositories.TicketTypeRepository,
	seatRepo *repositories.SeatRepository,
) *BookingService {
	return &BookingService{
		db:            db,
		bookingRepo:  bookingRepo,
		inventoryRepo: inventoryRepo,
		eventRepo:    eventRepo,
		ticketTypeRepo: ticketTypeRepo,
		seatRepo:     seatRepo,
	}
}

// BookGATickets handles GA booking with transaction and concurrency control
func (s *BookingService) BookGATickets(req *models.BookingRequest) (*models.BookingResponse, error) {
	// First, get the event date to verify it's GA
	eventDate, err := s.eventRepo.GetEventDateByID(req.EventDateID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if eventDate.SeatingMode != "GA" {
		return nil, fmt.Errorf("event date is not GA mode")
	}

	var response *models.BookingResponse

	err = s.db.WithTx(func(tx *sqlx.Tx) error {
		// Validate and update inventory for each tier
		totalAmount := 0.0
		totalTickets := 0

		for _, tier := range req.Tiers {
			// Atomically update inventory
			rowsAffected, err := s.inventoryRepo.UpdateGATicketInventory(tx, req.EventDateID, tier.TicketTypeID, tier.Quantity)
			if err != nil {
				return err
			}

			if rowsAffected == 0 {
				// Get ticket type name for error message
				_, remaining, err := s.inventoryRepo.GetGATicketPriceAndRemaining(req.EventDateID, tier.TicketTypeID)
				if err == nil {
					return fmt.Errorf("%w: Not enough tickets left for ticket type %d (remaining: %d, requested: %d)", 
						ErrInsufficientInventory, tier.TicketTypeID, remaining, tier.Quantity)
				}
				return fmt.Errorf("%w: Not enough tickets left", ErrInsufficientInventory)
			}

			// Get price for calculation
			price, _, err := s.inventoryRepo.GetGATicketPriceAndRemaining(req.EventDateID, tier.TicketTypeID)
			if err != nil {
				return err
			}

			totalAmount += price * float64(tier.Quantity)
			totalTickets += tier.Quantity
		}

		// Create order
		orderID, err := s.bookingRepo.CreateOrder(tx, req.UserID, totalTickets, fmt.Sprintf("%.2f", totalAmount), req.PaymentSource)
		if err != nil {
			return err
		}

		// Create tickets
		var tickets []*models.TicketResponse
		eventIDStr := strconv.Itoa(eventDate.EventID)
		userIDStr := strconv.Itoa(req.UserID)

		for _, tier := range req.Tiers {
			for i := 0; i < tier.Quantity; i++ {
				ticketID, err := s.bookingRepo.CreateTicket(tx, int(orderID), userIDStr, eventIDStr, req.EventDateID, tier.TicketTypeID, 0, req.CustomerName)
				if err != nil {
					return err
				}

				// Get ticket type name
				ticketType, err := s.ticketTypeRepo.GetTicketTypeByID(tier.TicketTypeID)
				ticketTypeName := fmt.Sprintf("TicketType-%d", tier.TicketTypeID)
				if err == nil && ticketType != nil {
					ticketTypeName = ticketType.Name
				}

				tickets = append(tickets, &models.TicketResponse{
					ID:         int(ticketID),
					TicketType: ticketTypeName,
					SeatLabel:  nil,
					ToName:     req.CustomerName,
				})
			}
		}

		response = &models.BookingResponse{
			OrderID:     int(orderID),
			TotalAmount: totalAmount,
			Tickets:     tickets,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}

// BookSeatedTickets handles seated booking with transaction and unique constraint protection
func (s *BookingService) BookSeatedTickets(req *models.BookingRequest) (*models.BookingResponse, error) {
	// First, get the event date to verify it's SEATED
	eventDate, err := s.eventRepo.GetEventDateByID(req.EventDateID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if eventDate.SeatingMode != "SEATED" {
		return nil, fmt.Errorf("event date is not SEATED mode")
	}

	var response *models.BookingResponse

	err = s.db.WithTx(func(tx *sqlx.Tx) error {
		// Check if any seats are already taken
		seatIDs := make([]int, len(req.Seats))
		for i, seat := range req.Seats {
			seatIDs[i] = seat.SeatID
		}

		bookedSeats, err := s.inventoryRepo.CheckSeatAvailability(req.EventDateID, seatIDs)
		if err != nil {
			return err
		}

		if len(bookedSeats) > 0 {
			return fmt.Errorf("%w: One or more selected seats are no longer available", ErrSeatAlreadyTaken)
		}

		// Calculate total amount
		totalAmount := 0.0
		for _, seatReq := range req.Seats {
			price, _, err := s.inventoryRepo.GetSeatPriceAndTicketType(req.EventDateID, seatReq.SeatID)
			if err != nil {
				return err
			}
			totalAmount += price
		}

		// Create order
		orderID, err := s.bookingRepo.CreateOrder(tx, req.UserID, len(req.Seats), fmt.Sprintf("%.2f", totalAmount), req.PaymentSource)
		if err != nil {
			return err
		}

		// Create tickets - unique constraint will prevent double-booking
		var tickets []*models.TicketResponse
		eventIDStr := strconv.Itoa(eventDate.EventID)
		userIDStr := strconv.Itoa(req.UserID)

		for _, seatReq := range req.Seats {
			// Get seat info for label
			price, ticketTypeID, err := s.inventoryRepo.GetSeatPriceAndTicketType(req.EventDateID, seatReq.SeatID)
			if err != nil {
				return err
			}
			_ = price

			ticketID, err := s.bookingRepo.CreateTicket(tx, int(orderID), userIDStr, eventIDStr, req.EventDateID, ticketTypeID, seatReq.SeatID, req.CustomerName)
			if err != nil {
				// Check if it's a unique constraint violation
				if isUniqueConstraintError(err) {
					return fmt.Errorf("%w: One or more selected seats are no longer available", ErrSeatAlreadyTaken)
				}
				return err
			}

			// Get seat label (section + row + number)
			seat, err := s.seatRepo.GetSeatByID(seatReq.SeatID)
			seatLabel := fmt.Sprintf("Seat-%d", seatReq.SeatID)
			if err == nil && seat != nil {
				seatLabel = seat.Section + seat.Row + seat.Number
			}

			// Get ticket type name
			ticketType, err := s.ticketTypeRepo.GetTicketTypeByID(ticketTypeID)
			ticketTypeName := fmt.Sprintf("TicketType-%d", ticketTypeID)
			if err == nil && ticketType != nil {
				ticketTypeName = ticketType.Name
			}

			tickets = append(tickets, &models.TicketResponse{
				ID:         int(ticketID),
				TicketType: ticketTypeName,
				SeatLabel:  &seatLabel,
				ToName:     req.CustomerName,
			})
		}

		response = &models.BookingResponse{
			OrderID:     int(orderID),
			TotalAmount: totalAmount,
			Tickets:     tickets,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}

// isUniqueConstraintError checks if an error is a unique constraint violation
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "Duplicate entry") || 
		strings.Contains(errStr, "UNIQUE constraint") || 
		strings.Contains(errStr, "uniq_ticket_eventdate_seat")
}

