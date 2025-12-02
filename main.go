package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"ticketbooth-backend/db"
	"ticketbooth-backend/handlers"
	"ticketbooth-backend/repositories"
	"ticketbooth-backend/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN not set")
	}

	authSecret := os.Getenv("AUTH_SECRET")
	if authSecret == "" {
		log.Fatal("AUTH_SECRET not set")
	}

	sqlxDB, err := sqlx.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	sqlxDB.SetMaxOpenConns(20)
	sqlxDB.SetMaxIdleConns(5)
	sqlxDB.SetConnMaxLifetime(5 * time.Minute)

	if err := sqlxDB.Ping(); err != nil {
		log.Fatal(err)
	}

	database := db.New(sqlxDB)

	// Initialize repositories
	eventRepo := repositories.NewEventRepository(database)
	availabilityRepo := repositories.NewAvailabilityRepository(database)
	bookingRepo := repositories.NewBookingRepository(database)
	inventoryRepo := repositories.NewInventoryRepository(database)
	ticketTypeRepo := repositories.NewTicketTypeRepository(database)
	seatRepo := repositories.NewSeatRepository(database)
	userRepo := repositories.NewUserRepository(database)

	// Initialize services
	bookingService := services.NewBookingService(database, bookingRepo, inventoryRepo, eventRepo, ticketTypeRepo, seatRepo)

	// Initialize handlers
	eventHandler := handlers.NewEventHandler(eventRepo, availabilityRepo)
	bookingHandler := handlers.NewBookingHandler(bookingService, bookingRepo)
	userHandler := handlers.NewUserHandler(userRepo, authSecret)

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Events
		r.Get("/events", eventHandler.GetEvents)
		r.Get("/event-dates/{id}", eventHandler.GetEventDate)
		r.Get("/event-dates/{id}/availability", eventHandler.GetAvailability)

		// Bookings
		r.Post("/bookings", bookingHandler.CreateBooking)
		r.Get("/orders/{id}", bookingHandler.GetOrder)
		r.Get("/orders", bookingHandler.GetOrders)

		// Users
		r.Post("/signup", userHandler.SignUp)
		r.Post("/login", userHandler.Login)
		r.Post("/users", userHandler.CreateUser)
		r.Put("/users/{id}", userHandler.UpdateUser)
	})

	srv := &http.Server{
		Addr:    ":4000",
		Handler: r,
	}

	log.Println("listening on :4000")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
