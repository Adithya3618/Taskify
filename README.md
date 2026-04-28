# Taskify

Taskify is a full-stack project and task management application for teams that need one place to plan work, track progress, and collaborate. The app combines project boards, Kanban task stages, task details, comments, checklists, labels, notifications, activity history, and project member management.

## Features

- Authentication with email/password and Google OAuth.
- Project dashboard with owned/shared board visibility, search, and filters.
- Kanban board with custom stages, task creation, editing, deletion, completion, and movement.
- Task search and filter tools for title, description, completion, due date, priority, and labels.
- Task metadata including priority, due dates, dynamic deadline escalation, notes, labels, comments, and checklist progress.
- Project collaboration with members, invites, owner permissions, activity history, and project chat.
- Notification center for project invites, assignments, and deadline reminders.
- Profile page for viewing and updating account details.
- Planner/calendar view for scheduled tasks.

## Tech Stack

Frontend:
- Angular 17
- TypeScript
- HTML/CSS
- Jasmine/Karma unit tests
- Cypress end-to-end tests

Backend:
- Go
- Gorilla Mux REST API
- SQLite
- JWT authentication
- WebSocket project chat
- Go unit tests

## Requirements

Install these before running the project:

- Node.js 20.x and npm
- Angular CLI 17.x
- Go 1.22 or newer
- SQLite support for Go through `github.com/mattn/go-sqlite3`

## Environment Setup

Backend environment variables are stored in `backend/.env`. A template is available at `backend/.env.example`.

Required or commonly used variables:

```bash
JWT_SECRET=your-secure-jwt-secret
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_EMAIL=your-email@example.com
SMTP_PASSWORD=your-app-password
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/auth/google/callback
GOOGLE_OAUTH_SCOPES=openid email profile
FRONTEND_URL=http://localhost:4200
```

The backend creates `taskify.db` in the directory where the Go server is started.

## Running Locally

Start the backend:

```bash
cd backend
go mod tidy
go run cmd/server/main.go
```

The backend runs on:

```text
http://localhost:8080
```

Start the frontend in another terminal:

```bash
cd frontend
npm install
npm start
```

The frontend runs on:

```text
http://localhost:4200
```

The Angular dev server uses `frontend/proxy.conf.json` to proxy `/api` to `http://localhost:8080` and `/ws` to `ws://localhost:8080`.

## Using the Application

1. Open `http://localhost:4200`.
2. Create an account, log in, or use Google OAuth if Google credentials are configured.
3. Create a project from the boards page.
4. Open a board to create stages and tasks.
5. Use the board controls to search tasks, filter by status, priority, due date, or label, and switch between Kanban, planner, dashboard, table, and timeline views.
6. Open a task to edit details, add comments, manage checklist items, assign labels, and update due dates or priority.
7. Use project settings to manage members and the activity tab to review recent project changes.
8. Use the notification bell to review unread updates and deadline reminders.
9. Use the profile menu to open account settings.

## Testing

Frontend unit tests:

```bash
cd frontend
npm test -- --watch=false
```

Targeted board unit tests:

```bash
cd frontend
npm test -- --watch=false --include src/app/pages/board/board.component.spec.ts
```

Cypress end-to-end tests:

```bash
cd frontend
npm run cy:run
```

Targeted board Cypress tests:

```bash
cd frontend
npm run cy:run -- --spec cypress/e2e/board.cy.ts
```

Backend unit tests:

```bash
cd backend
go test -v ./internal/testcases/... ./internal/auth/...
```

## Backend API Overview

Public endpoints:

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/health` | Health check |
| POST | `/api/auth/register` | Register a user |
| POST | `/api/auth/login` | Log in with email/password |
| POST | `/api/auth/google/id-token` | Log in with a Google ID token |
| GET | `/api/auth/google/login` | Start Google OAuth redirect |
| GET | `/api/auth/google/callback` | Complete Google OAuth redirect |
| POST | `/api/auth/forgot-password` | Request password reset OTP |
| POST | `/api/auth/verify-otp` | Verify password reset OTP |
| POST | `/api/auth/reset-password` | Reset password |

Protected endpoints require `Authorization: Bearer <jwt>`.

| Resource | Endpoints |
|----------|-----------|
| Current user | `GET /api/auth/me` |
| Projects | `POST /api/projects`, `GET /api/projects`, `GET /api/projects/{id}`, `PUT /api/projects/{id}`, `DELETE /api/projects/{id}` |
| Members and invites | `POST /api/projects/{id}/members`, `GET /api/projects/{id}/members`, `DELETE /api/projects/{id}/members/{userId}`, `POST /api/projects/{id}/invites`, `GET /api/invites/{id}`, `POST /api/invites/{id}/accept` |
| Stages | `POST /api/projects/{projectId}/stages`, `GET /api/projects/{projectId}/stages`, `GET /api/stages/{id}`, `PUT /api/stages/{id}`, `DELETE /api/stages/{id}` |
| Tasks | `POST /api/projects/{projectId}/stages/{stageId}/tasks`, `GET /api/projects/{projectId}/stages/{stageId}/tasks`, `GET /api/tasks/{id}`, `PUT /api/tasks/{id}`, `PUT /api/tasks/{id}/move`, `DELETE /api/tasks/{id}` |
| Comments | `POST /api/tasks/{id}/comments`, `GET /api/tasks/{id}/comments`, `PATCH /api/comments/{id}`, `DELETE /api/comments/{id}` |
| Subtasks | `POST /api/tasks/{id}/subtasks`, `GET /api/tasks/{id}/subtasks`, `PATCH /api/subtasks/{id}`, `DELETE /api/subtasks/{id}` |
| Messages | `POST /api/projects/{projectId}/messages`, `GET /api/projects/{projectId}/messages`, `GET /api/projects/{projectId}/messages/recent`, `DELETE /api/messages/{id}` |
| Activity | `GET /api/projects/{id}/activity`, `GET /api/projects/{id}/activity/recent` |
| Labels | `POST /api/projects/{id}/labels`, `GET /api/projects/{id}/labels`, `DELETE /api/labels/{id}`, `POST /api/tasks/{id}/labels`, `GET /api/tasks/{id}/labels`, `DELETE /api/tasks/{id}/labels/{labelId}` |
| Notifications | `GET /api/notifications`, `PATCH /api/notifications/read-all`, `PATCH /api/notifications/{id}/read` |
| Chat | `WS /ws/{projectId}` |

## Team

- Adithya - Backend development
- Jyothi Nandhan Repaka - Backend development
- Sai Meghana Barla - Frontend development
- Sai Sreeja Chava - Frontend development
