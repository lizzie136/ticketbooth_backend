package repositories

import (
	"ticketbooth-backend/db"
	"ticketbooth-backend/models"
)

type SeatRepository struct {
	db *db.DB
}

func NewSeatRepository(db *db.DB) *SeatRepository {
	return &SeatRepository{db: db}
}

// GetSeatByID fetches a seat by ID
func (r *SeatRepository) GetSeatByID(id int) (*models.Seat, error) {
	query := `SELECT id, section, row, number, is_accessible, venue_id, status FROM seat WHERE id = ?`

	var seat models.Seat
	err := r.db.QueryRow(query, id).Scan(
		&seat.ID, &seat.Section, &seat.Row, &seat.Number, &seat.IsAccessible, &seat.VenueID, &seat.Status,
	)
	if err != nil {
		return nil, err
	}

	return &seat, nil
}

