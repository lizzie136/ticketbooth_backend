package repositories

import (
	//"database/sql"
	"ticketbooth-backend/db"
	"github.com/jmoiron/sqlx"
)

type InventoryRepository struct {
	db *db.DB
}

func NewInventoryRepository(db *db.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

// UpdateGATicketInventory atomically updates GA ticket inventory
// Returns the number of rows affected (should be 1 if successful, 0 if insufficient inventory)
func (r *InventoryRepository) UpdateGATicketInventory(tx *sqlx.Tx, eventDateID int, ticketTypeID int, quantity int) (int64, error) {
	query := `
		UPDATE event_date_has_ticket_type
		SET remaining_tickets = remaining_tickets - ?
		WHERE event_date_id = ?
		  AND ticket_type_id = ?
		  AND remaining_tickets >= ?
	`

	result, err := tx.Exec(query, quantity, eventDateID, ticketTypeID, quantity)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

// GetGATicketPriceAndRemaining gets the price and remaining count for a GA ticket type
func (r *InventoryRepository) GetGATicketPriceAndRemaining(eventDateID int, ticketTypeID int) (float64, int, error) {
	query := `
		SELECT price, remaining_tickets
		FROM event_date_has_ticket_type
		WHERE event_date_id = ? AND ticket_type_id = ?
	`

	var price float64
	var remaining int
	err := r.db.QueryRow(query, eventDateID, ticketTypeID).Scan(&price, &remaining)
	if err != nil {
		return 0, 0, err
	}

	return price, remaining, nil
}

// CheckSeatAvailability checks if seats are available (not already booked)
func (r *InventoryRepository) CheckSeatAvailability(eventDateID int, seatIDs []int) ([]int, error) {
	if len(seatIDs) == 0 {
		return []int{}, nil
	}

	query := `
		SELECT seat_id
		FROM ticket
		WHERE event_date_id = ? AND seat_id IN (?)
	`

	query, args, err := sqlx.In(query, eventDateID, seatIDs)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookedSeatIDs []int
	for rows.Next() {
		var seatID int
		if err := rows.Scan(&seatID); err != nil {
			return nil, err
		}
		bookedSeatIDs = append(bookedSeatIDs, seatID)
	}

	return bookedSeatIDs, nil
}

// GetSeatPriceAndTicketType gets the price and ticket type for a seat
func (r *InventoryRepository) GetSeatPriceAndTicketType(eventDateID int, seatID int) (float64, int, error) {
	query := `
		SELECT price, ticket_type_id
		FROM event_date_has_seat
		WHERE event_date_id = ? AND seat_id = ?
	`

	var price float64
	var ticketTypeID int
	err := r.db.QueryRow(query, eventDateID, seatID).Scan(&price, &ticketTypeID)
	if err != nil {
		return 0, 0, err
	}

	return price, ticketTypeID, nil
}

