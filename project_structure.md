# Nex Play Auth Service Project Structure

This document provides a comprehensive overview of the project's directory and file structure.

## Directory Tree

```text
nex_play_auth_service/
├── cmd/
│   └── server/
│       └── main.go                 # Entry point of the application
├── config/
│   └── config.go                   # Configuration management
├── internal/
│   ├── domain/                     # Domain models and business entities
│   │   └── auth.go
│   ├── handler/                    # HTTP handlers and middleware
│   │   ├── auth/
│   │   │   ├── handler.go
│   │   │   └── request.go
│   │   └── middleware/
│   │       ├── auth.go
│   │       ├── middleware.go
│   │       └── ratelimiter.go
│   ├── repository/                 # Data access layer
│   │   └── db/
│   │       ├── db.go
│   │       ├── otp_repo.go
│   │       └── user_repo.go
│   └── service/                    # Business logic layer
│       └── auth_service/
│           └── auth_service.go
├── pkg/                            # Shared utilities and libraries
│   ├── email_formate_checker/
│   │   └── email_formate.go
│   ├── generate_otp/
│   │   └── otp.go
│   ├── hash/
│   │   └── hash.go
│   ├── jwt/
│   │   └── jwt.go
│   ├── mailer/
│   │   └── mailer.go
│   ├── password_strength_checker/
│   │   └── password_strength.go
│   ├── response/
│   │   └── res.go
│   └── username_checker/
│       └── username_checker.go
├── .env                            # Environment variables
├── .gitignore                      # Git ignore rules
├── go.mod                          # Go module definition
├── go.sum                          # Go module checksums
├── nex_play.db                     # SQLite database file
└── server.exe                      # Compiled binary (Windows)
```

## Description of Key Directories

- **`cmd/`**: Contains the main applications for this project. Each subdirectory here represents a separate executable.
- **`config/`**: Handles loading and accessing application configurations, typically from environment variables or files.
- **`internal/`**: Contains private application code. Packages under `internal` cannot be imported by code outside this project.
    - **`domain/`**: Defines the core business entities and logic that are independent of any specific framework.
    - **`handler/`**: Implements the HTTP/API layer, processing requests and returning responses.
    - **`repository/`**: Handles interaction with the data storage (e.g., SQLite via GORM).
    - **`service/`**: Implements the business logic of the application, orchestrating calls between handlers and repositories.
- **`pkg/`**: Contains library code that could potentially be used by other projects. Each package is self-contained and focused on a specific utility (e.g., hashing, JWT, mailing).






//

nex_play_media_service/
├── cmd/
│   └── server/
│       └── main.go                        # Entry point
├── config/
│   └── config.go                          # Config management (env vars, etc.)
├── internal/
│   ├── domain/                            # Core business entities
│   │   ├── media.go                       # Movie, Series, Episode structs
│   │   ├── genre.go                       # Genre entity
│   │   └── recommendation.go             # Recommendation entity
│   ├── handler/
│   │   ├── media/
│   │   │   ├── handler.go                 # HTTP handlers for media
│   │   │   └── request.go                 # Request/response DTOs
│   │   ├── search/
│   │   │   ├── handler.go                 # Search handlers
│   │   │   └── request.go
│   │   ├── stream/
│   │   │   ├── handler.go                 # Stream URL handlers
│   │   │   └── request.go
│   │   └── middleware/
│   │       ├── auth.go                    # JWT validation (from auth service)
│   │       ├── middleware.go
│   │       └── ratelimiter.go
│   ├── repository/
│   │   └── db/
│   │       ├── db.go                      # DB connection (Postgres recommended)
│   │       ├── movie_repo.go              # Movie CRUD
│   │       ├── series_repo.go             # Series + Episodes CRUD
│   │       ├── genre_repo.go              # Genre queries
│   │       ├── search_repo.go             # Search queries (full-text)
│   │       └── recommendation_repo.go    # Trending, latest, recommended queries
│   └── service/
│       └── media_service/
│           ├── media_service.go           # Movie/Series business logic
│           ├── search_service.go          # Search logic
│           ├── stream_service.go          # Stream URL generation logic
│           └── recommendation_service.go # Trending, latest, recommendations
├── pkg/
│   ├── response/
│   │   └── res.go                         # Shared response formatter
│   ├── pagination/
│   │   └── pagination.go                  # Page/limit/offset helpers
│   ├── slug/
│   │   └── slug.go                        # URL slug generator for titles
│   ├── cdn/
│   │   └── cdn.go                         # CDN/signed URL generator (S3, CF)
│   └── validator/
│       └── validator.go                   # Input validation helpers
├── .env
├── .gitignore
├── go.mod
├── go.sum
└── nex_play_media.db                      # Switch to Postgres for prod