package repositories

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"ticketbooth-backend/db"
	"ticketbooth-backend/models"
)

type BookingRepository struct {
	db *db.DB
}

func NewBookingRepository(db *db.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

// CreateOrder creates a new order
func (r *BookingRepository) CreateOrder(tx *sqlx.Tx, userID int, totalTickets int, amount string, paymentSource string) (int64, error) {
	query := "INSERT INTO `order` (User_id, total_tickets, amount, payment_source) VALUES (?, ?, ?, ?)"

	result, err := tx.Exec(query, userID, totalTickets, amount, paymentSource)
	if err != nil {
		return 0, err
	}

	orderID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return orderID, nil
}

// CreateTicket creates a new ticket
func (r *BookingRepository) CreateTicket(tx *sqlx.Tx, orderID int, userID string, eventID string, eventDateID int, ticketTypeID int, seatID int, toName string) (int64, error) {
	query := `
		INSERT INTO ticket (event_id, user_id, ticket_type_id, to_name, event_date_id, seat_id)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	var seatIDValue interface{}
	if seatID == 0 {
		seatIDValue = nil
	} else {
		seatIDValue = seatID
	}

	result, err := tx.Exec(query, eventID, userID, ticketTypeID, toName, eventDateID, seatIDValue)
	if err != nil {
		return 0, err
	}

	ticketID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Also insert into order_hast_tickets junction table
	_, err = tx.Exec("INSERT INTO order_hast_tickets (Order_id, ticket_id) VALUES (?, ?)", orderID, ticketID)
	if err != nil {
		return 0, err
	}

	return ticketID, nil
}

// GetOrderByID fetches an order with its tickets
func (r *BookingRepository) GetOrderByID(id int) (*models.Order, error) {
	// First get the order
	orderQuery := "SELECT id, User_id, total_tickets, amount, payment_source FROM `order` WHERE id = ?"

	var order models.Order
	err := r.db.QueryRow(orderQuery, id).Scan(
		&order.ID, &order.UserID, &order.TotalTickets, &order.Amount, &order.PaymentSource,
	)
	if err != nil {
		return nil, err
	}

	// Get tickets for this order
	ticketsQuery := `
		SELECT 
			t.id, t.event_id, t.user_id, t.ticket_type_id, t.to_name, t.event_date_id, t.seat_id,
			tt.id as ticket_type_id_full, tt.name as ticket_type_name,
			s.section, s.row, s.number,
			e.id as event_id_full, e.title as event_title,
			ed.date as event_date
		FROM ticket t
		INNER JOIN order_hast_tickets oht ON t.id = oht.ticket_id
		INNER JOIN ticket_type tt ON t.ticket_type_id = tt.id
		LEFT JOIN seat s ON t.seat_id = s.id
		INNER JOIN event_date ed ON t.event_date_id = ed.id
		INNER JOIN event e ON ed.event_id = e.id
		WHERE oht.Order_id = ?
	`

	rows, err := r.db.Query(ticketsQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*models.Ticket
	for rows.Next() {
		var ticket models.Ticket
		var ticketType models.TicketType
		var seatSection, seatRow, seatNumber sql.NullString
		var seatID sql.NullInt64
		var eventID int
		var eventTitle string
		var eventDate sql.NullTime

		err := rows.Scan(
			&ticket.ID, &ticket.EventID, &ticket.UserID, &ticket.TicketTypeID, &ticket.ToName, &ticket.EventDateID, &seatID,
			&ticketType.ID, &ticketType.Name,
			&seatSection, &seatRow, &seatNumber,
			&eventID, &eventTitle,
			&eventDate,
		)
		if err != nil {
			return nil, err
		}

		if seatID.Valid {
			ticket.SeatID = int(seatID.Int64)
		} else {
			ticket.SeatID = 0
		}

		ticket.TicketType = &ticketType
		if seatSection.Valid && seatRow.Valid && seatNumber.Valid {
			seat := &models.Seat{
				Section: seatSection.String,
				Row:     seatRow.String,
				Number:  seatNumber.String,
			}
			ticket.Seat = seat
		}

		// Store event info
		event := &models.Event{
			ID:    eventID,
			Title: eventTitle,
		}
		ticket.Event = event

		if eventDate.Valid {
			eventDateModel := &models.EventDate{
				ID:   ticket.EventDateID,
				Date: &eventDate.Time,
			}
			ticket.EventDate = eventDateModel
		}

		tickets = append(tickets, &ticket)
	}

	order.Tickets = tickets
	return &order, nil
}

// get all orders by user
func (r *BookingRepository) GetAllOrdersByUserID(userID string) ([]*models.Order, error) {
	query := `
		SELECT
			o.id, o.user_id, o.total_tickets, o.amount, o.payment_source, o.created_at,
			t.id, t.event_id, t.user_id, t.ticket_type_id, t.to_name, t.event_date_id, t.seat_id,
			tt.id, tt.name,
			s.section, s.row, s.number,
			e.id, e.title,
			ed.date
		FROM ` + "`order`" + ` o
		INNER JOIN order_hast_tickets oht ON o.id = oht.order_id
		INNER JOIN ticket t ON oht.ticket_id = t.id
		LEFT JOIN ticket_type tt ON t.ticket_type_id = tt.id
		LEFT JOIN seat s ON t.seat_id = s.id
		INNER JOIN event_date ed ON t.event_date_id = ed.id
		INNER JOIN event e ON ed.event_id = e.id
		WHERE o.user_id = ?
		ORDER BY o.created_at DESC, o.id, t.id
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orderMap := make(map[int]*models.Order)
	var orders []*models.Order

	for rows.Next() {
		var (
			orderID         int
			orderUserID     int
			orderTotal      sql.NullInt64
			orderAmount     sql.NullString
			orderPaymentSrc sql.NullString
			orderCreatedAt  sql.NullTime
			ticket          models.Ticket
			seatID          sql.NullInt64
			ticketType      models.TicketType
			seatSection     sql.NullString
			seatRow         sql.NullString
			seatNumber      sql.NullString
			eventID         int
			eventTitle      string
			eventDate       sql.NullTime
		)

		err := rows.Scan(
			&orderID, &orderUserID, &orderTotal, &orderAmount, &orderPaymentSrc, &orderCreatedAt,
			&ticket.ID, &ticket.EventID, &ticket.UserID, &ticket.TicketTypeID, &ticket.ToName, &ticket.EventDateID, &seatID,
			&ticketType.ID, &ticketType.Name,
			&seatSection, &seatRow, &seatNumber,
			&eventID, &eventTitle,
			&eventDate,
		)
		if err != nil {
			return nil, err
		}

		if seatID.Valid {
			ticket.SeatID = int(seatID.Int64)
		} else {
			ticket.SeatID = 0
		}

		ticket.TicketType = &ticketType
		if seatSection.Valid && seatRow.Valid && seatNumber.Valid {
			ticket.Seat = &models.Seat{
				Section: seatSection.String,
				Row:     seatRow.String,
				Number:  seatNumber.String,
			}
		}

		ticket.Event = &models.Event{
			ID:    eventID,
			Title: eventTitle,
		}

		if eventDate.Valid {
			ticket.EventDate = &models.EventDate{
				ID:   ticket.EventDateID,
				Date: &eventDate.Time,
			}
		}

		order, found := orderMap[orderID]
		if !found {
			totalTickets := 0
			if orderTotal.Valid {
				totalTickets = int(orderTotal.Int64)
			}

			order = &models.Order{
				ID:            orderID,
				UserID:        orderUserID,
				TotalTickets:  totalTickets,
				Amount:        "",
				PaymentSource: "",
				Tickets:       []*models.Ticket{},
			}
			if orderAmount.Valid {
				order.Amount = orderAmount.String
			}
			if orderPaymentSrc.Valid {
				order.PaymentSource = orderPaymentSrc.String
			}
			if orderCreatedAt.Valid {
				createdAt := orderCreatedAt.Time
				order.CreatedAt = &createdAt
			}
			orderMap[orderID] = order
			orders = append(orders, order)
		}

		order.Tickets = append(order.Tickets, &ticket)
	}

	return orders, nil
}
