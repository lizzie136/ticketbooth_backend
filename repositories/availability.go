package repositories

import (
	// "database/sql"
	"ticketbooth-backend/db"
	"ticketbooth-backend/models"
)

type AvailabilityRepository struct {
	db *db.DB
}

func NewAvailabilityRepository(db *db.DB) *AvailabilityRepository {
	return &AvailabilityRepository{db: db}
}

// GetGAAvailability fetches GA tiers with remaining tickets
func (r *AvailabilityRepository) GetGAAvailability(eventDateID int) ([]*models.TierAvailability, error) {
	query := `
		SELECT 
			tt.id, tt.name, edtt.price, edtt.remaining_tickets
		FROM event_date_has_ticket_type edtt
		INNER JOIN ticket_type tt ON edtt.ticket_type_id = tt.id
		WHERE edtt.event_date_id = ?
		ORDER BY tt.id
	`

	rows, err := r.db.Query(query, eventDateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tiers []*models.TierAvailability
	for rows.Next() {
		var tier models.TierAvailability
		err := rows.Scan(&tier.ID, &tier.Name, &tier.Price, &tier.Remaining)
		if err != nil {
			return nil, err
		}
		tiers = append(tiers, &tier)
	}

	return tiers, nil
}

// GetSeatedAvailability fetches seats with availability status
func (r *AvailabilityRepository) GetSeatedAvailability(eventDateID int) ([]*models.SectionAvailability, error) {
	query := `
		SELECT 
			s.section, s.row, s.number,
			s.id as seat_id,
			edhs.price,
			tt.id as ticket_type_id, tt.name as ticket_type_name,
			CASE WHEN t.id IS NULL THEN 1 ELSE 0 END as available
		FROM event_date_has_seat edhs
		INNER JOIN seat s ON edhs.seat_id = s.id
		INNER JOIN ticket_type tt ON edhs.ticket_type_id = tt.id
		LEFT JOIN ticket t ON t.event_date_id = edhs.event_date_id AND t.seat_id = edhs.seat_id
		WHERE edhs.event_date_id = ?
		ORDER BY s.section, s.row, s.number
	`

	rows, err := r.db.Query(query, eventDateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sectionMap := make(map[string]*models.SectionAvailability)

	for rows.Next() {
		var section, row, number string
		var seatID int
		var price float64
		var ticketTypeID int
		var ticketTypeName string
		var availableInt int

		err := rows.Scan(&section, &row, &number, &seatID, &price, &ticketTypeID, &ticketTypeName, &availableInt)
		if err != nil {
			return nil, err
		}

		sec, exists := sectionMap[section]
		if !exists {
			sec = &models.SectionAvailability{
				Section: section,
				Rows:    []*models.RowAvailability{},
			}
			sectionMap[section] = sec
		}

		// Find or create row
		var rowAvail *models.RowAvailability
		for _, r := range sec.Rows {
			if r.Row == row {
				rowAvail = r
				break
			}
		}
		if rowAvail == nil {
			rowAvail = &models.RowAvailability{
				Row:   row,
				Seats: []*models.SeatAvailability{},
			}
			sec.Rows = append(sec.Rows, rowAvail)
		}

		seatLabel := section + row + number
		seatAvail := &models.SeatAvailability{
			SeatID:     seatID,
			Label:      seatLabel,
			TicketType: ticketTypeName,
			Price:      price,
			Available:  availableInt == 1,
		}
		rowAvail.Seats = append(rowAvail.Seats, seatAvail)
	}

	sections := make([]*models.SectionAvailability, 0, len(sectionMap))
	for _, section := range sectionMap {
		sections = append(sections, section)
	}

	return sections, nil
}

