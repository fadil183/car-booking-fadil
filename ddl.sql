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
