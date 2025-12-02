package repositories

import (
	"ticketbooth-backend/db"
	"ticketbooth-backend/models"
)

type TicketTypeRepository struct {
	db *db.DB
}

func NewTicketTypeRepository(db *db.DB) *TicketTypeRepository {
	return &TicketTypeRepository{db: db}
}

// GetTicketTypeByID fetches a ticket type by ID
func (r *TicketTypeRepository) GetTicketTypeByID(id int) (*models.TicketType, error) {
	query := `SELECT id, name, description FROM ticket_type WHERE id = ?`

	var ticketType models.TicketType
	err := r.db.QueryRow(query, id).Scan(&ticketType.ID, &ticketType.Name, &ticketType.Description)
	if err != nil {
		return nil, err
	}

	return &ticketType, nil
}

