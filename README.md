# Car Booking API

REST API untuk penyewaan mobil: registrasi dan login user, top up deposit, melihat mobil yang tersedia, memesan mobil, dan melihat riwayat transaksi.

## Tech Stack

- Go 1.26
- [Echo v4](https://echo.labstack.com/) (HTTP framework)
- [GORM](https://gorm.io/) + PostgreSQL (Supabase)
- [godotenv](https://github.com/joho/godotenv) untuk membaca `.env`

## Fitur

- Register dan login. Password disimpan sebagai hash SHA-256 dengan salt acak.
- Top up deposit, dicatat ke tabel `history`.
- Daftar mobil dengan `availability > 0`.
- Booking mobil per rentang tanggal. Stok, deposit, `books`, dan `history` diperbarui dalam satu transaksi.
- Riwayat top up dan booking per user.

## Struktur Project

```
.
├── main.go                         # entry point, server di :8080
├── config/db.go                    # koneksi PostgreSQL
├── app/
│   ├── handler/routes.go           # definisi route
│   ├── handler/encryption/sha.go   # hash dan verifikasi password
│   └── repository/user_repository.go  # handler + query database
├── ddl.sql                         # skema tabel dan data awal mobil
├── API_DOCS.txt                    # dokumentasi lengkap request/response
└── example.env                     # contoh konfigurasi environment
```

## Menjalankan

### 1. Prasyarat

- Go 1.26 atau lebih baru
- Database PostgreSQL

### 2. Siapkan database

Jalankan isi [ddl.sql](ddl.sql) pada database Anda (membuat tabel `users`, `cars`, `books`, `history` dan mengisi contoh data mobil).

### 3. Konfigurasi environment

Salin `example.env` menjadi `.env`, lalu isi nilainya:

```
DB_PSQL_HOST=
DB_PSQL_PORT=
DB_PSQL_USR=
DB_PSQL_PW=
DB_PSQL_NAME=
```

### 4. Jalankan server

```
go run .
```

Server berjalan di `http://localhost:8080`.

## Endpoint

Semua endpoint memakai method `POST` dengan body JSON. Tidak ada token; endpoint yang butuh identitas user (`/topup`, `/books`, `/history`) menerima `email` dan `password` di body.

| Endpoint    | Fungsi                                  | Body                                                  |
|-------------|-----------------------------------------|-------------------------------------------------------|
| `/register` | Daftar user baru                        | `email`, `password`                                   |
| `/login`    | Verifikasi email dan password           | `email`, `password`                                   |
| `/topup`    | Tambah deposit                          | `email`, `password`, `amount`                         |
| `/rent`     | Daftar mobil yang tersedia              | tidak ada                                             |
| `/books`    | Pesan mobil                             | `email`, `password`, `car_id`, `start_date`, `end_date` |
| `/history`  | Riwayat top up dan booking              | `email`, `password`                                   |

Contoh:

```
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

Detail parameter, contoh response, dan semua kode error ada di [API_DOCS.txt](API_DOCS.txt).

## Aturan Booking

- Format tanggal `YYYY-MM-DD`; periode dihitung `[start_date, end_date)`.
- `start_date` tidak boleh di masa lalu dan `end_date` harus setelah `start_date`.
- Total biaya = `rental_cost` x jumlah hari, dipotong dari deposit.
- Setiap booking mengurangi `availability` mobil sebanyak 1.

## Skema Database

| Tabel     | Kolom                                                                                 |
|-----------|---------------------------------------------------------------------------------------|
| `users`   | `userID`, `email`, `password`, `deposit_amount`                                       |
| `cars`    | `carID`, `name`, `transmission`, `availability`, `rental_cost`, `category`            |
| `books`   | `bookID`, `carID`, `userID`, `booking_period`, `created_at`, `updated_at`             |
| `history` | `historyID`, `userID`, `bookID`, `carID`, `type`, `amount`, `balance_after`, `description`, `created_at` |

## Catatan

- Belum ada autentikasi berbasis sesi atau token; kirim kredensial pada setiap request yang membutuhkannya dan gunakan HTTPS saat deploy.
- SHA-256 dengan salt lebih aman daripada tanpa salt, tetapi `bcrypt` atau `argon2` lebih disarankan untuk produksi.

