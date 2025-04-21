# Go Auth Service 🚀

Go Auth Service is a lightweight, modular authentication service written in Go. It provides secure user authentication using JWT tokens, rate limiting, and PostgreSQL as the database backend. This module can be easily integrated into your Go projects to handle user authentication and session management. 😊

## Features ✨

- **User Authentication**: Secure user registration and login with hashed passwords (bcrypt).
- **JWT Tokens**: Stateless authentication using JSON Web Tokens. 🔐
- **Rate Limiting**: Protect endpoints with IP-based rate limiting. 🚦
- **PostgreSQL Integration**: Store user data and sessions securely in a PostgreSQL database. 🗄️
- **Account Locking**: Lock accounts after multiple failed login attempts to prevent brute force attacks. 🚫
- **Session Management**: Track and revoke active sessions. 🔄

## Getting Started 🛠️

### Prerequisites 📋

- Go 1.20+ installed on your system.
- A running PostgreSQL instance.
- Environment variables configured in a `.env` file or your system.

### Installation 📦

1. Clone the repository:

   ```bash
   git clone https://github.com/Stewz00/go-auth-service.git
   cd go-auth-service
   ```

2. Install dependencies:

   ```bash
   go mod tidy
   ```

3. Set up the database:

   - Create a PostgreSQL database (e.g., `authdb`).
   - Run the SQL schema to create the necessary tables:
     ```bash
     psql -U <username> -d authdb -f internal/database/schema.sql
     ```

4. Configure environment variables in a `.env` file:
   ```env
   PORT=8080
   JWT_SECRET=mysecretkey
   DATABASE_URL=postgres://<username>:<password>@localhost:5432/authdb?sslmode=disable
   ```

### Usage 🚀

#### Running the Service 🏃‍♂️

Start the service by running the following command:

```bash
go run cmd/server/main.go
```

The service will start on the port specified in the `.env` file (default: `8080`).

#### API Endpoints 🌐

| Endpoint         | Method | Description                         | Rate Limit              |
| ---------------- | ------ | ----------------------------------- | ----------------------- |
| `/health`        | GET    | Health check endpoint               | 100 requests/min per IP |
| `/auth/register` | POST   | Register a new user                 | 10 requests/min per IP  |
| `/auth/login`    | POST   | Authenticate a user and get a token | 10 requests/min per IP  |
| `/auth/logout`   | POST   | Revoke the user's active session    | 100 requests/min per IP |

#### Example Requests 📬

1. **Register a User**:

   ```bash
   curl -X POST http://localhost:8080/auth/register \
   -H "Content-Type: application/json" \
   -d '{"email": "user@example.com", "password": "securepassword"}'
   ```

2. **Login**:

   ```bash
   curl -X POST http://localhost:8080/auth/login \
   -H "Content-Type: application/json" \
   -d '{"email": "user@example.com", "password": "securepassword"}'
   ```

   Response:

   ```json
   {
     "token": "your-jwt-token"
   }
   ```

3. **Logout**:
   ```bash
   curl -X POST http://localhost:8080/auth/logout \
   -H "Authorization: Bearer your-jwt-token"
   ```

### Integration into Your Project 🤝

To use this module in your Go project:

1. Import the necessary packages:

   ```go
   import (
       "github.com/Stewz00/go-auth-service/internal/handler"
       "github.com/Stewz00/go-auth-service/internal/service"
       "github.com/Stewz00/go-auth-service/internal/repository"
   )
   ```

2. Initialize the service and handlers:

   ```go
   db, _ := database.New("your-database-url")
   userRepo := repository.NewUserRepository(db)
   authService := service.NewAuthService(userRepo, "your-jwt-secret")
   authHandler := handler.NewAuthHandler(authService)
   ```

3. Add the routes to your router:
   ```go
   r := chi.NewRouter()
   r.Post("/auth/register", authHandler.Register)
   r.Post("/auth/login", authHandler.Login)
   r.Post("/auth/logout", authHandler.Logout)
   ```

### Testing 🧪

The service includes both unit tests and integration tests to ensure reliability and correctness.

#### Running Unit Tests

Unit tests use mock implementations and can be run without a database:

```bash
go test ./internal/service/... -v
```

#### Running Integration Tests

Integration tests require a PostgreSQL test database. Before running:

1. Create a test database:

   ```bash
   createdb -U postgres authdb_test
   ```

2. Apply the schema to the test database:

   ```bash
   psql -U postgres -d authdb_test -f internal/database/schema.sql
   ```

3. Run the integration tests:
   ```bash
   go test ./internal/test/integration/... -v
   ```

#### Test Coverage

To run tests with coverage reporting:

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

#### What's Tested 🎯

1. **Unit Tests**:

   - User registration
   - Login authentication
   - JWT token validation
   - Session management
   - Account lockout mechanism

2. **Integration Tests**:
   - Complete authentication flow (register → login → logout)
   - Failed login attempts and account locking
   - Token revocation
   - Database interactions
   - API endpoint responses

#### Running with Test Environment

For development, you can use the test environment configuration:

```env
PORT=8080
JWT_SECRET=test-secret
DATABASE_URL=postgres://postgres:postgres@localhost:5432/authdb_test?sslmode=disable
```

### Security Features 🔒

- **Password Hashing**: Passwords are hashed using bcrypt with a cost factor of 12.
- **JWT Tokens**: Tokens are signed with a secret key and include expiration and unique IDs for session tracking.
- **Rate Limiting**: Protects endpoints from abuse with IP-based rate limiting.
- **Account Locking**: Accounts are locked after 5 failed login attempts.

### Limitations ⚠️

While this service provides a robust foundation for authentication, it has some limitations that you should be aware of:

1. **Single Database Backend**:

   - The service currently supports only PostgreSQL as the database backend. Adding support for other databases would require additional implementation.

2. **Basic Rate Limiting**:

   - Rate limiting is IP-based, which may not be effective in scenarios where users share the same IP (e.g., behind a corporate proxy).

3. **No Multi-Factor Authentication (MFA)**:

   - The service does not include MFA, which is a critical feature for high-security applications.

4. **No Email Verification**:

   - User accounts are created without email verification, which could allow the use of invalid or fake email addresses.

5. **Limited Token Revocation**:

   - JWT tokens are stateless, and while sessions are tracked in the database, already-issued tokens cannot be forcefully invalidated until they expire.

6. **Password Reset**:

   - The service does not include a password reset mechanism. This would need to be implemented separately.

7. **Scaling Considerations**:

   - The service is designed for small to medium-scale applications. For high-scale systems, additional optimizations (e.g., distributed rate limiting, caching) may be required.

8. **No HTTPS Enforcement**:

   - The service does not enforce HTTPS. It is recommended to deploy it behind a reverse proxy (e.g., NGINX) with HTTPS enabled.

9. **Limited Logging and Monitoring**:

   - The service includes basic logging but lacks advanced monitoring and alerting features.

10. **No Role-Based Access Control (RBAC)**:
    - The service does not include role-based access control or permissions management. This would need to be added for more complex applications.

### Development 🧑‍💻

To run the service locally for development:

1. Start the PostgreSQL database.
2. Run the service:
   ```bash
   go run cmd/server/main.go
   ```

### License 📜

This project is licensed under the MIT License.

---

Feel free to contribute or raise issues in the repository. Happy coding! 😊
