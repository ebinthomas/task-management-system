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
The project follows a clean architecture pattern with the following components:

```
.
├── cmd/
│   └── api/              # Application entry point
├── pkg/
│   ├── api/             # API handlers and middleware
│   ├── models/          # Domain models
│   ├── repository/      # Data access layer
│   └── service/         # Business logic layer
└── internal/
    └── database/        # Database configuration and migrations
```

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



### Core Features
- RESTful API for task management
- JWT-based authentication and authorization
- Role-based access control
- Resource ownership validation
- PostgreSQL database for persistent storage
- Redis caching for improved performance
- Request rate limiting
- Logging middleware
- Environment-based configuration

### Monitoring & Metrics
- AWS CloudWatch integration
- Service state monitoring
- Health check endpoints
- Configurable alarms
- Metric collection and visualization
- Default alarms for critical services


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

