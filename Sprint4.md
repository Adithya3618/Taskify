# Sprint 4 - Taskify Project

## Sprint 4 Summary

Sprint 4 focused on finishing the product experience, adding regression coverage for the newest frontend features, and updating documentation for final submission. The frontend work in this branch covers profile access, board discovery refinements, task search on the board, and documentation for running, testing, and presenting the completed application.

## Frontend Work Completed - Sreeja

### Issue #106 - Create `/profile` Route and `ProfileComponent`

- Added a dedicated profile page reachable from account settings.
- Displayed the current user's name, email, and account information.
- Added editable profile form behavior with validation before saving changes.
- Integrated profile updates with the existing authentication service.
- Added unit tests for rendering profile details, saving valid changes, rejecting invalid email input, and disabling save while unchanged.

### Issue #104 - Add Board Search and Filter on the Home/Boards Page

- Added board search on the boards page for project names and descriptions.
- Added board membership filtering so users can distinguish solo boards from shared boards.
- Added active-refinement state and clear behavior for board search/filter controls.
- Added unit tests for name/description search, solo/shared membership filtering, active refinement reporting, and clearing filters.

### Issue #105 - Add Unit Tests and Cypress E2E Tests for Task Search

- Added board task-search unit tests covering:
  - title search
  - description search
  - case-insensitive and trimmed input
  - total match count across stages
  - visible/empty task state
  - clearing search through `clearFilters()`
- Added Cypress tests for board task search covering:
  - filtering task cards by title
  - filtering task cards by description across columns
  - clearing search to restore all visible tasks
  - showing an empty state when no task matches
- Updated board test mocks so task search tests run with the current board dependencies for project members, comments, subtasks, and activity.

### Documentation

- Reworked the front-page `README.md` with:
  - project overview
  - feature list
  - local run requirements
  - backend and frontend startup commands
  - app usage walkthrough
  - frontend, Cypress, and backend test commands
  - backend API overview
- Added this `Sprint4.md` report for Sprint 4 submission requirements.

## Backend Work Completed

### Backend: Project Task Search Endpoint

Implemented a backend endpoint so the board search UI can search tasks across
all stages in a project:

```http
GET /api/projects/{id}/tasks/search?q=backend
Authorization: Bearer <JWT>
```

Response shape:

```json
[
  {
    "task_id": 1,
    "title": "Build search endpoint",
    "description": "Search task titles and descriptions",
    "stage_id": 2,
    "stage_name": "In Progress",
    "deadline": "2026-05-01T10:00:00Z",
    "priority": "high",
    "assigned_to": "user-123"
  }
]
```

Behavior:

- Requires JWT authentication.
- Requires project access through project membership.
- Searches task titles and descriptions.
- Matches search text case-insensitively.
- Trims surrounding whitespace from the search query.
- Searches across all stages in the project.
- Treats `%` and `_` in search text as literal characters instead of SQL wildcards.
- Includes `stage_name` so results can show where each task belongs.
- Orders results by stage position, task position, then task ID.
- Returns an empty JSON array when no tasks match.
- Returns `400 Bad Request` when the `q` query parameter is missing or blank.
- Returns `400 Bad Request` when the `q` query parameter exceeds maximum length (100 characters).
- Returns `404 Not Found` when the project does not exist.
- Returns `403 Forbidden` when the user does not have project access.

### Backend: Paginated Project Activity Feed

Implemented and hardened the backend activity feed endpoint used by the
Dashboard recent activity section:

```http
GET /api/projects/{id}/activity?page=1&limit=20
Authorization: Bearer <JWT>
```

Response shape:

```json
{
  "logs": [
    {
      "id": 1,
      "user_name": "Owner User",
      "action": "task_created",
      "entity_type": "task",
      "entity_title": "Created task",
      "created_at": "2026-04-28T12:00:00Z"
    }
  ],
  "total": 14,
  "page": 1
}
```

Behavior:

- Requires JWT authentication.
- Requires project access through project membership.
- Supports `page` and `limit` query parameters.
- Supports optional `user_id`, `from`, and `to` filters.
- Defaults to `page=1` and `limit=20`.
- Clamps invalid page values to `1`.
- Clamps invalid limit values to `20`.
- Caps `limit` at `100`.
- Keeps the total count independent from the capped page size.
- Orders activity logs by `created_at` descending.
- Returns `404 Not Found` when the project does not exist.
- Returns `403 Forbidden` when the user does not have project access.
- Returns an empty `logs` array when the project has no activity.

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
- Allows project members with access to persist the shared stage order.
- Accepts an ordered array of stage IDs.
- Rejects an empty `stage_ids` array.
- Rejects incomplete `stage_ids` arrays that omit existing project stages.
- Rejects duplicate stage IDs.
- Rejects stage IDs that do not belong to the project.
- Returns `404 Not Found` when the project does not exist.
- Updates all stage positions inside a single database transaction.
- Rolls back the transaction if any submitted stage ID is invalid.
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
- Returns `404 Not Found` when the project does not exist.
- Returns `403 Forbidden` when the user does not have project access.
- Returns `400 Bad Request` when the project ID is negative or zero.
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
- Rejects task create/update with `400 Bad Request` if `start_date` comes strictly after `deadline`.
- Returns `start_date` in task read responses.

Supported task create/update field:

```json
{
  "start_date": "2026-04-20T00:00:00Z"
}
```

## Frontend Unit Tests

Sprint 4 frontend unit test files include:

| File | Coverage |
|------|----------|
| `frontend/src/app/pages/profile/profile.component.spec.ts` | Profile rendering, valid save, invalid email rejection, unchanged form save state |
| `frontend/src/app/pages/home/home.component.spec.ts` | Board search, solo/shared filters, active refinement state, clear behavior |
| `frontend/src/app/pages/board/board.component.spec.ts` | Task search by title/description, trimmed case-insensitive matching, match counts, empty state support, clear filters |

Previously existing frontend unit tests also continue to cover app shell behavior, authentication service behavior, theme service behavior, notification service behavior, notification bell UI, board filters, board view calculations, and planner-board behavior.

## Cypress E2E Tests

Sprint 4 task-search Cypress coverage was added to `frontend/cypress/e2e/board.cy.ts`:

- `filters task cards by title and shows the result count`
- `filters task cards by description across columns`
- `clear search restores all visible tasks`
- `shows an empty state when no task matches the search`

The project also includes Cypress suites for:

| File | Coverage |
|------|----------|
| `frontend/cypress/e2e/welcome.cy.ts` | Welcome page and navigation |
| `frontend/cypress/e2e/login.cy.ts` | Login form and password visibility |
| `frontend/cypress/e2e/signup.cy.ts` | Signup form validation and navigation |
| `frontend/cypress/e2e/google-oauth.cy.ts` | Google OAuth buttons, callback, loading, and error states |
| `frontend/cypress/e2e/board.cy.ts` | Board task actions, filters, search, stages, details modal, top bar, comments, subtasks, and activity |
| `frontend/cypress/e2e/planner.cy.ts` | Planner loading, board/planner navigation, calendar tasks, month controls, and planner modal |
| `frontend/cypress/e2e/labels.cy.ts` | Label manager, label assignment, label filters, and dynamic priority display |
| `frontend/cypress/e2e/notifications.cy.ts` | Notification bell, dropdown, read state, and deadline reminders |
| `frontend/cypress/e2e/project-members.cy.ts` | Project member search, add, duplicate handling, remove, undo, and permissions |
| `frontend/cypress/e2e/task-comments.cy.ts` | Board/planner comments, author-only controls, posting, editing, deletion, and counts |
| `frontend/cypress/e2e/checklists.cy.ts` | Board/planner checklist creation, completion, deletion, and progress |
| `frontend/cypress/e2e/activity-history.cy.ts` | Activity tab, member/date filters, pagination, and empty state |

## Backend Unit Tests

Backend test coverage is organized under `backend/internal/testcases` and `backend/internal/auth`.

| File | Coverage |
|------|----------|
| `backend/internal/testcases/auth_service_test.go` | Email normalization/validation, auth request/response structures, auth errors |
| `backend/internal/testcases/auth_controller_test.go` | Auth controller invalid JSON, current-user guard, password reset flows |
| `backend/internal/testcases/jwt_service_test.go` | JWT generation, validation, expiration, claims, secret lookup |
| `backend/internal/testcases/otp_service_test.go` | OTP generation, verification, reset token validation |
| `backend/internal/testcases/user_repository_test.go` | User create/read/update and email existence checks |
| `backend/internal/testcases/database_test.go` | SQLite setup, operations, transaction behavior |
| `backend/internal/testcases/models_test.go` | Core model structures and response conversion |
| `backend/internal/testcases/controllers_test.go` | Project, stage, task, and message controller guards and constructors |
| `backend/internal/testcases/services_test.go` | Core project, stage, task, and message service constructors |
| `backend/internal/testcases/middleware_test.go` | JWT middleware and request context helpers |
| `backend/internal/testcases/task_enhancements_test.go` | Task deadline, priority, assignee, validation, and controller round trips |
| `backend/internal/testcases/comment_test.go` | Comment service/controller CRUD, validation, ownership, and unauthorized paths |
| `backend/internal/testcases/subtask_test.go` | Subtask service/controller CRUD, ordering, validation, and unauthorized paths |
| `backend/internal/testcases/project_member_service_test.go` | Membership, invites, roles, owner rules, pagination, and transaction safety |
| `backend/internal/testcases/activity_service_test.go` | Activity logging, filters, pagination, label activity, and concurrency |
| `backend/internal/testcases/label_service_test.go` | Label CRUD, task labels, duplicate names, color validation, and access checks |
| `backend/internal/testcases/notification_service_test.go` | Notification CRUD, read state, unread count, duplicate deadline reminders, and self-notification rules |
| `backend/internal/auth/controller/auth_controller_test.go` | Google OAuth controller request validation and configuration errors |
| `backend/internal/auth/services/auth_service_test.go` | Google login create/link/reactivate flows, verified email checks, invalid OAuth state |
| `backend/internal/auth/repository/auth_identity_repository_test.go` | Google auth identity upsert behavior |

Backend test command:

```bash
cd backend
go test -v ./internal/testcases/... ./internal/auth/...
```

### Detailed Backend Test Coverage

Added timeline-focused backend tests in:

```text
backend/internal/testcases/timeline_test.go
```

Added stage-reorder-focused backend tests in:

```text
backend/internal/testcases/stage_reorder_test.go
```

Added activity-endpoint-focused backend tests in:

```text
backend/internal/testcases/activity_endpoint_test.go
```

Added task-search-focused backend tests in:

```text
backend/internal/testcases/task_search_test.go
```

Test coverage includes:

Task search:

- Searches task titles and descriptions across a project.
- Returns only tasks from the requested project.
- Performs case-insensitive matching.
- Trims whitespace around the search query.
- Allows project members to search shared project tasks.
- Treats SQL wildcard characters as literal search text.
- Verifies ordering by stage position, task position, then task ID.
- Includes task metadata needed by frontend search results.
- Includes `stage_name` for result grouping and display.
- Returns an empty array when no tasks match.
- Rejects blank search queries.
- Rejects invalid project IDs.
- Rejects unauthenticated requests.
- Returns `404` for missing projects.
- Returns `403` for users without project access.

Activity feed:

- Returns paginated activity logs.
- Returns `logs`, `total`, and `page` fields.
- Orders logs by `created_at` descending.
- Returns the second page correctly.
- Normalizes invalid page and limit values to defaults.
- Caps overly large `limit` values at `100`.
- Filters activity by user and RFC3339 date ranges.
- Returns an empty `logs` array for projects with no activity.
- Rejects invalid project IDs.
- Rejects unauthenticated requests.
- Returns `404` for missing projects.
- Returns `403` for users without project access.

Stage reorder:

- Updates positions according to the submitted order.
- Persists new positions to the database.
- Returns stages ordered by the new positions.
- Rejects an empty `stage_ids` array.
- Rejects duplicate stage IDs.
- Rejects stage IDs from a different project.
- Rolls back partial updates when validation fails during reorder.
- Allows project members to reorder shared project stages.
- Rejects missing projects.
- Rejects users without project access.
- Confirms the controller returns `403 Forbidden` for inaccessible projects.
- Confirms controller returns updated ordered stages.

Timeline:

- Returns tasks with deadlines.
- Returns tasks with start dates but no deadline.
- Confirms tasks with deadlines are ordered before start-only tasks.
- Excludes tasks with no `start_date` and no `deadline`.
- Includes `stage_name` in each timeline item.
- Returns an empty array when there are no dated tasks.
- Allows access for project members.
- Rejects access for users outside the project.
- Returns `404` for missing projects.
- Maps project access errors to `403` and missing project errors to `404`.
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

Run activity feed tests:

```bash
cd backend
go test ./internal/testcases -run ActivityController_GetActivity
```

Run task search tests:

```bash
cd backend
go test ./internal/testcases -run Task.*Search
```

Run all backend tests:

```bash
cd backend
go test ./...
```

## Updated Backend API Documentation

Public endpoints:

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/health` | Health check |
| POST | `/api/auth/register` | Register a new user |
| POST | `/api/auth/login` | Log in with email/password |
| POST | `/api/auth/google/id-token` | Log in with a Google ID token |
| GET | `/api/auth/google/login` | Start Google OAuth redirect |
| GET | `/api/auth/google/callback` | Complete Google OAuth redirect |
| POST | `/api/auth/forgot-password` | Request password reset OTP |
| POST | `/api/auth/verify-otp` | Verify password reset OTP |
| POST | `/api/auth/reset-password` | Reset password |

Protected endpoints use `Authorization: Bearer <jwt>`.

| Resource | Endpoints |
|----------|-----------|
| Current user | `GET /api/auth/me` |
| Projects | `POST /api/projects`, `GET /api/projects`, `GET /api/projects/{id}`, `PUT /api/projects/{id}`, `DELETE /api/projects/{id}` |
| Members | `POST /api/projects/{id}/members`, `GET /api/projects/{id}/members`, `DELETE /api/projects/{id}/members/{userId}` |
| Invites | `POST /api/projects/{id}/invites`, `GET /api/invites/{id}`, `POST /api/invites/{id}/accept` |
| Stages | `POST /api/projects/{projectId}/stages`, `GET /api/projects/{projectId}/stages`, `GET /api/stages/{id}`, `PUT /api/stages/{id}`, `DELETE /api/stages/{id}` |
| Tasks | `POST /api/projects/{projectId}/stages/{stageId}/tasks`, `GET /api/projects/{projectId}/stages/{stageId}/tasks`, `GET /api/tasks/{id}`, `PUT /api/tasks/{id}`, `PUT /api/tasks/{id}/move`, `DELETE /api/tasks/{id}` |
| Comments | `POST /api/tasks/{id}/comments`, `GET /api/tasks/{id}/comments`, `PATCH /api/comments/{id}`, `DELETE /api/comments/{id}` |
| Subtasks | `POST /api/tasks/{id}/subtasks`, `GET /api/tasks/{id}/subtasks`, `PATCH /api/subtasks/{id}`, `DELETE /api/subtasks/{id}` |
| Messages | `POST /api/projects/{projectId}/messages`, `GET /api/projects/{projectId}/messages`, `GET /api/projects/{projectId}/messages/recent`, `DELETE /api/messages/{id}` |
| Activity | `GET /api/projects/{id}/activity`, `GET /api/projects/{id}/activity/recent` |
| Labels | `POST /api/projects/{id}/labels`, `GET /api/projects/{id}/labels`, `DELETE /api/labels/{id}`, `POST /api/tasks/{id}/labels`, `GET /api/tasks/{id}/labels`, `DELETE /api/tasks/{id}/labels/{labelId}` |
| Notifications | `GET /api/notifications`, `PATCH /api/notifications/read-all`, `PATCH /api/notifications/{id}/read` |
| Project chat | `WS /ws/{projectId}` |

### Detailed Backend Endpoint Documentation

### GET `/api/projects/{id}/tasks/search`

Purpose:

- Provides project-wide task search for the board search experience.
- Lets the frontend search all task columns without loading each stage
  separately.
- Returns compact task display fields plus stage information.

Path parameters:

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `id` | integer | yes | Project ID |

Query parameters:

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `q` | string | yes | Search text matched against task title and description |

Success response:

- Status: `200 OK`
- Body: plain JSON array of matching tasks.

Example request:

```http
GET /api/projects/10/tasks/search?q=backend
Authorization: Bearer <JWT>
```

Example response:

```json
[
  {
    "task_id": 42,
    "title": "Backend search endpoint",
    "description": "Add project task search",
    "stage_id": 3,
    "stage_name": "Review",
    "deadline": "2026-05-01T10:00:00Z",
    "priority": "high",
    "assigned_to": "user-123"
  }
]
```

Error responses:

| Status | Scenario |
|--------|----------|
| `400 Bad Request` | Project ID is not numeric |
| `400 Bad Request` | `q` is missing or blank |
| `401 Unauthorized` | Missing or invalid authentication |
| `403 Forbidden` | User is not a member of the project |
| `404 Not Found` | Project does not exist |
| `500 Internal Server Error` | Unexpected database or service error |

### GET `/api/projects/{id}/activity`

Purpose:

- Provides recent project activity for the Dashboard activity feed.
- Gives the frontend a paginated list of project events.
- Keeps the response focused on display fields for activity cards.

Path parameters:

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `id` | integer | yes | Project ID |

Query parameters:

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `page` | integer | no | `1` | Page number |
| `limit` | integer | no | `20` | Number of logs per page, max `100` |

Success response:

- Status: `200 OK`
- Body: activity feed object with `logs`, `total`, and `page`.

Example request:

```http
GET /api/projects/10/activity?page=1&limit=20
Authorization: Bearer <JWT>
```

Example response:

```json
{
  "logs": [
    {
      "id": 42,
      "user_name": "Owner User",
      "action": "task_moved",
      "entity_type": "task",
      "entity_title": "Moved task to Done",
      "created_at": "2026-04-28T12:00:00Z"
    }
  ],
  "total": 1,
  "page": 1
}
```

Error responses:

| Status | Scenario |
|--------|----------|
| `400 Bad Request` | Project ID is not numeric |
| `400 Bad Request` | `page` or `limit` parameter is provided but not a valid number |
| `400 Bad Request` | `from` or `to` date filters are provided but not in RFC3339 format |
| `401 Unauthorized` | Missing or invalid authentication |
| `403 Forbidden` | User is not a member of the project |
| `404 Not Found` | Project does not exist |
| `500 Internal Server Error` | Unexpected database or service error |

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

## Submission Notes

For the narrated video presentation, demonstrate:

- Taskify overview for someone new to the project.
- New Sprint 4 frontend functionality: profile page, board search/filter on the home page, and board task search.
- Main frontend workflows: authentication, boards, task details, labels, comments, checklists, planner, notifications, members, and activity.
- Backend API responsibilities and protected route structure.
- Unit test and Cypress test results, including Sprint 3 and Sprint 4 coverage.

### Additional Backend Demo Notes

Task search endpoint demo points:

- Show the board search box or explain the board search dependency.
- Search for a task title and show matching task results.
- Search for text from a task description and show that it also matches.
- Explain that results include `stage_name`, `priority`, `deadline`, and
  `assigned_to`.
- Mention that the endpoint enforces project access before returning results.

Activity feed endpoint demo points:

- Show the Dashboard activity section.
- Explain that the frontend calls the paginated activity endpoint.
- Show that the endpoint returns `logs`, `total`, and `page`.
- Explain that newest project activity appears first.
- Mention that missing projects return `404` and inaccessible projects return `403`.

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
