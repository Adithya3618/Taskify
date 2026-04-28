# Sprint 4

## Work Completed

### Backend: Project Stage Reorder Endpoint

Implemented a backend endpoint to persist Kanban column order after drag-and-drop
reordering in the board view:

```http
PUT /api/projects/{id}/stages/reorder
Authorization: Bearer <JWT>
Content-Type: application/json
```

Request body:

```json
{
  "stage_ids": [3, 1, 2]
}
```

Response shape:

```json
[
  {
    "id": 3,
    "user_id": "user-123",
    "project_id": 10,
    "name": "Done",
    "position": 0,
    "created_at": "2026-04-28T12:00:00Z",
    "updated_at": "2026-04-28T12:05:00Z"
  },
  {
    "id": 1,
    "user_id": "user-123",
    "project_id": 10,
    "name": "Backlog",
    "position": 1,
    "created_at": "2026-04-28T12:00:00Z",
    "updated_at": "2026-04-28T12:05:00Z"
  }
]
```

Behavior:

- Requires JWT authentication.
- Requires access to the project.
- Accepts an ordered array of stage IDs.
- Rejects an empty `stage_ids` array.
- Rejects duplicate stage IDs.
- Rejects stage IDs that do not belong to the project.
- Returns `404 Not Found` when the project does not exist.
- Updates all stage positions inside a single database transaction.
- Returns the updated stage list ordered by the new `position` values.

### Backend: Project Timeline Endpoint

Implemented a backend endpoint for the Timeline/Gantt view:

```http
GET /api/projects/{id}/timeline
Authorization: Bearer <JWT>
```

The endpoint returns project tasks that have a `start_date` or a `deadline`.
The response is pre-formatted for timeline rendering and includes the task's
stage name so the frontend can group rows without making extra stage lookups.

Response shape:

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

Behavior:

- Requires JWT authentication.
- Requires project access through project membership or ownership.
- Returns only tasks with `start_date` or `deadline`.
- Includes `stage_name` for timeline row grouping.
- Orders deadline tasks by ascending deadline.
- Places tasks without a deadline after dated deadline tasks.
- Returns an empty JSON array when the project has no dated tasks.

### Backend: Task Start Date Support

Added `start_date` support to tasks so timeline items can represent both start
and end dates.

Updated backend behavior:

- Added `start_date` to the tasks table schema.
- Added legacy migration support for existing SQLite databases.
- Added `start_date` to the backend task model.
- Added `start_date` to task create requests.
- Added `start_date` to task update requests.
- Preserves existing `start_date` when older clients update a task without
  sending the field.
- Returns `start_date` in task read responses.

Supported task create/update field:

```json
{
  "start_date": "2026-04-20T00:00:00Z"
}
```

## Backend API Documentation

### PUT `/api/projects/{id}/stages/reorder`

Purpose:

- Persists the visual order of Kanban columns after drag-and-drop.
- Prevents column order from resetting after refresh or reload.
- Keeps all position updates atomic by applying them in one transaction.

Path parameters:

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `id` | integer | yes | Project ID |

Request body:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `stage_ids` | integer array | yes | Ordered list of stage IDs for the project |

Example request:

```json
{
  "stage_ids": [3, 1, 2]
}
```

Authentication:

- Requires `Authorization: Bearer <JWT>`.
- Uses project access middleware before reordering stages.
- Allows users with project access to persist column order.

Success response:

- Status: `200 OK`
- Body: plain JSON array of updated stages ordered by position.

Error responses:

| Status | Scenario |
|--------|----------|
| `400 Bad Request` | Project ID is not numeric |
| `400 Bad Request` | Request body is invalid JSON |
| `400 Bad Request` | `stage_ids` is empty |
| `400 Bad Request` | `stage_ids` contains duplicates |
| `400 Bad Request` | A stage ID does not belong to the project |
| `401 Unauthorized` | Missing or invalid authentication |
| `403 Forbidden` | User is not a member of the project |
| `404 Not Found` | Project does not exist |
| `500 Internal Server Error` | Unexpected database or service error |

### GET `/api/projects/{id}/timeline`

Purpose:

- Provides compact task data for the frontend timeline view.
- Avoids requiring the frontend to fetch every stage and task separately.
- Keeps the timeline response focused on display fields.

Path parameters:

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `id` | integer | yes | Project ID |

Authentication:

- Requires `Authorization: Bearer <JWT>`.
- Returns `401 Unauthorized` if the user is not authenticated.
- Uses project access checks before returning project data.

Success response:

- Status: `200 OK`
- Body: plain JSON array of timeline tasks.

Error responses:

| Status | Scenario |
|--------|----------|
| `400 Bad Request` | Project ID is not numeric |
| `401 Unauthorized` | Missing or invalid authentication |
| `403 Forbidden` | User is not a member of the project |
| `500 Internal Server Error` | Unexpected database or service error |

## Backend Unit Tests

Added timeline-focused backend tests in:

```text
backend/internal/testcases/timeline_test.go
```

Added stage-reorder-focused backend tests in:

```text
backend/internal/testcases/stage_reorder_test.go
```

Test coverage includes:

Stage reorder:

- Updates positions according to the submitted order.
- Persists new positions to the database.
- Returns stages ordered by the new positions.
- Rejects an empty `stage_ids` array.
- Rejects duplicate stage IDs.
- Rejects stage IDs from a different project.
- Rejects missing projects.
- Rejects users without project access.
- Confirms controller returns updated ordered stages.

Timeline:

- Returns tasks with deadlines.
- Returns tasks with start dates but no deadline.
- Excludes tasks with no `start_date` and no `deadline`.
- Includes `stage_name` in each timeline item.
- Returns an empty array when there are no dated tasks.
- Allows access for project members.
- Rejects access for users outside the project.
- Rejects invalid project IDs.
- Rejects unauthenticated requests.
- Confirms the controller returns a plain JSON array.

Run timeline tests:

```bash
cd backend
go test ./internal/testcases -run Timeline
```

Run stage reorder tests:

```bash
cd backend
go test ./internal/testcases -run Stage.*Reorder
```

Run all backend tests:

```bash
cd backend
go test ./...
```

## Notes For Video Demo

Stage reorder endpoint demo points:

- Show Kanban columns before reordering.
- Drag columns into a new order in the frontend.
- Explain that the frontend sends the ordered stage IDs to the backend.
- Explain that the backend updates positions in a single transaction.
- Refresh the board and confirm the column order remains saved.

Timeline endpoint demo points:

- Show a task with `start_date` and `deadline`.
- Show a task with only `deadline`.
- Show that an undated task is not returned by the timeline endpoint.
- Explain that the endpoint joins stages so each task includes `stage_name`.
- Explain that project access is enforced before returning timeline data.
