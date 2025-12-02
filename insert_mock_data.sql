-- insert_mock_data.sql
-- Mock data for ticketing system

USE ticketbooth;

-- Ticket Types
INSERT INTO ticket_type
  (id, name, description, based_on, contant_name, contact_email)
VALUES
  (1, 'General Admission', 'Standard paid entry', 'payment', 'Box Office', 'tickets@example.com'),
  (2, 'Reserved Seating',  'Paid ticket with assigned seat', 'payment', 'Box Office', 'tickets@example.com'),
  (3, 'Free Ticket',       'Complimentary ticket', 'payment', 'Promotions', 'promo@example.com'),
  (4, 'Donation Ticket',   'Pay-what-you-want', 'payment', 'Fundraising', 'donations@example.com'),
  (5, 'Early Bird',        'Discounted early purchase', 'time_purchase', 'Marketing', 'marketing@example.com'),
  (6, 'Ultra Early Bird',  'Very early discount', 'time_purchase', 'Marketing', 'marketing@example.com'),
  (7, 'Pass', 'Generic pass', 'validity', 'Box Office', 'tickets@example.com'),
  (8, 'Virtual Pass', 'Online access', 'validity', 'Virtual Team', 'virtual@example.com'),
  (9, 'One-day Pass', 'Single-day access', 'validity', 'Box Office', 'tickets@example.com'),
  (10, 'Multi-day Pass', 'Multiple-day access', 'validity', 'Box Office', 'tickets@example.com'),
  (11, 'Session Entry', 'Single session', 'validity', 'Programming', 'sessions@example.com'),
  (12, 'All-Access Pass', 'All days', 'validity', 'Programming', 'allaccess@example.com'),
  (13, 'Group Ticket', 'Group bundle', 'special', 'Group Sales', 'groups@example.com'),
  (14, 'VIP Ticket', 'Premium perks', 'special', 'VIP Desk', 'vip@example.com'),
  (15, 'Member Ticket', 'Members only', 'special', 'Membership', 'members@example.com'),
  (16, 'Targeted Discount', 'Audience segment discount', 'discount', 'Marketing', 'marketing@example.com'),
  (17, 'Flash Sale', 'Time-limited sale', 'discount', 'Marketing', 'marketing@example.com'),
  (18, 'Coded Discount', 'Promo code required', 'discount', 'Marketing', 'marketing@example.com'),
  (19, 'Themed Promo Code Ticket', 'Themed promo', 'discount', 'Marketing', 'marketing@example.com');

-- Venue
INSERT INTO venue (id, name, description, slug, capacity, venue_type, accessible_weelchair)
VALUES (1, 'Downtown Arena', 'Main arena', 'downtown-arena', 10000, 'indoor', 1);

-- Event
INSERT INTO event (id, slug, title, description)
VALUES (1, 'summer-fest', 'Summer Fest 2025', 'Two nights of music.');

-- Event Dates
INSERT INTO event_date (id, id_venue, tota_tickets, event_id, seating_mode, date)
VALUES
  (1, '1', 5000, 1, 'GA', '2025-07-15 20:00:00'),
  (2, '1', 5000, 1, 'SEATED', '2025-07-16 20:00:00');

-- GA Inventory
INSERT INTO event_date_has_ticket_type
  (event_date_id, ticket_type_id, max_quantity, remaining_tickets, price, number_people_included, expiration_date)
VALUES
  (1, 1, 3000, 3000, 10.00, 1, NULL),
  (1, 5,  500,  500,  8.00, 1, '2025-06-01 23:59:59'),
  (1, 14, 200, 200, 50.00, 1, NULL);

-- Seats
INSERT INTO seat (id, section, `row`, `number`, is_accessible, venue_id)
VALUES
  (1, 'A', '1', '1', 0, 1),
  (2, 'A', '1', '2', 0, 1),
  (3, 'A', '1', '3', 0, 1),
  (4, 'B', '1', '1', 1, 1);

-- Seat inventory
INSERT INTO event_date_has_seat (event_date_id, seat_id, price, ticket_type_id)
VALUES
  (2, 1, 100.00, 14),
  (2, 2, 100.00, 14),
  (2, 3, 50.00,  1),
  (2, 4, 50.00, 15);

-- Users
INSERT INTO user (id, username, name, last_name, email, hashed_password, date_created, date_updated)
VALUES
  (1, 'aliceex', 'Alice', 'Example', 'alice123@gmail.com', '52fd0e88bc0677bf4e30963621cb33181947b00eb237cff1b156d333c9a1db6d', NOW(), NOW()),
  (2, 'bobbyo', 'Bob', 'Organizer', 'organizer@example.com', '1c3cfcc72db6b55b814afbfd8a53163b961e76e743ed81d35cf573f88f738c93', NOW(), NOW());

-- Order
INSERT INTO `order` (id, User_id, total_tickets, amount, payment_source)
VALUES (1, 1, 2, '200.00', 'test-card-4242');

-- Tickets
INSERT INTO ticket (id, event_id, user_id, ticket_type_id, to_name, event_date_id, seat_id)
VALUES
  (1, '1', '1', 14, 'Alice Example', 2, 1),
  (2, '1', '1', 14, 'Alice Example', 2, 2);

-- Order Tickets Junction
INSERT INTO order_hast_tickets (Order_id, ticket_id)
VALUES (1, 1), (1, 2);

-- Roles
INSERT INTO role (id, name, description)
VALUES
  (1, 'ADMIN', 'Full access to manage events, venues, and tickets'),
  (2, 'ORGANIZER', 'Can manage events and tickets for their shows'),
  (3, 'VIEWER', 'Read-only access to events and basic reports');

-- Permissions
INSERT INTO permission (id, code, description)
VALUES
  (1, 'EVENT_READ', 'View events and event dates'),
  (2, 'EVENT_WRITE', 'Create and update events and event dates'),
  (3, 'VENUE_READ', 'View venues'),
  (4, 'VENUE_WRITE', 'Create and update venues'),
  (5, 'TICKET_MANAGE', 'Create and manage ticket inventory and bookings'),
  (6, 'REPORT_VIEW', 'View high-level reports and stats');

-- Role to permission mapping
INSERT INTO role_has_permission (role_id, permission_id)
VALUES
  -- ADMIN gets everything
  (1, 1),
  (1, 2),
  (1, 3),
  (1, 4),
  (1, 5),
  (1, 6),

  -- ORGANIZER: event + ticket management, basic read
  (2, 1),
  (2, 2),
  (2, 3),
  (2, 5),

  -- VIEWER: read-only
  (3, 1),
  (3, 3),
  (3, 6);

-- Assign roles to users
-- Here we give Alice (user.id = 1) the ORGANIZER role by default.
INSERT INTO role_has_user (role_id, user_id)
VALUES
  (2, 1),
  (1, 2);
