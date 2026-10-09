-- Active: 1791431720533@@db.jjbkjgfgysdvxpwjjjow.supabase.co@5432@postgres
-- use postgres;
CREATE Table users (
    userID INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    deposit_amount DOUBLE PRECISION NOT NULL DEFAULT 0
);

-- drop Table users;
