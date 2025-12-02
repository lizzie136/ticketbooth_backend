# Setup Instructions

## Prerequisites

1. **Go** (1.21 or later)
2. **MySQL** (running locally or via Docker)
3. **Make** (for using Makefile commands)

## Database Setup

1. **Create the database** (if not already created):
   ```bash
   mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS ticketbooth;"
   ```

2. **Initialize the schema**:
   ```bash
   make db-init
   ```
   Or manually:
   ```bash
   mysql -u root -p < schema.sql
   ```

3. **Insert mock data** (optional):
   ```bash
   make db-seed
   ```
   Or manually:
   ```bash
   mysql -u root -p ticketbooth < insert_mock_data.sql
   ```

## Environment Configuration

1. **Create a `.env` file** from the example:
   ```bash
   cp env.example .env
   ```

2. **Edit `.env`** and update the `DB_DSN` with your MySQL credentials:
   ```bash
   DB_DSN=username:password@tcp(localhost:3306)/ticketbooth?parseTime=true
   ```

   Examples:
   - No password: `DB_DSN=root@tcp(localhost:3306)/ticketbooth?parseTime=true`
   - With password: `DB_DSN=root:mypassword@tcp(localhost:3306)/ticketbooth?parseTime=true`
   - Different host/port: `DB_DSN=user:pass@tcp(127.0.0.1:3307)/ticketbooth?parseTime=true`

3. **Alternative: Export directly** (if you don't want to use .env):
   ```bash
   export DB_DSN="root:password@tcp(localhost:3306)/ticketbooth?parseTime=true"
   ```

## Running the Application

1. **Install dependencies**:
   ```bash
   make deps
   ```

2. **Run the application**:
   ```bash
   make run
   ```

   The Makefile will automatically load variables from `.env` if present.

   Or run directly:
   ```bash
   go run main.go
   ```
   (Make sure `DB_DSN` is exported in your shell)

3. **The API will be available at**: `http://localhost:4000`

## API Endpoints

- `GET /health` - Health check
- `GET /api/events` - List all events
- `GET /api/event-dates/:id` - Get event date details
- `GET /api/event-dates/:id/availability` - Get availability
- `POST /api/bookings` - Create a booking
- `GET /api/orders/:id` - Get order details

## Testing

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage
```

## Troubleshooting

- **"DB_DSN not set"**: Make sure you have a `.env` file or have exported `DB_DSN` in your shell
- **Connection refused**: Check that MySQL is running and the host/port in `DB_DSN` is correct
- **Access denied**: Verify your MySQL username and password in `DB_DSN`
- **Database doesn't exist**: Run `make db-init` to create and initialize the database

