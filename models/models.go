package models

import "time"

// Database Models

type Event struct {
	ID          int    `db:"id" json:"id"`
	Slug        string `db:"slug" json:"slug"`
	Title       string `db:"title" json:"title"`
	Description string `db:"description" json:"description"`
}

type Venue struct {
	ID                   int    `db:"id" json:"id"`
	Name                 string `db:"name" json:"name"`
	Description          string `db:"description" json:"description"`
	Slug                 string `db:"slug" json:"slug"`
	Capacity             int    `db:"capacity" json:"capacity"`
	VenueType            string `db:"venue_type" json:"venueType"`
	AccessibleWheelchair bool   `db:"accessible_weelchair" json:"accessibleWheelchair"`
}

type EventDate struct {
	ID           int        `db:"id" json:"id"`
	IDVenue      string     `db:"id_venue" json:"-"`
	TotalTickets int        `db:"tota_tickets" json:"-"`
	EventID      int        `db:"event_id" json:"-"`
	SeatingMode  string     `db:"seating_mode" json:"seatingMode"`
	Date         *time.Time `db:"date" json:"date"`
	// Joined fields
	Event *Event `json:"event,omitempty"`
	Venue *Venue `json:"venue,omitempty"`
}

type TicketType struct {
	ID          int    `db:"id" json:"id"`
	Name        string `db:"name" json:"name"`
	Description string `db:"description" json:"description"`
}

type Seat struct {
	ID           int    `db:"id" json:"id"`
	Section      string `db:"section" json:"section"`
	Row          string `db:"row" json:"row"`
	Number       string `db:"number" json:"number"`
	IsAccessible bool   `db:"is_accessible" json:"isAccessible"`
	VenueID      int    `db:"venue_id" json:"-"`
	Status       string `db:"status" json:"status"`
}

type EventDateHasTicketType struct {
	EventDateID          int        `db:"event_date_id" json:"-"`
	TicketTypeID         int        `db:"ticket_type_id" json:"-"`
	MaxQuantity          int        `db:"max_quantity" json:"-"`
	RemainingTickets     int        `db:"remaining_tickets" json:"remaining"`
	Price                float64    `db:"price" json:"price"`
	NumberPeopleIncluded int        `db:"number_people_included" json:"-"`
	ExpirationDate       *time.Time `db:"expiration_date" json:"-"`
	// Joined fields
	TicketType *TicketType `json:"ticketType,omitempty"`
}

type EventDateHasSeat struct {
	EventDateID  int     `db:"event_date_id" json:"-"`
	SeatID       int     `db:"seat_id" json:"-"`
	Price        float64 `db:"price" json:"price"`
	TicketTypeID int     `db:"ticket_type_id" json:"-"`
	// Joined fields
	Seat       *Seat       `json:"seat,omitempty"`
	TicketType *TicketType `json:"ticketType,omitempty"`
}

type Order struct {
	ID            int        `db:"id" json:"id"`
	UserID        int        `db:"User_id" json:"userId"`
	TotalTickets  int        `db:"total_tickets" json:"totalTickets"`
	Amount        string     `db:"amount" json:"amount"`
	PaymentSource string     `db:"payment_source" json:"paymentSource"`
	CreatedAt     *time.Time `db:"created_at" json:"createdAt,omitempty"`
	// Joined fields
	Tickets []*Ticket `json:"tickets,omitempty"`
}

type Ticket struct {
	ID           int    `db:"id" json:"id"`
	EventID      string `db:"event_id" json:"-"`
	UserID       string `db:"user_id" json:"-"`
	TicketTypeID int    `db:"ticket_type_id" json:"-"`
	ToName       string `db:"to_name" json:"toName"`
	EventDateID  int    `db:"event_date_id" json:"-"`
	SeatID       int    `db:"seat_id" json:"-"`
	// Joined fields
	TicketType *TicketType `json:"ticketType,omitempty"`
	Seat       *Seat       `json:"seat,omitempty"`
	Event      *Event      `json:"event,omitempty"`
	EventDate  *EventDate  `json:"eventDate,omitempty"`
}

type User struct {
	ID             int        `db:"id" json:"id"`
	Username       string     `db:"username" json:"username"`
	FirstName      string     `db:"name" json:"firstName"`
	LastName       string     `db:"last_name" json:"lastName"`
	Email          string     `db:"email" json:"email"`
	Password       string     `db:"-" json:"-"`
	HashedPassword string     `db:"hashed_password" json:"-"`
	CreatedAt      *time.Time `db:"date_created" json:"createdAt,omitempty"`
	UpdatedAt      *time.Time `db:"date_updated" json:"updatedAt,omitempty"`
}

// API Request/Response DTOs

type EventDateResponse struct {
	ID          int        `json:"id"`
	Event       *EventInfo `json:"event"`
	Date        string     `json:"date"`
	Venue       *VenueInfo `json:"venue"`
	SeatingMode string     `json:"seatingMode"`
}

type EventInfo struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type VenueInfo struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
}

type EventListItem struct {
	ID          int              `json:"id"`
	Slug        string           `json:"slug"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Dates       []*EventDateItem `json:"dates"`
}

type EventDateItem struct {
	ID          int    `json:"id"`
	Date        string `json:"date"`
	VenueName   string `json:"venueName"`
	SeatingMode string `json:"seatingMode"`
}

type GAAvailabilityResponse struct {
	SeatingMode string              `json:"seatingMode"`
	Tiers       []*TierAvailability `json:"tiers"`
}

type TierAvailability struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Remaining int     `json:"remaining"`
}

type SeatedAvailabilityResponse struct {
	SeatingMode string                 `json:"seatingMode"`
	Sections    []*SectionAvailability `json:"sections"`
}

type SectionAvailability struct {
	Section string             `json:"section"`
	Rows    []*RowAvailability `json:"rows"`
}

type RowAvailability struct {
	Row   string              `json:"row"`
	Seats []*SeatAvailability `json:"seats"`
}

type SeatAvailability struct {
	SeatID     int     `json:"seatId"`
	Label      string  `json:"label"`
	TicketType string  `json:"ticketType"`
	Price      float64 `json:"price"`
	Available  bool    `json:"available"`
}

type BookingRequest struct {
	EventDateID   int                   `json:"eventDateId"`
	CustomerName  string                `json:"customerName"`
	PaymentSource string                `json:"paymentSource"`
	UserID        int                   `json:"userId"`
	Tiers         []*TierBookingRequest `json:"tiers,omitempty"` // For GA
	Seats         []*SeatBookingRequest `json:"seats,omitempty"` // For seated
}

type TierBookingRequest struct {
	TicketTypeID int `json:"ticketTypeId"`
	Quantity     int `json:"quantity"`
}

type SeatBookingRequest struct {
	SeatID       int `json:"seatId"`
	TicketTypeID int `json:"ticketTypeId"`
}

type BookingResponse struct {
	OrderID     int               `json:"orderId"`
	TotalAmount float64           `json:"totalAmount"`
	Tickets     []*TicketResponse `json:"tickets"`
}

type TicketResponse struct {
	ID         int     `json:"id"`
	TicketType string  `json:"ticketType"`
	SeatLabel  *string `json:"seatLabel"`
	ToName     string  `json:"toName"`
}

type OrderResponse struct {
	ID           int                    `json:"id"`
	CreatedAt    string                 `json:"createdAt"`
	CustomerName string                 `json:"customerName"`
	TotalAmount  float64                `json:"totalAmount"`
	Tickets      []*OrderTicketResponse `json:"tickets"`
}

type OrderTicketResponse struct {
	ID         int     `json:"id"`
	EventTitle string  `json:"eventTitle"`
	EventDate  string  `json:"eventDate"`
	TicketType string  `json:"ticketType"`
	SeatLabel  *string `json:"seatLabel"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type CreateUserRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

type UpdateUserRequest struct {
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	Email     *string `json:"email"`
	Username  *string `json:"username"`
}

type UserResponse struct {
	ID        int     `json:"id"`
	Username  string  `json:"username"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Email     string  `json:"email"`
	CreatedAt *string `json:"createdAt,omitempty"`
	UpdatedAt *string `json:"updatedAt,omitempty"`
}

type LoginRequest struct {
	Email    *string `json:"email"`
	Username *string `json:"username"`
	Password string  `json:"password"`
}

type LoginResponse struct {
	Token string        `json:"token"`
	User  *UserResponse `json:"user"`
}
