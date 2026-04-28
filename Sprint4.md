# Sprint 4

## Work Completed

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

Test coverage includes:

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

Run all backend tests:

```bash
cd backend
go test ./...
```

## Notes For Video Demo

Timeline endpoint demo points:

- Show a task with `start_date` and `deadline`.
- Show a task with only `deadline`.
- Show that an undated task is not returned by the timeline endpoint.
- Explain that the endpoint joins stages so each task includes `stage_name`.
- Explain that project access is enforced before returning timeline data.
