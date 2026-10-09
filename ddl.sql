-- Active: 1791431720533@@db.jjbkjgfgysdvxpwjjjow.supabase.co@5432@postgres
-- use postgres;
CREATE Table users (
    userID INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    deposit_amount DOUBLE PRECISION NOT NULL DEFAULT 0
);

-- drop Table users;

CREATE TABLE cars (
    carID INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    transmission VARCHAR(20) NOT NULL ,
    availability INT NOT NULL CHECK (rental_cost >= 0) DEFAULT 0,
    rental_cost NUMERIC(12, 2) NOT NULL CHECK (rental_cost >= 0),
    category VARCHAR(20) NOT NULL
);

INSERT INTO cars (name, transmission,availability, rental_cost, category)
VALUES 
    ('Toyota Avanza 1.5 G','AT', 1, 450000, 'MPV'),
    ('Honda HR-V 1.5 SE','MT', 2, 650000, 'SUV'),
    ('Hyundai Ioniq 5 Signature','EV', 3, 1200000, 'EV'),
    ('Sigra X 1.2 ','MT', 3, 1200000, 'LCGC');

-- drop Table cars;
-- TRUNCATE Table cars

CREATE TABLE books (
    bookID INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    carID INT NOT NULL,
    userID INT NOT NULL,
    booking_period DATERANGE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_books_users FOREIGN KEY (userID)
        REFERENCES users(userID)
        ON DELETE RESTRICT,
    CONSTRAINT fk_books_cars FOREIGN KEY (carID)
        REFERENCES cars(carID)
        ON DELETE RESTRICT,
    CONSTRAINT check_valid_booking_period CHECK (NOT isempty(booking_period))
);

-- drop Table books;

-- history: log of user activities (top up and booking)
CREATE TABLE history (
    historyID INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    userID INT NOT NULL,
    bookID INT,
    carID INT,
    type VARCHAR(20) NOT NULL CHECK (type IN ('TOPUP', 'BOOKING')),
    amount NUMERIC(12, 2) NOT NULL CHECK (amount > 0),
    balance_after NUMERIC(12, 2) NOT NULL,
    description VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_history_users FOREIGN KEY (userID)
        REFERENCES users(userID)
        ON DELETE RESTRICT,
    CONSTRAINT fk_history_books FOREIGN KEY (bookID)
        REFERENCES books(bookID)
        ON DELETE RESTRICT,
    CONSTRAINT fk_history_cars FOREIGN KEY (carID)
        REFERENCES cars(carID)
        ON DELETE RESTRICT
);

CREATE INDEX idx_history_user_created ON history (userID, created_at DESC);

-- drop Table history;

