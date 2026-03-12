# LK Backend (Go)

Backend for user personal account and assortment catalog with clean modular architecture, PostgreSQL, JWT access token, stateful refresh token storage and revocation.

## Stack

- Go 1.22
- HTTP router: Chi
- PostgreSQL 16
- DB driver: pgx/v5 (`pgxpool`)
- Auth: JWT (access), random refresh token + hash in DB
- Password hashing: bcrypt
- Config: `.env`
- Migrations: SQL files + `migrate/migrate` container
- Docker / Docker Compose

## Architecture

```text
cmd/app
internal/
  app/
  auth/
  config/
  domain/
  dto/
  handler/
  middleware/
  repository/
  service/
  utils/
migrations/
Dockerfile
docker-compose.yml
.env.example
```

Layers:
- `handler` - HTTP parsing/response
- `service` - business logic
- `repository` - PostgreSQL access
- `auth` - JWT, bcrypt, refresh token manager
- `middleware` - request id, logging, recovery, auth
- `config` - env loading and validation

## Run with Docker Compose

1. Create env file:

```bash
cp .env.example .env
```

2. Start services:

```bash
docker compose up --build
```

Services:
- `postgres` (5432)
- `migrator` (applies `migrations/*.sql`)
- `app` (8080)

Health check:

```bash
curl http://localhost:8080/api/v1/health
```

## Run locally (without Docker)

1. Start PostgreSQL and create DB.
2. Create `.env` (copy from `.env.example`, set `DB_HOST=localhost`).
3. Apply migrations (example with `migrate` CLI):

```bash
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/lk_db?sslmode=disable" up
```

4. Run app:

```bash
go run ./cmd/app
```

## Migrations

- `000001_init.*.sql` - auth/users
- `000002_catalog.*.sql` - catalog tables
- `000003_catalog_seed.*.sql` - initial categories/strengths/flavors

Tables:
- `users`
- `refresh_tokens`
- `product_categories`
- `tobacco_flavors`
- `tobacco_strengths`
- `products`
- `product_flavors`

## Environment variables

Required:
- `JWT_ACCESS_SECRET`
- `JWT_REFRESH_SECRET`

Main variables:

```env
APP_NAME=lk-backend
APP_ENV=development
APP_PORT=8080

DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=lk_db
DB_SSLMODE=disable

JWT_ACCESS_SECRET=super_access_secret
JWT_REFRESH_SECRET=super_refresh_secret
JWT_ACCESS_TTL_MINUTES=15
JWT_REFRESH_TTL_HOURS=720

HTTP_READ_TIMEOUT_SECONDS=10
HTTP_WRITE_TIMEOUT_SECONDS=10
HTTP_IDLE_TIMEOUT_SECONDS=60
HTTP_SHUTDOWN_TIMEOUT_SECONDS=10
```

## API routes

Base path: `/api/v1`

Health:
- `GET /health`

Auth:
- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/refresh`
- `POST /auth/logout`
- `PATCH /auth/change-password` (protected)

User:
- `GET /users/me` (protected)
- `PATCH /users/me` (protected)

Catalog public:
- `GET /categories`
- `GET /tobacco/flavors`
- `GET /tobacco/strengths`
- `GET /products`
- `GET /products/{id}`

Catalog admin (requires `admin` role):
- `GET /admin/categories`
- `POST /admin/categories`
- `PATCH /admin/categories/{id}`
- `DELETE /admin/categories/{id}`
- `GET /admin/tobacco/flavors`
- `POST /admin/tobacco/flavors`
- `PATCH /admin/tobacco/flavors/{id}`
- `DELETE /admin/tobacco/flavors/{id}`
- `GET /admin/tobacco/strengths`
- `POST /admin/tobacco/strengths`
- `PATCH /admin/tobacco/strengths/{id}`
- `DELETE /admin/tobacco/strengths/{id}`
- `GET /admin/products`
- `POST /admin/products`
- `PATCH /admin/products/{id}`
- `DELETE /admin/products/{id}`
- `PATCH /admin/products/{id}/stock`

## Error format

```json
{
  "error": {
    "code": "invalid_credentials",
    "message": "invalid email or password"
  }
}
```

Codes used:
- `validation_error`
- `email_already_exists`
- `invalid_credentials`
- `unauthorized`
- `forbidden`
- `user_not_found`
- `internal_error`
- `invalid_refresh_token`
- `category_not_found`
- `product_not_found`
- `flavor_not_found`
- `strength_not_found`
- `invalid_stock_operation`
- `insufficient_stock`

## cURL examples

Register:

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "StrongPass123",
    "first_name": "Egor",
    "last_name": "Cherkashin",
    "middle_name": "Ivanovich",
    "phone": "+79999999999",
    "age": 25
  }'
```

Login:

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "StrongPass123"
  }'
```

Refresh:

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token>"}'
```

Logout:

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token>"}'
```

Get profile:

```bash
curl http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <access_token>"
```

Update profile:

```bash
curl -X PATCH http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Egor",
    "last_name": "Cherkashin",
    "phone": "+78888888888",
    "age": 26
  }'
```

Create category (admin):

```bash
curl -X POST http://localhost:8080/api/v1/admin/categories \
  -H "Authorization: Bearer <admin_access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "snack",
    "name": "Закуски",
    "description": "Закуски к заказу"
  }'
```

Create product (admin):

```bash
curl -X POST http://localhost:8080/api/v1/admin/products \
  -H "Authorization: Bearer <admin_access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "category_id": "<hookah_tobacco_category_uuid>",
    "name": "Darkside Grape Mint",
    "description": "Табак для кальяна с виноградно-мятным вкусом",
    "price": "890.00",
    "stock_quantity": 15,
    "unit": "pack",
    "strength_id": "<medium_strength_uuid>",
    "flavor_ids": ["<grape_flavor_uuid>", "<mint_flavor_uuid>"]
  }'
```

List active products:

```bash
curl "http://localhost:8080/api/v1/products?category_code=hookah_tobacco&in_stock=true&limit=20&offset=0"
```

Update stock (admin):

```bash
curl -X PATCH http://localhost:8080/api/v1/admin/products/<product_id>/stock \
  -H "Authorization: Bearer <admin_access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "operation": "increment",
    "quantity": 5
  }'
```

## Security notes

- Passwords are hashed with `bcrypt`
- Access tokens are JWT with short TTL
- JWT claims include `user_id` and `role`
- Refresh tokens are random, hashed in DB, revocable, and rotated on refresh
- Private routes require `Authorization: Bearer <access_token>`
- Inactive users (`is_active=false`) are blocked
- Password hash and raw tokens are never returned by API
- Admin routes require `admin` role (`403 forbidden` for non-admin)

## Prepared for extension

Project structure is modular and ready for new domains:
- products / menu
- categories
- cart
- orders / order_items
- addresses
- payments
- admin and role-based access
