# Scanner Auth Service

A production-grade authentication service built in Go with JWT tokens, PostgreSQL, and bcrypt password hashing. Implements a clean layered architecture following SOLID principles.

## Overview

Scanner Auth provides a complete authentication backend with user registration, login, JWT token management, and account management. The service uses a three-tier architecture (Handler → Service → Repository) ensuring clean separation of concerns and easy testing.

## Architecture

```
Client
  ↓
Gin Router + Middleware
  ↓
Handlers (HTTP layer)
  ↓
Services (Business logic)
  ↓
Repositories (Data access)
  ↓
PostgreSQL
```

### Design Principles

- **SOLID Architecture**: Each layer has a single responsibility
- **Interface-based**: Dependencies are injected through interfaces
- **Testable**: Mock repositories and services for unit tests
- **Type-safe**: Proper error handling with domain errors
- **Transactional**: Multi-step operations use database transactions

## Features

### Authentication
- User registration with email validation
- Login with bcrypt password verification
- JWT access tokens (15-minute TTL)
- Refresh token rotation (7-day TTL)
- Automatic session logout

### Security
- Password hashing with bcrypt (cost 12)
- Account lockout after 5 failed login attempts (15 minutes)
- Rate limiting (10 requests/minute on auth endpoints)
- JWT signature verification (HS256)
- Soft deletes for user data (30-day grace period)

### User Management
- Retrieve user profile
- Update profile (name, phone, email)
- Change password with current password verification
- Delete account with password confirmation
- Last login tracking

## Project Structure

```
scannerAuth/
├── cmd/api/
│   ├── main.go              # Entry point
│   └── app.go               # App initialization & wiring
├── internal/
│   ├── config/
│   │   └── config.go        # Environment configuration
│   ├── domain/
│   │   ├── user.go          # User entity
│   │   ├── token.go         # RefreshToken & Claims
│   │   └── errors.go        # Domain-level errors
│   ├── dto/
│   │   ├── auth_request.go  # Register, Login, Refresh requests
│   │   ├── user_request.go  # Update, ChangePassword, Delete requests
│   │   └── response.go      # UserResponse, AuthResponse
│   ├── handler/
│   │   ├── auth_handler.go  # Auth endpoints
│   │   ├── user_handler.go  # User profile endpoints
│   │   └── health_handler.go
│   ├── middleware/
│   │   ├── auth.go          # JWT verification
│   │   ├── error.go         # Error response mapping
│   │   ├── logger.go        # Structured logging
│   │   └── ratelimit.go     # Request rate limiting
│   ├── service/
│   │   ├── auth_service.go  # Auth business logic
│   │   ├── user_service.go  # User business logic
│   │   └── interfaces.go    # Service contracts
│   ├── repository/
│   │   ├── interfaces.go    # Repository contracts
│   │   └── gormrepo/
│   │       ├── models.go    # GORM models (UserModel, RefreshTokenModel)
│   │       ├── mapper.go    # Domain ↔ Model conversion
│   │       ├── user_repo.go
│   │       ├── token_repo.go
│   │       ├── tx_manager.go
│   │       └── errors.go
│   ├── pkg/
│   │   ├── token/
│   │   │   ├── manager.go       # TokenManager interface
│   │   │   └── jwt_manager.go   # JWT implementation
│   │   ├── database/
│   │   │   ├── gorm.go          # GORM connection setup
│   │   │   └── migrate.go       # Database migrations
│   │   ├── rediscache/
│   │   │   └── redis.go         # Redis connection (optional)
│   │   └── hasher/
│   │       ├── hasher.go        # PasswordHasher interface
│   │       └── bcrypt.go        # bcrypt implementation
│   └── router/
│       └── router.go        # Route registration & middleware chain
├── go.mod
└── go.sum
```

## API Endpoints

All endpoints are prefixed with `/api/v1`.

### Authentication Endpoints

#### POST `/auth/register`
Register a new user account.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123",
  "full_name": "John Doe",
  "phone": "+919876543210"
}
```

**Response (201 Created):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "full_name": "John Doe",
  "phone": "+919876543210",
  "role": "user",
  "is_active": true,
  "created_at": "2026-10-17T16:50:21Z",
  "updated_at": "2026-10-17T16:50:21Z"
}
```

**Errors:**
- `400 VALIDATION_ERROR` - Invalid email, password, or name format
- `409 EMAIL_TAKEN` - Email already registered

---

#### POST `/auth/login`
Authenticate user and receive access + refresh tokens.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123"
}
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "base64url_encoded_token",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "full_name": "John Doe",
    "role": "user",
    "is_active": true,
    "created_at": "2026-10-17T16:50:21Z",
    "updated_at": "2026-10-17T16:50:21Z"
  }
}
```

**Errors:**
- `400 VALIDATION_ERROR` - Missing email or password
- `401 INVALID_CREDENTIALS` - Wrong email or password
- `423 ACCOUNT_LOCKED` - Account locked after 5 failed attempts

---

#### POST `/auth/refresh`
Rotate refresh token and get new access token.

**Request:**
```json
{
  "refresh_token": "base64url_encoded_token"
}
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "new_base64url_token",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": { ... }
}
```

**Errors:**
- `401 UNAUTHORIZED` - Invalid, expired, or revoked refresh token

---

#### POST `/auth/logout`
Revoke refresh token and end session.

**Request:**
```json
{
  "refresh_token": "base64url_encoded_token"
}
```

**Response (204 No Content)**

---

### User Endpoints
*All user endpoints require `Authorization: Bearer <access_token>` header*

#### GET `/users/me`
Retrieve authenticated user's profile.

**Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "full_name": "John Doe",
  "phone": "+919876543210",
  "role": "user",
  "is_active": true,
  "last_login_at": "2026-10-17T16:50:21Z",
  "created_at": "2026-10-17T16:50:21Z",
  "updated_at": "2026-10-17T16:50:21Z"
}
```

**Errors:**
- `401 UNAUTHORIZED` - Missing or invalid token
- `404 NOT_FOUND` - User not found

---

#### PUT `/users/me`
Update user profile (name, email, phone).

**Request:**
```json
{
  "full_name": "Jane Doe",
  "email": "jane@example.com",
  "phone": "+919876543211"
}
```

**Response (200 OK):** Updated user object

**Errors:**
- `400 VALIDATION_ERROR` - Invalid field format
- `401 UNAUTHORIZED` - Missing token
- `409 EMAIL_TAKEN` - New email already in use

---

#### PUT `/users/me/password`
Change user password.

**Request:**
```json
{
  "current_password": "SecurePass123",
  "new_password": "NewSecurePass456"
}
```

**Response (204 No Content)**

**Notes:**
- All refresh tokens are revoked after password change
- User must re-login on other devices

**Errors:**
- `400 VALIDATION_ERROR` - New password same as current or invalid format
- `401 UNAUTHORIZED` - Current password incorrect

---

#### DELETE `/users/me`
Soft delete user account (reversible for 30 days).

**Request:**
```json
{
  "password": "SecurePass123"
}
```

**Response (204 No Content)**

**Notes:**
- Account is soft-deleted; data is preserved
- All refresh tokens are revoked
- User cannot login until account is restored

**Errors:**
- `401 UNAUTHORIZED` - Incorrect password

---

### Health Endpoint

#### GET `/health`
Check service health status.

**Response (200 OK):**
```json
{
  "status": "ok",
  "message": "Scanner Auth service is running"
}
```

---

## Database Schema

### users table
```sql
id                   UUID (primary key)
email                VARCHAR(255) - unique index
password_hash        TEXT
full_name            VARCHAR(120)
phone                VARCHAR(20) - nullable
role                 VARCHAR(20) - default 'user'
is_active            BOOLEAN - default true
failed_login_attempts INT - default 0
locked_until         TIMESTAMPTZ - nullable
last_login_at        TIMESTAMPTZ - nullable
created_at           TIMESTAMPTZ
updated_at           TIMESTAMPTZ
deleted_at           TIMESTAMPTZ - nullable (soft delete)
```

### refresh_tokens table
```sql
id                   UUID (primary key)
user_id              UUID (foreign key → users.id)
token_hash           CHAR(64) - unique (SHA-256 hash)
family_id            UUID - for detecting token reuse
user_agent           TEXT - nullable
ip_address           INET - nullable
expires_at           TIMESTAMPTZ
revoked_at           TIMESTAMPTZ - nullable
created_at           TIMESTAMPTZ
```

## Setup & Installation

### Prerequisites
- Go 1.22+
- PostgreSQL 16+
- Redis (optional, for future enhancements)

### Environment Configuration

Create `.env` file in project root:

```env
APP_PORT=8080
GIN_MODE=debug

# Database
DB_URL=postgres://user:password@localhost:5432/scanner_auth?sslmode=disable
DB_MAX_CONNS=20

# JWT
JWT_SECRET=your-secret-key-minimum-32-bytes-long
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h
JWT_ISSUER=scanner-auth
JWT_AUDIENCE=scanner-auth

# Security
BCRYPT_COST=12

# CORS
CORS_ORIGINS=http://localhost:3000

# Migrations
RUN_MIGRATIONS=true

# Redis (optional)
REDIS_URL=redis://localhost:6379
```

### Running the Service

```bash
# Install dependencies
go mod tidy

# Run migrations and start server
go run ./cmd/api/
```

Server will start on `http://localhost:8080`

## Testing with cURL

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "TestPass123",
    "full_name": "Test User"
  }'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "TestPass123"
  }'

# Get profile (replace TOKEN with access_token from login)
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer TOKEN"

# Health check
curl http://localhost:8080/health
```

## Error Response Format

All errors follow a consistent format:

```json
{
  "code": "ERROR_CODE",
  "message": "Human-readable error message",
  "details": "Optional field-level validation details"
}
```

**Common error codes:**
- `VALIDATION_ERROR` - 400 Bad Request
- `INVALID_CREDENTIALS` - 401 Unauthorized
- `UNAUTHORIZED` - 401 Unauthorized (expired/invalid token)
- `FORBIDDEN` - 403 Forbidden
- `NOT_FOUND` - 404 Not Found
- `EMAIL_TAKEN` - 409 Conflict
- `ACCOUNT_LOCKED` - 423 Locked
- `RATE_LIMIT_EXCEEDED` - 429 Too Many Requests
- `INTERNAL_ERROR` - 500 Internal Server Error

## Implementation Notes

### Password Security
- Passwords hashed with bcrypt using cost factor 12
- Passwords must be 8-72 characters (bcrypt limit)
- Must contain at least one letter and one digit
- Hashes never stored in logs or responses

### Token Management
- Access tokens valid for 15 minutes
- Refresh tokens valid for 7 days
- Refresh token rotation on each use (old token revoked)
- Token reuse detection via family ID
- Entire token family revoked if reuse detected (security feature)

### Rate Limiting
- 10 requests per minute per IP on login/register endpoints
- Resets after 60 seconds of inactivity
- Returns 429 Too Many Requests when exceeded

### Concurrency
- Database connection pooling (max 20 connections)
- Idle connection timeout: 5 minutes
- Max connection lifetime: 30 minutes
- All queries use context for proper cancellation

## Production Considerations

1. **Switch GIN_MODE to release** for production
2. **Use strong JWT_SECRET** (minimum 32 bytes, randomly generated)
3. **Enable HTTPS** with proper TLS certificates
4. **Configure CORS_ORIGINS** to trusted domains only
5. **Set up PostgreSQL backups** and WAL archiving
6. **Monitor failed login attempts** for security
7. **Implement API logging** and monitoring
8. **Use environment variables** for all sensitive config
9. **Enable database SSL** connection in production
10. **Set up Redis** for distributed rate limiting if scaling

## Future Enhancements

- Email verification on registration
- Forgot password flow with email reset link
- OAuth2/SSO integration
- Multi-factor authentication (MFA)
- Admin user management
- Audit logging
- Session management dashboard
