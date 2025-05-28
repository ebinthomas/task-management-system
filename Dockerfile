FROM golang:1.22-alpine

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev postgresql-client

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/main cmd/api/main.go

# Use a smaller image for the final container
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata postgresql-client

# Copy the binary from builder
COPY --from=0 /app/main .

# Copy schema and init files
COPY internal/database/migrations/001_create_tasks_table.sql ./internal/database/migrations/
COPY scripts/init.sh .
RUN chmod +x init.sh

# Copy environment file
COPY --from=0 /app/env.example .env

EXPOSE 8080

# Wait for database and run init script
CMD sh -c "until pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER; do echo 'Waiting for database...'; sleep 2; done; \
    ./init.sh" 
