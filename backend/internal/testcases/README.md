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
| `stage_reorder_test.go` | Project stage reorder service and controller behavior | 3 tests |
| `activity_endpoint_test.go` | Paginated project activity endpoint behavior | 4 tests |
| `task_search_test.go` | Project-wide task search service and endpoint behavior | 5 tests |

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

### Sprint 4 Stage Reorder Tests
```bash
# Run only the stage reorder tests
go test ./internal/testcases -run Stage.*Reorder
```

### Sprint 4 Activity Feed Tests
```bash
# Run only the activity endpoint tests
go test ./internal/testcases -run ActivityController_GetActivity
```

### Sprint 4 Task Search Tests
```bash
# Run only the project task search tests
go test ./internal/testcases -run Task.*Search
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
- Stage reorder endpoint response format
- Activity feed response format and pagination
- Task search endpoint response format and validation

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

### Stage Reorder Tests
- Updates stage positions from an ordered ID list.
- Persists reordered positions in the database.
- Returns reordered stages in position order.
- Rejects empty `stage_ids`.
- Rejects duplicate stage IDs.
- Rejects stage IDs from another project.
- Rejects missing projects.
- Rejects users without project access.

### Activity Feed Tests
- Returns paginated project activity logs.
- Returns top-level `logs`, `total`, and `page` fields.
- Orders activity by `created_at` descending.
- Returns the second page of activity.
- Returns an empty activity feed for projects without logs.
- Rejects invalid project IDs.
- Rejects unauthenticated requests.
- Returns `404` for missing projects.
- Returns `403` for users without project access.

### Task Search Tests
- Searches task titles and descriptions.
- Searches across all stages in the requested project.
- Performs case-insensitive matching.
- Prevents tasks from other projects from leaking into results.
- Returns compact task result fields with `stage_name`.
- Returns an empty array when no tasks match.
- Rejects missing or blank query text.
- Rejects invalid project IDs and unauthenticated requests.
- Returns `404` for missing projects.
- Returns `403` for users without project access.

### Task Enhancement Tests
- Creates tasks with `start_date`.
- Reads `start_date` through single-task and stage-task queries.
- Updates `start_date` through controller JSON requests.
- Preserves `start_date` when older clients omit the field.
- Clears `start_date` when clients explicitly send `null`.
- Keeps deadline, priority, and assignee behavior backward compatible.

## Sprint 4 Backend Test Notes

Sprint 4 added a timeline endpoint for the frontend Timeline/Gantt view, a
stage reorder endpoint for Kanban column persistence, and a paginated activity
feed endpoint for the Dashboard. Sprint 4 also added project-wide task search
for the board search experience. These endpoints depend on project access
checks, database ordering, pagination, validation, and compact response
formatting, so the tests are split by behavior:

- `timeline_test.go` covers endpoint-specific timeline behavior.
- `task_enhancements_test.go` covers reusable task date behavior.
- `stage_reorder_test.go` covers stage order persistence and validation.
- `activity_endpoint_test.go` covers dashboard activity feed pagination and errors.
- `task_search_test.go` covers project-wide task search matching and errors.

This separation keeps the endpoint tests focused on the public API while the
task enhancement tests verify that the underlying task model continues to work
for regular board and planner flows. The stage reorder tests focus on the
backend contract needed by the frontend drag-and-drop board. The activity
endpoint tests focus on the Dashboard feed contract and its pagination metadata.
The task search tests focus on the board search contract and project access
rules.

## Task Search Acceptance Coverage

The task search tests map directly to the Sprint 4 acceptance criteria:

| Acceptance Criteria | Test Coverage |
|---------------------|---------------|
| Searches tasks by title | `TestTaskService_SearchProjectTasksMatchesTitleAndDescription` |
| Searches tasks by description | `TestTaskService_SearchProjectTasksMatchesTitleAndDescription` |
| Searches within the requested project only | `TestTaskService_SearchProjectTasksMatchesTitleAndDescription` |
| Returns task metadata and stage name | `TestTaskService_SearchProjectTasksMatchesTitleAndDescription` |
| Returns plain JSON array | `TestTaskController_SearchProjectTasksReturnsPlainArray` |
| Returns empty array for no matches | `TestTaskService_SearchProjectTasksReturnsEmptyArray` |
| Rejects blank query text | `TestTaskService_SearchProjectTasksValidatesProjectAndAccess` |
| Returns 404 if project not found | `TestTaskService_SearchProjectTasksValidatesProjectAndAccess` |
| Rejects invalid project IDs | `TestTaskController_SearchProjectTasksValidatesRequest` |
| Rejects missing auth | `TestTaskController_SearchProjectTasksValidatesRequest` |
| Rejects users without access | `TestTaskController_SearchProjectTasksValidatesRequest` |

## Activity Feed Acceptance Coverage

The activity endpoint tests map directly to the Sprint 4 acceptance criteria:

| Acceptance Criteria | Test Coverage |
|---------------------|---------------|
| Logs project actions | `TestActivityController_GetActivityReturnsPaginatedFeed` |
| Paginates with page and limit | `TestActivityController_GetActivityReturnsPaginatedFeed` |
| Returns second page correctly | `TestActivityController_GetActivityReturnsSecondPage` |
| Orders by created_at descending | `TestActivityController_GetActivityReturnsPaginatedFeed` |
| Returns 404 if project not found | `TestActivityController_GetActivityValidation` |
| Rejects missing auth | `TestActivityController_GetActivityValidation` |
| Rejects invalid project IDs | `TestActivityController_GetActivityValidation` |
| Rejects users without access | `TestActivityController_GetActivityValidation` |

## Stage Reorder Acceptance Coverage

The stage reorder tests map directly to the Sprint 4 acceptance criteria:

| Acceptance Criteria | Test Coverage |
|---------------------|---------------|
| Accepts ordered array of stage IDs | `TestStageService_ReorderStagesUpdatesPositions` |
| Updates position field for each stage | `TestStageService_ReorderStagesUpdatesPositions` |
| Uses persisted database order | `TestStageService_ReorderStagesUpdatesPositions` |
| Returns updated stages in new order | `TestStageController_ReorderStagesReturnsUpdatedOrder` |
| Rejects stage IDs from another project | `TestStageService_ReorderStagesRejectsInvalidRequests` |
| Returns error for missing project | `TestStageService_ReorderStagesRejectsInvalidRequests` |
| Rejects invalid reorder input | `TestStageService_ReorderStagesRejectsInvalidRequests` |

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
