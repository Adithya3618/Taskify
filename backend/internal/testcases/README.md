# Test Cases - Backend Unit Tests

This folder contains all unit tests for the Taskify backend, organized by functionality.

## Test Files

| File | Description | Test Count |
|------|-------------|------------|
| `jwt_service_test.go` | JWT token generation and validation | 8 tests |
| `otp_service_test.go` | OTP generation and verification | 6 tests |
| `auth_service_test.go` | Email validation, normalization, auth structures | 12 tests |
| `auth_controller_test.go` | HTTP request handling for auth endpoints | 11 tests |
| `middleware_test.go` | JWT middleware, context helpers | 10 tests |
| `models_test.go` | User, Project, Stage, Task, Message models | 13 tests |
| `controllers_test.go` | Project, Stage, Task, Message controller authorization | 14 tests |
| `services_test.go` | Service constructor tests | 4 tests |
| `user_repository_test.go` | Database operations for users | 11 tests |
| `database_test.go` | Database connection and SQL operations | 6 tests |
| `timeline_test.go` | Timeline endpoint, project access, dated task filtering | 7 tests |
| `task_enhancements_test.go` | Task deadline, priority, assignee, and start date behavior | 6 tests |

**Total: 100+ unit tests**

## Running Tests

### Using Shell Scripts (Recommended)
```bash
# Make scripts executable (if not already)
chmod +x alltestcase.sh run_all_tests.sh

# Run all tests with coverage
./alltestcase.sh
# or
./run_all_tests.sh
```

### Using Go Commands
```bash
# Run all tests in the testcases package
go test -v ./internal/testcases/...

# Run tests with coverage
go test -cover ./internal/testcases/...

# Run specific test file
go test -v ./internal/testcases/... -run "TestJWT"
```

### Sprint 4 Timeline Tests
```bash
# Run only the timeline endpoint tests
go test ./internal/testcases -run Timeline

# Run task enhancement tests related to timeline date support
go test ./internal/testcases -run 'TaskService_Create|TaskController_Enhancement|TaskController_StartDate'
```

## Test Categories

### Authentication Tests
- JWT token generation and validation
- OTP service (generation, verification, reset tokens)
- Email validation and normalization
- Auth service errors

### Controller Tests
- Request validation (invalid JSON, missing fields)
- Authorization checks (401 for missing user context)
- Error response handling
- Timeline endpoint response format and request validation

### Model Tests
- Structure validation
- JSON tag verification
- Zero value handling

### Repository Tests
- Database CRUD operations
- User lookup by email/ID
- Password updates

### Database Tests
- Connection handling
- SQL operations (CREATE, INSERT, SELECT)
- Transaction support

### Timeline Tests
- Returns tasks with `deadline`.
- Returns tasks with `start_date`.
- Excludes tasks without either timeline date.
- Includes `stage_name` for frontend row grouping.
- Allows project members to read timeline data.
- Rejects users without project access.
- Rejects invalid project IDs and unauthenticated requests.

### Task Enhancement Tests
- Creates tasks with `start_date`.
- Reads `start_date` through single-task and stage-task queries.
- Updates `start_date` through controller JSON requests.
- Preserves `start_date` when older clients omit the field.
- Clears `start_date` when clients explicitly send `null`.
- Keeps deadline, priority, and assignee behavior backward compatible.

## Sprint 4 Backend Test Notes

Sprint 4 added a timeline endpoint for the frontend Timeline/Gantt view. The
new endpoint depends on task date fields, project access checks, and compact
response formatting, so the tests are split across two files:

- `timeline_test.go` covers endpoint-specific timeline behavior.
- `task_enhancements_test.go` covers reusable task date behavior.

This separation keeps the endpoint tests focused on the public API while the
task enhancement tests verify that the underlying task model continues to work
for regular board and planner flows.

## Timeline Endpoint Acceptance Coverage

The timeline tests map directly to the Sprint 4 acceptance criteria:

| Acceptance Criteria | Test Coverage |
|---------------------|---------------|
| Returns only tasks with a deadline or start date | `TestTaskService_GetProjectTimelineFiltersDatedTasks` |
| Includes stage name for grouping rows | `TestTaskService_GetProjectTimelineFiltersDatedTasks` |
| Ordered by deadline ascending | `TestTaskService_GetProjectTimelineFiltersDatedTasks` |
| Returns empty array if no tasks have dates | `TestTaskService_GetProjectTimelineReturnsEmptyArray` |
| Requires project access | `TestTaskService_GetProjectTimelineRequiresProjectAccess` |
| Allows project members | `TestTaskService_GetProjectTimelineAllowsProjectMember` |
| Returns plain JSON array | `TestTaskController_GetProjectTimelineReturnsPlainArray` |

## Manual API Check

After starting the backend server, the endpoint can be checked with:

```bash
curl -H "Authorization: Bearer <JWT>" \
  http://localhost:8080/api/projects/<project-id>/timeline
```

Expected successful response:

```json
[
  {
    "task_id": 1,
    "title": "Build timeline endpoint",
    "stage_id": 2,
    "stage_name": "In Progress",
    "start_date": "2026-04-20T00:00:00Z",
    "deadline": "2026-04-30T00:00:00Z",
    "priority": "high",
    "assigned_to": "user-123"
  }
]
```
