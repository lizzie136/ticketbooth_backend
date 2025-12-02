package repositories

import (
	"database/sql"
	"time"

	"ticketbooth-backend/db"
	"ticketbooth-backend/models"
)

type EventRepository struct {
	db *db.DB
}

func NewEventRepository(db *db.DB) *EventRepository {
	return &EventRepository{db: db}
}

// GetAllEvents fetches all events with their dates
func (r *EventRepository) GetAllEvents() ([]*models.EventListItem, error) {
	query := `
		SELECT 
			e.id, e.slug, e.title, e.description,
			ed.id as date_id, ed.date, ed.seating_mode,
			v.name as venue_name
		FROM event e
		LEFT JOIN event_date ed ON e.id = ed.event_id
		LEFT JOIN venue v ON CAST(ed.id_venue AS UNSIGNED) = v.id
		ORDER BY e.id, ed.id
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	eventMap := make(map[int]*models.EventListItem)

	for rows.Next() {
		var eventID int
		var slug, title, description string
		var dateID sql.NullInt64
		var date sql.NullTime
		var seatingMode sql.NullString
		var venueName sql.NullString

		err := rows.Scan(&eventID, &slug, &title, &description, &dateID, &date, &seatingMode, &venueName)
		if err != nil {
			return nil, err
		}

		event, exists := eventMap[eventID]
		if !exists {
			event = &models.EventListItem{
				ID:          eventID,
				Slug:        slug,
				Title:       title,
				Description: description,
				Dates:       []*models.EventDateItem{},
			}
			eventMap[eventID] = event
		}

		if dateID.Valid {
			dateItem := &models.EventDateItem{
				ID:          int(dateID.Int64),
				SeatingMode: seatingMode.String,
				VenueName:   venueName.String,
			}
			if date.Valid {
				dateItem.Date = date.Time.Format(time.RFC3339)
			}
			event.Dates = append(event.Dates, dateItem)
		}
	}

	events := make([]*models.EventListItem, 0, len(eventMap))
	for _, event := range eventMap {
		events = append(events, event)
	}

	return events, nil
}

// GetEventDateByID fetches a single event_date with event and venue
func (r *EventRepository) GetEventDateByID(id int) (*models.EventDate, error) {
	query := `
		SELECT 
			ed.id, ed.id_venue, ed.tota_tickets, ed.event_id, ed.seating_mode, ed.date,
			e.id as event_id, e.slug, e.title, e.description,
			v.id as venue_id, v.name, v.description, v.slug, v.capacity, v.venue_type, v.accessible_weelchair
		FROM event_date ed
		INNER JOIN event e ON ed.event_id = e.id
		LEFT JOIN venue v ON CAST(ed.id_venue AS UNSIGNED) = v.id
		WHERE ed.id = ?
	`

	var eventDate models.EventDate
	var event models.Event
	var venue models.Venue
	var date sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&eventDate.ID, &eventDate.IDVenue, &eventDate.TotalTickets, &eventDate.EventID, &eventDate.SeatingMode, &date,
		&event.ID, &event.Slug, &event.Title, &event.Description,
		&venue.ID, &venue.Name, &venue.Description, &venue.Slug, &venue.Capacity, &venue.VenueType, &venue.AccessibleWheelchair,
	)
	if err != nil {
		return nil, err
	}

	if date.Valid {
		eventDate.Date = &date.Time
	}

	eventDate.Event = &event
	eventDate.Venue = &venue

	return &eventDate, nil
}

