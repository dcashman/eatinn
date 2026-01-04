# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

EatInn is a production-ready REST API for managing recipes, built with Go and PostgreSQL. The application provides full CRUD operations for recipes with advanced filtering, user authentication, and background email processing.

**Current Status**: ✅ **Production-Ready MVP**
- 8 fully operational API endpoints
- Complete user authentication with activation tokens
- Recipe management with complex relational data
- IP-based rate limiting and graceful shutdown
- Background email sending with proper cleanup

**Implementation Status**: Chapters 1-15 of "Let's Go Further" are complete. The application is ready for deployment and use.

**Important**: This project follows the patterns and best practices from "Let's Go Further" by Alex Edwards. Both "Let's Go" and "Let's Go Further" are available at `~/projects/docs/lets-go/html/` and `~/projects/docs/lets-go-further/html/` for reference. When patterns conflict, prefer "Let's Go Further" guidance.

## Development Commands

### Running the Application

```bash
go run ./cmd/api -db-dsn=$EATINN_DB_DSN
```

**Server Configuration Flags:**
- `-port`: API server port (default: 4000)
- `-env`: Environment (development|staging|production, default: development)

**Database Configuration Flags:**
- `-db-dsn`: PostgreSQL DSN (default: $EATINN_DB_DSN env var)
- `-db-max-open-conns`: PostgreSQL max open connections (default: 25)
- `-db-max-idle-conns`: PostgreSQL max idle connections (default: 25)
- `-db-max-idle-time`: PostgreSQL max connection idle time (default: 15m)

**Rate Limiting Configuration Flags:**
- `-limiter-rps`: Maximum requests per second (default: 2)
- `-limiter-burst`: Maximum burst size (default: 4)
- `-limiter-enabled`: Enable rate limiter (default: true)

**SMTP Configuration Flags:**
- `-smtp-host`: SMTP server host (default: sandbox.smtp.mailtrap.io)
- `-smtp-port`: SMTP server port (default: 2525)
- `-smtp-username`: SMTP username (default: test credentials)
- `-smtp-password`: SMTP password (default: test credentials)
- `-smtp-sender`: Email sender address (default: EatInn <no-reply@eatinn.dcashman.net>)

### Database Setup

The application expects PostgreSQL connection via the `EATINN_DB_DSN` environment variable.

Example DSN format:
```
postgres://username:password@localhost/eatinn?sslmode=disable
```

Database migrations are located in the `migrations/` directory and should be applied in order.

### Testing

Test data is available in `test/test_recipes.json` for manual API testing.

## Architecture

### Project Structure

```
cmd/api/              - HTTP server and handlers
  main.go             - Application entry point, config, DB setup, WaitGroup
  server.go           - HTTP server with graceful shutdown
  routes.go           - Route definitions using julienschmidt/httprouter
  recipes.go          - Recipe CRUD handlers (all implemented)
  users.go            - User registration and activation handlers
  tokens.go           - Authentication token generation handler
  context.go          - Request context helpers for user management
  helpers.go          - JSON encoding/decoding, background tasks, parameter extraction
  errors.go           - Centralized error response handlers (11 types)
  middleware.go       - HTTP middleware (panic recovery, rate limiting, authentication)
  healthcheck.go      - Health check endpoint

internal/
  data/               - Data layer (models and database access)
    models.go         - Models struct, custom Duration type, error definitions
    recipes.go        - Recipe model with full CRUD operations
    users.go          - User model with password hashing (bcrypt cost 12)
    tokens.go         - Token model with cryptographic token generation
    filters.go        - Pagination and sorting support
  validator/          - Input validation utilities
    validator.go      - Validator type and helper functions
  mailer/             - Email sending functionality
    mailer.go         - SMTP mailer with template support
    templates/        - Embedded email templates

migrations/           - SQL database migrations (4 migrations)
test/                 - Test data and fixtures
bin/                  - Compiled binaries
```

### Key Architectural Patterns

**Application Context Pattern**: The `application` struct in `cmd/api/main.go:46` is the central dependency container that holds:
- Configuration (server, database, rate limiting, SMTP)
- Structured logger (slog)
- Data models (Recipes, Users, Tokens)
- Mailer instance
- WaitGroup for tracking background goroutines

All handlers are methods on this struct, providing access to dependencies without globals.

**Data Model Layer**: The `internal/data` package provides:
- `Models` struct that wraps all model types (`RecipeModel`, `UserModel`, `TokenModel`)
- `NewModels(db)` factory function for initialization
- Each model has its own file and encapsulates database operations
- Custom `Duration` type with JSON marshaling/unmarshaling

**Envelope Response Pattern**: All JSON responses use an envelope wrapper (see `cmd/api/helpers.go:29`):
```go
envelope{"recipe": recipeData}
envelope{"error": errorMessage}
```

**Database Transaction Pattern**: Complex operations like `Insert()` in `recipes.go:69` use explicit transactions to ensure atomicity when inserting across multiple related tables (recipes, ingredients, equipment, instructions, images).

### Database Schema

The schema uses a normalized relational design with 4 migrations:

**Recipe Tables (Migration 000001, 000002):**
- **recipes**: Core recipe metadata (name, description, notes, prep_time interval, active_time interval, servings, version)
- **ingredients**: Normalized ingredient names (deduplicated with UNIQUE constraint)
- **equipment**: Normalized equipment names (deduplicated with UNIQUE constraint)
- **recipe_ingredients**: Junction table with quantity, unit, optional flag
- **recipe_equipment**: Junction table for required equipment
- **recipe_instructions**: Step-by-step instructions with step_number, text, notes
- **recipe_images**: Image URLs with ENUM type (thumbnail, main, step)
- **recipe_instruction_images**: Links images to specific instruction steps
- **tags** / **recipe_tags**: Tagging system (schema exists, not yet implemented in code)

**User & Authentication Tables (Migration 000003, 000004):**
- **users**: User accounts with citext email (case-insensitive), password_hash (bytea), activated (boolean), version (optimistic locking)
- **tokens**: Authentication and activation tokens with hash (SHA-256), user_id (FK with CASCADE), expiry, scope

**Key Schema Features:**
- Optimistic locking via `version` fields (recipes, users)
- CASCADE deletes throughout the schema
- Check constraints: prep_time >= 0, active_time >= 0, servings > 0
- UNIQUE constraints for data deduplication
- Custom ENUM type for image categorization

The `instructions` field in the recipes table stores JSONB for flexibility, but the normalized `recipe_instructions` table is primarily used for structured step-by-step data.

### API Endpoints

All endpoints are **fully implemented and operational**:

**Health Check:**
- `GET /v1/healthcheck` - Service health status

**Recipes (Full CRUD + List):**
- `GET /v1/recipes` - List recipes with filtering, sorting, and pagination ✅
- `POST /v1/recipes` - Create new recipe (requires activated user) ✅
- `GET /v1/recipes/:id` - Get single recipe with all related data ✅
- `PATCH /v1/recipes/:id` - Update recipe with optimistic locking (requires activated user) ✅
- `DELETE /v1/recipes/:id` - Delete recipe (requires activated user) ✅

**Users:**
- `POST /v1/users` - Register new user account ✅
- `PUT /v1/users/activated` - Activate user account with token ✅

**Authentication:**
- `POST /v1/tokens/authentication` - Generate authentication token (24h expiry) ✅

**Filtering Options for GET /v1/recipes:**
- `name` - Filter by recipe name (case-insensitive partial match)
- `ingredients` - Filter by ingredients (comma-separated list)
- `equipment` - Filter by required equipment (comma-separated list)
- `prep_time` - Maximum prep time in minutes
- `active_time` - Maximum active time in minutes
- `sort` - Sort by: id, name, prep_time, active_time (prefix with `-` for descending)
- `page` - Page number (default: 1)
- `page_size` - Results per page (default: 20, max: 100)

### Validation

The `internal/validator` package provides a flexible validation system:
- `Validator` type with an error map
- `Check()` method for adding conditional validation errors
- Helper functions: `PermittedValue()`, `Matches()`, `Unique()`

Validation is applied in handlers before database operations (see `recipes.go:83`).

### Error Handling

Centralized error response handlers in `cmd/api/errors.go` (11 types):
- `serverErrorResponse()` - 500 Internal Server Error with logging
- `notFoundResponse()` - 404 Not Found
- `methodNotAllowedResponse()` - 405 Method Not Allowed
- `badRequestResponse()` - 400 Bad Request
- `failedValidationResponse()` - 422 Unprocessable Entity with validation errors
- `editConflictResponse()` - 409 Conflict (optimistic locking failure)
- `rateLimitExceededResponse()` - 429 Too Many Requests
- `invalidCredentialsResponse()` - 401 Unauthorized (bad email/password)
- `invalidAuthenticationTokenResponse()` - 401 Unauthorized (bad token)
- `authenticationRequiredResponse()` - 401 Unauthorized (missing token)
- `inactiveAccountResponse()` - 403 Forbidden (account not activated)

Panic recovery middleware wraps all routes with proper `Connection: close` header handling (see `middleware.go:17`).

### JSON Handling

`readJSON()` helper in `helpers.go:31` provides robust JSON parsing with:
- 1MB request body limit
- Disallowed unknown fields
- Detailed error messages for malformed JSON
- Protection against multiple JSON values

## Current Status - Production Ready Core Features

### ✅ Fully Implemented Features

**Recipe Management (100% Complete):**
- ✅ Full CRUD operations with complex JOIN queries
- ✅ Advanced filtering by name, ingredients, equipment, and time constraints
- ✅ Sorting and pagination with metadata
- ✅ Transaction-based inserts across 8+ related tables
- ✅ Optimistic locking for updates (version field)
- ✅ CASCADE deletes via database schema

**User Management & Authentication (100% Complete):**
- ✅ User registration with email/password
- ✅ Password hashing with bcrypt (cost 12)
- ✅ Account activation via secure tokens (3-day expiry)
- ✅ Welcome email sent in background goroutine
- ✅ Authentication token generation (24-hour expiry)
- ✅ Token-based request authentication
- ✅ Cryptographically secure token generation (crypto/rand + SHA-256)

**Middleware & Infrastructure (100% Complete):**
- ✅ Panic recovery with connection close
- ✅ IP-based rate limiting (configurable RPS/burst)
- ✅ Bearer token authentication middleware
- ✅ User activation checks
- ✅ Request context management (user injection)
- ✅ Graceful shutdown with signal handling (SIGINT/SIGTERM)
- ✅ Background task tracking with WaitGroup
- ✅ 30-second shutdown timeout

**Email System (100% Complete):**
- ✅ SMTP mailer with embedded templates
- ✅ HTML and plain text support
- ✅ Background email sending (non-blocking)
- ✅ Panic recovery in background tasks
- ✅ Template system (subject/plainBody/htmlBody)

**Data Models (100% Complete):**
- ✅ Recipe: Insert, Get, Update, Delete, GetAll
- ✅ User: Insert, GetByEmail, Update, GetForToken
- ✅ Token: New, Insert, DeleteAllForUser
- ✅ Custom Duration type with JSON marshaling/unmarshaling
- ✅ Filters and Metadata for pagination

### 📝 Known TODOs

1. **Ingredient/Equipment Normalization** (recipes.go:63):
   - Convert ingredient and equipment names to lowercase for better deduplication
   - Currently case-sensitive matching may create duplicate entries

2. **Tag System**:
   - Database tables exist (tags, recipe_tags)
   - No model methods or handlers implemented yet
   - Would enable categorization: cuisine types, difficulty levels, dietary restrictions

3. **Public/Private Recipes**:
   - `public` boolean field exists in recipes table
   - Not currently enforced in GetAll queries
   - Could restrict recipe visibility based on creator/household

4. **Creator Tracking**:
   - `creator` field exists in Recipe struct
   - Not stored in database or used in queries
   - Would enable "my recipes" filtering

### 🎯 Production Readiness

The application is **production-ready** for its core use case: recipe management with user authentication. All "Let's Go Further" chapters 1-15 patterns are implemented:
- ✅ Ch 1-4: Project setup, JSON handling
- ✅ Ch 5-6: Database setup and migrations
- ✅ Ch 7-8: CRUD with optimistic locking
- ✅ Ch 9: Filtering, sorting, pagination
- ✅ Ch 10: Rate limiting
- ✅ Ch 11: Graceful shutdown
- ✅ Ch 12: User registration
- ✅ Ch 13: Background emails
- ✅ Ch 14: User activation
- ✅ Ch 15: Authentication

**Not Yet Implemented:**
- ⏸️ Ch 16: Permissions (deferred per user request)
- ❌ Ch 17: CORS (needed for frontend integration)
- ❌ Ch 18: Metrics (observability)

### 🚀 Recommended Next Steps

See "Future Architecture Evolution" section below for roadmap toward permissions, CORS, households, and the "family restaurant ordering" feature.

## Common Development Tasks

### When Adding New Endpoints:
1. Define the handler method on `*application` in the appropriate file (e.g., `recipes.go`)
2. Add route in `routes.go:22-24`
3. Follow this pattern in handlers:
   ```go
   func (app *application) exampleHandler(w http.ResponseWriter, r *http.Request) {
       // 1. Parse and extract request data
       // 2. Validate input using validator package
       // 3. Call model method
       // 4. Handle errors appropriately
       // 5. Return JSON response using writeJSON()
   }
   ```
4. Use existing error response helpers for consistency
5. Always use `context.WithTimeout()` for database operations (3 seconds recommended)

### When Adding New Models:
1. Create new file in `internal/data/` (e.g., `users.go`)
2. Add model to `Models` struct in `models.go:16`
3. Initialize in `NewModels()` factory function
4. Follow the pattern of `RecipeModel` with a DB field and CRUD methods
5. Each CRUD method should:
   - Accept context with timeout
   - Use placeholder parameters ($1, $2, etc.) to prevent SQL injection
   - Return appropriate errors (ErrRecordNotFound, ErrEditConflict)
   - Use transactions for multi-table operations

### When Adding Database Migrations:
1. Create numbered `.up.sql` and `.down.sql` files in `migrations/`
2. Follow existing naming convention: `000001_description.up.sql`
3. Include both DDL and appropriate `IF NOT EXISTS` clauses
4. Always create corresponding `.down.sql` to rollback changes
5. Drop tables in reverse order of creation in down migration

### When Implementing CRUD Operations (Following "Let's Go Further" Ch. 7):

**Insert Pattern:**
```go
func (m Model) Insert(item *Item) error {
    query := `INSERT INTO items (field1, field2) VALUES ($1, $2) RETURNING id, created_at, version`

    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    return m.DB.QueryRowContext(ctx, query, item.Field1, item.Field2).Scan(&item.ID, &item.CreatedAt, &item.Version)
}
```

**Get Pattern:**
```go
func (m Model) Get(id int64) (*Item, error) {
    if id < 1 {
        return nil, ErrRecordNotFound
    }

    query := `SELECT id, created_at, field1, field2, version FROM items WHERE id = $1`

    var item Item

    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    err := m.DB.QueryRowContext(ctx, query, id).Scan(&item.ID, &item.CreatedAt, &item.Field1, &item.Field2, &item.Version)

    if err != nil {
        switch {
        case errors.Is(err, sql.ErrNoRows):
            return nil, ErrRecordNotFound
        default:
            return nil, err
        }
    }

    return &item, nil
}
```

**Update Pattern (with Optimistic Locking):**
```go
func (m Model) Update(item *Item) error {
    query := `
        UPDATE items
        SET field1 = $1, field2 = $2, version = version + 1
        WHERE id = $3 AND version = $4
        RETURNING version`

    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    err := m.DB.QueryRowContext(ctx, query, item.Field1, item.Field2, item.ID, item.Version).Scan(&item.Version)
    if err != nil {
        switch {
        case errors.Is(err, sql.ErrNoRows):
            return ErrEditConflict
        default:
            return err
        }
    }

    return nil
}
```

**Delete Pattern:**
```go
func (m Model) Delete(id int64) error {
    if id < 1 {
        return ErrRecordNotFound
    }

    query := `DELETE FROM items WHERE id = $1`

    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    result, err := m.DB.ExecContext(ctx, query, id)
    if err != nil {
        return err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }

    if rowsAffected == 0 {
        return ErrRecordNotFound
    }

    return nil
}
```

## Key Patterns from "Let's Go Further"

### Error Handling Best Practices
- **Always check errors explicitly** - Never ignore errors
- **Log server errors, hide from clients** - Use `serverErrorResponse()` for unexpected errors
- **Provide actionable client errors** - Use appropriate status codes (400, 404, 422, etc.)
- **Define custom error types** - e.g., `ErrRecordNotFound`, `ErrEditConflict`
- **Use error helpers consistently** - All error responses should go through helper functions

### Database Best Practices
- **Always use context with timeout** - 3 seconds is recommended for most operations
- **Use placeholder parameters** - Always use $1, $2, etc., never string concatenation
- **Handle sql.ErrNoRows** - Convert to application-specific ErrRecordNotFound
- **Implement optimistic locking** - Use version field and check in WHERE clause
- **Use transactions for multi-table operations** - Ensure atomicity (already done in Insert)
- **Close rows with defer** - Always `defer rows.Close()` when using Query()

### JSON Handling Best Practices
- **Use envelope pattern** - Wrap all responses: `envelope{"data": value}`
- **Limit request body size** - Use `http.MaxBytesReader` (1MB recommended)
- **Disallow unknown fields** - Call `dec.DisallowUnknownFields()`
- **Check for multiple JSON values** - Ensure only one object per request
- **Provide detailed error messages** - Handle each JSON error type specifically
- **Use json.MarshalIndent** - Make responses readable for debugging

### Validation Best Practices
- **Collect all errors** - Don't stop at first validation error
- **Validate at data layer** - Keep validation logic with models
- **Return 422 for validation errors** - Use Unprocessable Entity status
- **Provide field-specific errors** - Map field names to error messages
- **Validate before database operations** - Check in handler before calling model

### Security Best Practices
- **Rate limit all endpoints** - Implement IP-based or global rate limiting
- **Use bcrypt for passwords** - Cost factor 12 recommended
- **Store token hashes only** - Never store plaintext tokens
- **Implement CORS carefully** - Use safelist, never wildcard with credentials
- **Set appropriate timeouts** - ReadTimeout, WriteTimeout, IdleTimeout on server
- **Use HTTPS in production** - Never send credentials over plain HTTP

## Future Architecture Evolution

**Note**: Phases 1 and parts of Phase 2 are already complete. This roadmap outlines remaining work for the "family restaurant ordering" system.

### ✅ Phase 1: User Management & Authentication (COMPLETE)
All work from "Let's Go Further" chapters 12-15 is done:
- ✅ User Model with bcrypt password hashing
- ✅ Token System with cryptographic token generation
- ✅ Authentication middleware with Bearer token support
- ✅ User registration and activation endpoints
- ✅ Background email sending for activation

### ⏸️ Phase 2: Authorization & Permissions (PARTIALLY COMPLETE)
**Already Implemented:**
- ✅ `requireAuthenticatedUser()` middleware
- ✅ `requireActivatedUser()` middleware
- ✅ Middleware chaining on recipe endpoints

**Deferred (per user request after reading Ch. 16):**
- ⏸️ Fine-grained permissions system (permissions, user_permissions tables)
- ⏸️ Permission checking middleware (`requirePermission()`)
- ⏸️ Role-based access control beyond activated/not-activated

**When to Implement:** Before multi-tenant (household) features to properly scope access.

### Phase 3: Multi-Tenant Architecture
1. **Households**
   - Add households table (represents a "family" unit)
   - Add household_members junction table
   - Link recipes to households (add household_id to recipes)

2. **Roles**
   - Implement chef role (creates recipes, fulfills orders)
   - Implement family_member role (browses recipes, places orders)
   - Add role checking in authorization middleware

### Phase 4: Order Management System
1. **Order Schema**
   - Create orders table (status, requested_for timestamp, chef_id, requester_id)
   - Create order_items table (links to recipes)
   - Add chef_availability table for scheduling

2. **Order Endpoints**
   - `GET /v1/recipes?household_id=:id` - Browse available recipes
   - `POST /v1/orders` - Place order (family member)
   - `GET /v1/orders` - List orders (role-filtered view)
   - `PATCH /v1/orders/:id` - Update order status (chef only)

3. **Notifications**
   - Implement background email worker (Ch. 13)
   - Send order confirmation emails
   - Send status update emails
   - Consider real-time updates (SSE or WebSockets)

### Phase 5: Advanced Features & Production Readiness

**Already Implemented:**
- ✅ **Filtering & Pagination** (Ch. 9) - Recipe search with metadata, multiple filter types
- ✅ **Rate Limiting** (Ch. 10) - IP-based rate limiting, configurable via flags
- ✅ **Graceful Shutdown** (Ch. 11) - Signal handling with WaitGroup for background tasks

**Still Needed:**
1. **CORS** (Ch. 17) - HIGH PRIORITY for frontend integration
   - Add CORS middleware
   - Configure trusted origins
   - Handle preflight requests

2. **Metrics** (Ch. 18) - Observability
   - Expose metrics endpoint with expvar
   - Track request counts, response times, errors
   - Monitor background task queue depth

3. **Full-Text Search** - Enhancement to existing filtering
   - PostgreSQL full-text search on recipe names/descriptions
   - Search ranking and highlighting

4. **Testing** - Quality assurance
   - Integration tests with test database
   - Handler unit tests
   - Table-driven test suites
   - Race detector testing (`go test -race`)

5. **Deployment** (Ch. 19-20)
   - Build scripts and versioning
   - Production configuration
   - Deployment to cloud provider
   - HTTPS/TLS setup

## Book Reference

For detailed implementation guidance, refer to specific chapters:

### "Let's Go Further" (Primary Reference for API Development)
Available at: `~/projects/docs/lets-go-further/html/`

- Ch. 2: Getting Started (Project Structure)
- Ch. 3-4: Sending/Parsing JSON
- Ch. 5: Database Setup and Configuration
- Ch. 6: SQL Migrations
- Ch. 7: CRUD Operations
- Ch. 8: Advanced CRUD (Partial Updates, Optimistic Locking)
- Ch. 9: Filtering, Sorting, and Pagination
- Ch. 10: Rate Limiting
- Ch. 11: Graceful Shutdown
- Ch. 12: User Model Setup and Registration
- Ch. 13: Sending Emails (Background Tasks)
- Ch. 14: User Activation
- Ch. 15: Authentication
- Ch. 16: Authorization
- Ch. 17: Cross-Origin Requests
- Ch. 18: Metrics
- Ch. 19: Building, Versioning, and Quality Control
- Ch. 20: Deployment and Hosting

### "Let's Go" (Complementary Reference for Fundamentals)
Available at: `~/projects/docs/lets-go/html/`

- Ch. 2: Foundations (Handlers, Routing, HTTP Basics)
- Ch. 3: Configuration and Error Handling
- Ch. 4: Database-Driven Responses
- Ch. 5: Dynamic HTML Templates (if adding web UI)
- Ch. 6: Middleware
- Ch. 8: Processing Forms (validation patterns)
- Ch. 9: Stateful HTTP (sessions)
- Ch. 10: Security Improvements
- Ch. 11: User Authentication (cookie-based)
- Ch. 13: Testing (comprehensive unit/integration testing)
- Ch. 14: Conclusion and Further Reading

**When to Reference Each:**
- **Let's Go Further**: For API-specific patterns (JSON, tokens, CORS, rate limiting)
- **Let's Go**: For deeper understanding of fundamentals (testing, concurrency, HTTP details)

## Additional Patterns from "Let's Go"

While "Let's Go Further" is the primary guide for this API project, "Let's Go" (the first book) provides valuable complementary patterns, especially around testing, error handling details, and foundational concepts.

### Testing Best Practices (Comprehensive from Let's Go)

**Table-Driven Tests:**
```go
func TestHumanDate(t *testing.T) {
    tests := []struct {
        name string
        tm   time.Time
        want string
    }{
        {
            name: "UTC",
            tm:   time.Date(2020, 12, 17, 10, 0, 0, 0, time.UTC),
            want: "17 Dec 2020 at 10:00",
        },
        {
            name: "Empty",
            tm:   time.Time{},
            want: "",
        },
        {
            name: "CET",
            tm:   time.Date(2020, 12, 17, 10, 0, 0, 0, time.FixedZone("CET", 1*60*60)),
            want: "17 Dec 2020 at 09:00",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            hd := humanDate(tt.tm)
            if hd != tt.want {
                t.Errorf("want %q; got %q", tt.want, hd)
            }
        })
    }
}
```

**Testing HTTP Handlers:**
```go
func TestPing(t *testing.T) {
    // Create response recorder
    rr := httptest.NewRecorder()
    
    // Create request
    r, err := http.NewRequest(http.MethodGet, "/", nil)
    if err != nil {
        t.Fatal(err)
    }
    
    // Call handler
    ping(rr, r)
    
    // Check response
    rs := rr.Result()
    
    if rs.StatusCode != http.StatusOK {
        t.Errorf("want %d; got %d", http.StatusOK, rs.StatusCode)
    }
    
    defer rs.Body.Close()
    body, err := io.ReadAll(rs.Body)
    if err != nil {
        t.Fatal(err)
    }
    
    if string(body) != "OK" {
        t.Errorf("want body to equal %q", "OK")
    }
}
```

**Testing Middleware:**
```go
func TestSecureHeaders(t *testing.T) {
    rr := httptest.NewRecorder()
    
    r, err := http.NewRequest(http.MethodGet, "/", nil)
    if err != nil {
        t.Fatal(err)
    }
    
    // Create mock next handler
    next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })
    
    // Test middleware wrapping next handler
    secureHeaders(next).ServeHTTP(rr, r)
    
    rs := rr.Result()
    
    // Check headers were set
    frameOptions := rs.Header.Get("X-Frame-Options")
    if frameOptions != "deny" {
        t.Errorf("want %q; got %q", "deny", frameOptions)
    }
    
    xssProtection := rs.Header.Get("X-XSS-Protection")
    if xssProtection != "1; mode=block" {
        t.Errorf("want %q; got %q", "1; mode=block", xssProtection)
    }
}
```

**Test Commands:**
```bash
go test ./...                    # Run all tests
go test -v ./cmd/api            # Verbose output
go test -run="^TestPing$"       # Run specific test
go test -parallel 4             # Parallel execution
go test -race                   # Enable race detector (critical for concurrent code)
go test -failfast               # Stop on first failure
go test -cover                  # Show coverage
go test -coverprofile=/tmp/profile.out  # Save coverage profile
go tool cover -html=/tmp/profile.out    # View coverage in browser
```

### Enhanced Error Handling with Stack Traces

While "Let's Go Further" covers error handling, "Let's Go" adds the detail of using `debug.Stack()` for server errors:

```go
func (app *application) serverError(w http.ResponseWriter, err error) {
    trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
    // Output at depth 2 to get correct file:line of caller
    app.errorLog.Output(2, trace)
    
    http.Error(w, http.StatusText(http.StatusInternalServerError), 500)
}
```

**Key Insight**: Using `log.Output(2, message)` reports the file and line number 2 frames back in the call stack, giving you the actual location of the error rather than the helper function location.

### Panic Recovery Details

"Let's Go" emphasizes an important detail about panic recovery:

```go
func (app *application) recoverPanic(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                // Set Connection: close header to trigger HTTP server
                // to automatically close the connection
                w.Header().Set("Connection", "close")
                app.serverError(w, fmt.Errorf("%s", err))
            }
        }()
        
        next.ServeHTTP(w, r)
    })
}
```

**Critical**: 
- Panic recovery only works in the same goroutine
- If you spawn background goroutines in handlers, they need their own recovery
- Setting `Connection: close` ensures the connection is closed after panic

### HTTP Handler Fundamentals

"Let's Go" provides deeper explanation of the handler interface:

**The Handler Interface:**
```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

**Three Ways to Create Handlers:**

1. **Handler as struct with method:**
```go
type home struct {
    logger *log.Logger
}

func (h *home) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Handler logic
}
```

2. **Handler as function with http.HandlerFunc adapter:**
```go
func home(w http.ResponseWriter, r *http.Request) {
    // Handler logic
}

mux.Handle("/", http.HandlerFunc(home))
// Or use HandleFunc shortcut:
mux.HandleFunc("/", home)
```

3. **Handler with closure for dependency injection:**
```go
func home(logger *log.Logger) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        logger.Print("Home page accessed")
        // Handler logic
    }
}

mux.HandleFunc("/", home(myLogger))
```

**Important Concurrency Note**: Each HTTP request is handled in its own goroutine. Be mindful of:
- Race conditions when accessing shared resources
- Panic recovery doesn't propagate across goroutines
- Use `go test -race` to detect data races

### HTTP Response Header Nuances

**Header Canonicalization:**
```go
// These are equivalent (canonicalized):
w.Header().Set("content-type", "application/json")
w.Header().Set("Content-Type", "application/json")

// To avoid canonicalization (rare), use map directly:
w.Header()["X-XSS-Protection"] = []string{"1; mode=block"}
```

**Header Methods:**
```go
w.Header().Set("Cache-Control", "public")        // Overwrites existing
w.Header().Add("Cache-Control", "max-age=3600")  // Appends
w.Header().Del("Cache-Control")                  // Deletes
w.Header().Get("Cache-Control")                  // Gets first value
```

**Order Matters:**
```go
// CORRECT
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(201)
w.Write([]byte(`{"id": 123}`))

// WRONG - WriteHeader must come before Write
w.WriteHeader(201)
w.Header().Set("Content-Type", "application/json")  // Too late!
w.Write([]byte(`{"id": 123}`))
```

### Request Routing Details (stdlib servemux)

While this project uses httprouter, understanding stdlib servemux helps understand routing concepts:

**Path Matching Rules:**
- **Fixed paths**: `/snippet/create` matches exactly
- **Subtree paths**: `/static/` matches `/static/*` (trailing slash matters!)
- **Longer patterns win**: `/static/js/` beats `/static/`
- **Root pattern `/`**: Catch-all, matches everything

**Restricting Root Pattern:**
```go
func home(w http.ResponseWriter, r *http.Request) {
    // Prevent "/" from matching all paths
    if r.URL.Path != "/" {
        http.NotFound(w, r)
        return
    }
    // Handler logic
}
```

**Method-Based Routing (Manual):**
```go
func createSnippet(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.Header().Set("Allow", http.MethodPost)
        w.WriteHeader(405)
        w.Write([]byte("Method Not Allowed"))
        return
    }
    // POST handler logic
}
```

### Configuration Best Practices

**Command-Line Flags for Configuration:**
```go
// In main.go
addr := flag.String("addr", ":4000", "HTTP network address")
dsn := flag.String("dsn", os.Getenv("DB_DSN"), "Database DSN")
flag.Parse()

// Access with *addr, *dsn (dereference pointers)
```

**Environment Variables as Defaults:**
```go
flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("EATINN_DB_DSN"), "PostgreSQL DSN")
```

This pattern (already used in the project) allows environment variables as defaults with command-line override capability.

### Leveled Logging

"Let's Go" uses separate loggers for info and error messages:

```go
infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
```

**Note**: "Let's Go Further" uses structured logging with `slog` package (Go 1.21+), which is preferred for this project. However, the concept of separating info/error streams is the same.

### Form Validation Pattern (Adaptable to JSON)

While "Let's Go" focuses on HTML forms, the validation pattern can inspire JSON validation:

```go
type Form struct {
    Values url.Values
    Errors map[string]string
}

func (f *Form) Required(fields ...string) {
    for _, field := range fields {
        value := f.Values.Get(field)
        if strings.TrimSpace(value) == "" {
            f.Errors[field] = "This field cannot be blank"
        }
    }
}

func (f *Form) MaxLength(field string, d int) {
    value := f.Values.Get(field)
    if value == "" {
        return
    }
    if utf8.RuneCountInString(value) > d {
        f.Errors[field] = fmt.Sprintf("This field is too long (maximum is %d characters)", d)
    }
}

func (f *Form) PermittedValues(field string, opts ...string) {
    value := f.Values.Get(field)
    if value == "" {
        return
    }
    for _, opt := range opts {
        if value == opt {
            return
        }
    }
    f.Errors[field] = "This field is invalid"
}

func (f *Form) Valid() bool {
    return len(f.Errors) == 0
}
```

**Adaptation for APIs**: The `validator` package in this project already follows this pattern with `Check()`, `Valid()`, and error map. The book shows additional validators that could be added (MinLength, Email format, URL format, etc.).

## Key Differences: Let's Go vs Let's Go Further

Understanding what's different helps avoid confusion:

| Aspect | Let's Go | Let's Go Further |
|--------|----------|------------------|
| **Output** | HTML (server-side rendering) | JSON (REST API) |
| **Sessions** | Cookie-based encrypted sessions | Stateless token authentication |
| **CSRF** | Required (form-based apps) | Not applicable (stateless API) |
| **Forms** | HTML form processing & validation | JSON request validation |
| **Templates** | Extensive html/template usage | No templates |
| **Middleware** | Session, CSRF, security headers | Rate limiting, CORS, auth tokens |
| **Routing** | stdlib servemux, then pat router | httprouter/chi |
| **Logging** | `log` package with custom loggers | `slog` structured logging |
| **Testing Focus** | Handler/middleware unit tests | Integration testing with database |

**For This Project**: Follow "Let's Go Further" patterns since it's a REST API. Use "Let's Go" for deeper understanding of fundamentals, especially testing and concurrency concepts.
