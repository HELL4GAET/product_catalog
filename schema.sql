-- Таблица карточки продукта
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    price INTEGER NOT NULL CHECK (price > 0),
    available BOOLEAN NOT NULL DEFAULT TRUE,
    description TEXT,
    image_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Таблица пользователей
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

