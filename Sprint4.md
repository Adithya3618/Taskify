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

## Submission Notes

For the narrated video presentation, demonstrate:

- Taskify overview for someone new to the project.
- New Sprint 4 frontend functionality: profile page, board search/filter on the home page, and board task search.
- Main frontend workflows: authentication, boards, task details, labels, comments, checklists, planner, notifications, members, and activity.
- Backend API responsibilities and protected route structure.
- Unit test and Cypress test results, including Sprint 3 and Sprint 4 coverage.
