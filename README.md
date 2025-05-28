# Task Management System

This Task Management System is designed as a microservice that handles task-related operations with a focus on scalability and maintainability. The system follows clean architecture principles and demonstrates best practices in microservice design.

### Key Features:
- CRUD operations for tasks
- Pagination support for listing tasks
- Filtering tasks by status
- RESTful API design
- Scalable architecture
- Clear separation of concerns


## Architecture
The project follows a clean architecture pattern with clear separation of concerns. Each layer has specific responsibilities and dependencies flow inward, with the domain layer at the center.

```
.
├── cmd/
│   └── api/                 # Application entry points
│       └── main.go         # Main application bootstrap
├── pkg/                    # Public packages
│   ├── api/               # API layer
│   │   ├── handlers/      # HTTP request handlers
│   │   ├── middleware/    # HTTP middleware components
│   │   ├── routes/        # Route definitions
│   │   └── responses/     # Response structures
│   ├── models/            # Domain models and interfaces
│   │   ├── task.go       # Task domain model
│   │   └── errors.go     # Domain error definitions
│   ├── service/           # Business logic layer
│   │   ├── task.go       # Task service implementation
│   │   └── interfaces.go  # Service interfaces
│   ├── repository/        # Data access layer
│   │   ├── postgres/      # PostgreSQL implementation
│   │   └── cache/        # Redis cache implementation
│   ├── health/           # Health check components
│   │   ├── checker.go    # Health check implementation
│   │   └── types.go      # Health check types
│   └── metrics/          # Metrics collection
├── internal/             # Private application code
│   ├── config/          # Configuration management
│   ├── database/        # Database utilities
│   └── testutil/        # Test helpers and utilities
└── scripts/             # Utility scripts
    ├── migrations/      # Database migration scripts
    └── deployment/      # Deployment scripts
```

### Layer Responsibilities

1. **API Layer** (`pkg/api/`)
   - HTTP request handling
   - Input validation
   - Authentication & authorization
   - Request/response transformation
   - API versioning

2. **Service Layer** (`pkg/service/`)
   - Business logic implementation
   - Transaction management
   - Domain rules enforcement
   - Service orchestration

3. **Repository Layer** (`pkg/repository/`)
   - Data persistence operations
   - Cache management
   - Data access abstraction
   - Query optimization

4. **Domain Layer** (`pkg/models/`)
   - Business entities
   - Value objects
   - Domain interfaces
   - Business rules

### Design Principles
- Dependency Injection
- Interface-based design
- Single Responsibility
- Dependency Inversion
- Clean Architecture patterns

### Design Decisions

1. **Single Responsibility Principle**: Each package has a specific responsibility:
   - `api`: Handles HTTP requests and responses
   - `models`: Defines domain models
   - `repository`: Manages data persistence
   - `service`: Implements business logic

2. **Scalability**:
   - Stateless design allows horizontal scaling
   - Database connection pooling
   - Containerization-ready architecture

3. **Inter-Service Communication**:
   For future expansion (e.g., adding a User Service), the system can be extended using:
   - REST APIs for synchronous communication
   - Message queues (e.g., RabbitMQ) for asynchronous communication
   - gRPC for high-performance inter-service communication



### Additional Features
- PostgreSQL database for persistent storage
- JWT-based authorization
- Role-based access control
- Redis caching for improved performance
- Basic Rate limiter
- Metrics and Monitoring(Cloudwatch)
- Health checks
- API Versioning
- Unit tests



## API Documentation

### Public Endpoints
- `GET /health` - System health status

#### Tasks

- `GET /api/v1/tasks`
  - List tasks with pagination and filtering
  - Query parameters:
    - `page`: Page number (default: 1)
    - `limit`: Items per page (default: 10)
    - `status`: Filter by status (optional)

- `POST /api/v1/tasks`
  - Create a new task
  
- `GET /api/v1/tasks/{id}`
  - Get task by ID
  
- `PUT /api/v1/tasks/{id}`
  - Update task by ID
  
- `DELETE /api/v1/tasks/{id}`
  - Delete task by ID

### Example Requests/Responses

#### Create Task
```json
POST /api/v1/tasks
Request:
{
    "title": "Complete project documentation",
    "description": "Write comprehensive documentation for the task management system",
    "status": "pending",
    "due_date": "2024-03-20T15:00:00Z"
}

Response:
{
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "title": "Complete project documentation",
    "description": "Write comprehensive documentation for the task management system",
    "status": "pending",
    "due_date": "2024-03-20T15:00:00Z",
    "created_at": "2024-03-15T10:00:00Z",
    "updated_at": "2024-03-15T10:00:00Z"
}
```


## Server Configuration
- `SERVER_PORT`: API server port (default: "8080")








## Setup and Installation

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 14 or higher

1. ### Running the Application with Docker
```bash
# Using docker-compose
docker-compose up -d --build
```

2. ### Running the Service

1. Clone the repository:
   ```bash
   git clone https://github.com/ebinthomas/task-management-system.git
   cd task-management-system
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up environment variables:
   # Edit env-example with your configuration


4. Run the service:
   ```bash
   go run cmd/api/main.go
   ```



### Testing with JWT Tokens

1. **Build the Token Generator**:
```bash
go build -o bin/token-gen cmd/tools/token_gen.go
```

2. **Generate Tokens**:
```bash
# Generate admin token
./bin/token-gen -role admin -secret your-development-secret -issuer dev-auth

# Generate user token
./bin/token-gen -role user -secret your-development-secret -issuer dev-auth

# Generate viewer token
./bin/token-gen -role viewer -secret your-development-secret -issuer dev-auth

# Using token in requests
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/tasks
```


#### Security Notes

1. The token generator is for **development and testing only**
2. In production:
   - Use a proper Authentication Service
   - Use strong secrets
   - Set appropriate token expiration
   - Implement token refresh mechanism
   - Use HTTPS for all API calls



#Additional Featues:

- Health checks
- Unit tests

1. ## PostgreSQL database for persistent storage

#### Database Configuration
- `DB_HOST`: PostgreSQL host (default: "localhost")
- `DB_PORT`: PostgreSQL port (default: "5432")
- `DB_USER`: Database username (default: "postgres")
- `DB_PASSWORD`: Database password (default: "postgres")
- `DB_NAME`: Database name (default: "taskdb")

2.##JWT Basesd Authentication
    ### Config
    - `AUTH_SECRET`: JWT signing secret (required)
    - `AUTH_ISSUER`: JWT issuer (required)


3.## Role-Based Access Control
    a. **Admin Role**:
    - Full access to all endpoints
    - Can perform all CRUD operations
    - Can access all tasks

    b. **User Role**:
    - Can create new tasks
    - Can read all tasks
    - Can update/delete own tasks only

    c. **Viewer Role**:
    - Can only read tasks
    - No write access

4. ## Redis Cache
    The service implements a Redis-based caching system for improved performance
    
    ### Caching Strategy
    - Redis-based distributed caching
    - 5-minute default TTL
    - Automatic cache invalidation on write operations
    - Cache middleware for all API routes
    - Cache bypass options available

    ### Cache Configuration
    ```bash
    # Redis Configuration
    - `REDIS_ADDR`: Redis server address
    - `REDIS_PASSWORD`: Redis server password
    ````


5. ## Rate Limiting
    The system implements a two-tier rate limiting approach:
        - Primary rate limiting at API Gateway level.
        - Safety net rate limiting at service level:
            - Allows 1000 requests per second with burst of 100
            - Prevents service abuse in case of API Gateway bypass
            - Implemented using golang.org/x/time/rate


6. ## Metrics and Monitoring (AWS CloudWatch)
    The system uses AWS CloudWatch for metrics collection and monitoring:

    ### Key Metrics

    #### API Metrics
    - `RequestDuration`: Tracks API endpoint response times
    - `APICallCount`: Tracks API call volumes with dimensions
    
    #### Cache Metrics
    - `CacheOperations`: Tracks cache performance(Hit or Miss)

    #### Service State Metrics
    - `{serviceName}Status`: Tracks service component health
        - `Values`: UP(1.0), DOWN(0.0), DEGRADED(0.5)
        - Component monitored: Database, Cache, System
    
    #### System Metrics
    - `num_goroutines`: Number of active goroutines
    - `heap_in_use`: Memory heap usage
    
    ### CloudWatch Alarms

    #### Default Alarms
    - `DatabaseDown` (threshold: 0.5)
    - `CacheDown` (threshold: 0.5)
    - `SystemDegraded` (threshold: 0.5)

    #### Custom Alert Rules (from alert.rules.yml):
    - `HighErrorRate`: 5xx errors over threshold
    - `HighLatency`: 95th percentile latency
    - `HighTaskCreationRate`: > 100 tasks/5min
    - `CacheFailureRate`: > 10% failure rate
    - `DatabaseConnectionIssues`: Any connection errors

    #### Metric & Monitoring Configs
    `ENABLE_METRICS`: Enable metrics collection (true/false)
    `ENABLE_ALARMS`: Enable alarm system (true/false)
    `ALARM_PROVIDER`: Alarm service provider (default: "cloudwatch")
    `AWS_REGION`: AWS region for CloudWatch

    #### AWS Configuration for Cloudwatch
    `AWS_REGION`=us-west-2
    `AWS_ACCESS_KEY_ID`=your-access-key
    `AWS_SECRET_ACCESS_KEY`=your-secret-key

    #### Alarm Thresholds configs
    `HIGH_ERROR_RATE_THRESHOLD`=5.0    # 5% error rate
    `HIGH_LATENCY_THRESHOLD`=1.0       # 1 second

7. ## Health Checks
    The system implements a comprehensive health check system to monitor service health and dependencies.

    ### Health Check Features
    - **Component Status Monitoring**:
      - Database connectivity
      - Redis cache availability
      - System metrics (memory, goroutines)
    
    ### Health Check Endpoint
    ```bash
    GET /health
    
    Response:
    {
        "status": "UP",
        "timestamp": "2024-03-15T10:00:00Z",
        "version": "1.0.0",
        "services": {
            "database": {
                "status": "UP",
                "message": "Database connection successful"
            },
            "cache": {
                "status": "UP",
                "message": "Redis connection successful"
            }
        },
        "system": {
            "go_version": "go1.21",
            "num_goroutines": 10,
            "num_cpu": 8,
            "heap_in_use": 1234567
        }
    }
    ```


8. ## API Versioning
    The architecture allows for multiple API versions, though at present, only version 1.0 is implemented. We can extend support to other versions in the future if needed.

9. ## Unit Tests
    The project includes comprehensive unit tests to ensure reliability and maintainability.

    ### Test Coverage
    - API handlers and middleware
    - Service layer business logic
    - Repository layer data access
    - Cache operations
    - Health checks
    - Authentication and authorization
    - Monitoring and metrics

    ### Key Test Features
    - Mock implementations for external dependencies
    - Test fixtures and helpers
    - Table-driven tests
    - Integration tests for critical paths
    - Concurrent test execution
    - Redis test instance using miniredis
    - Mock implementations for AWS services

    ### Running Tests
    ```bash
    # Run all tests
    go test ./...

    # Run tests with coverage
    go test -cover ./...

    # Run tests for a specific package
    go test ./pkg/health/...
    ```

    ### Test Organization
    - Tests are co-located with the code they test
    - Each package contains its own test files
    - Mock implementations are defined in `_test.go` files
    - Test utilities and helpers are in `internal/testutil`

    ### Testing Best Practices
    1. **Isolation**: Each test is independent and can run in isolation
    2. **Deterministic**: Tests produce the same results on each run
    3. **Fast**: Tests are optimized for quick execution
    4. **Readable**: Test cases clearly describe their intent
    5. **Maintainable**: Tests follow DRY principles and use helper functions
    6. **Complete**: Edge cases and error conditions are covered





## Running the Application

1. Start the services:
```bash
docker-compose up -d
```





