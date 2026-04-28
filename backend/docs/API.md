# Taskify API Documentation

## Overview
Taskify is a task management backend API built with Go, following clean architecture (Controller → Service → Repository).

---

## Authentication

All protected endpoints require JWT Bearer token in Authorization header:
```
Authorization: Bearer <token>
```

### Endpoints

#### POST /api/auth/register
Register a new user.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "name": "John Doe"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe"
  },
  "token": "jwt.token.here"
}
```

#### POST /api/auth/login
Login with email/password.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Response:**
```json
{
  "success": true,
  "data": {...},
  "token": "jwt.token.here"
}
```

#### GET /api/auth/me (Protected)
Get current authenticated user.

**Response:**
```json
{
  "id": "uuid",
  "name": "John Doe",
  "email": "user@example.com",
  "role": "user",
  "is_active": true,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

#### PUT /api/auth/me (Protected)
Update current user's display name.

**Request:**
```json
{
  "name": "New Name"  // Max 100 characters
}
```

**Responses:**
- Success: `200` with updated user
- Empty name: `400` "name is required"
- Too long: `400` "name must be 100 characters or less"

---

## Projects

#### POST /api/projects (Protected)
Create a new project.

**Request:**
```json
{
  "name": "Project Name",
  "description": "Optional description"
}
```

#### GET /api/projects (Protected)
List all projects for authenticated user.

#### GET /api/projects/:id (Protected)
Get project by ID.

#### GET /api/projects/:id/members (Protected)
List project members.

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "user_id": "uuid",
      "name": "John Doe",
      "email": "user@example.com",
      "role": "owner"
    }
  ]
}
```

#### GET /api/projects/:id/stats (Protected)
Get project statistics.

**Response:**
```json
{
  "success": true,
  "data": {
    "total_tasks": 10,
    "completed_tasks": 5,
    "overdue_tasks": 2,
    "tasks_by_stage": [
      {"stage_id": 1, "stage_name": "To Do", "count": 3},
      {"stage_id": 2, "stage_name": "Done", "count": 5}
    ],
    "completion_rate": 50.0
  }
}
```

---

## Tasks

#### POST /api/projects/:projectId/stages/:stageId/tasks (Protected)
Create a task in a stage.

#### PUT /api/tasks/:id/assign (Protected)
Assign or unassign a task.

**Request:**
```json
{
  "assigned_to": "user-uuid"  // or null to unassign
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "title": "Task Title",
    "assigned_to": "user-uuid" | null,
    "stage_id": 1,
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

**Error Codes:**
- `404` - Task not found
- `403` - Requester not a project member  
- `400` - Assignee not a project member

---

## Stages

#### POST /api/projects/:id/stages (Protected)
Create a stage.

**Request:**
```json
{
  "name": "To Do",
  "position": 1,
  "is_final": false  // true if this is a completion stage
}
```

---

## Member Invites

#### POST /api/projects/:id/invites (Protected)
Create an invite.

**Request:**
```json
{
  "email": "invitee@example.com",
  "role": "member"
}
```

#### POST /api/projects/:id/members/accept (Protected)
Accept invite code.

**Request:**
```json
{
  "code": "invite-code-here"
}
```

---

## Error Responses

All errors follow this format:
```json
{
  "success": false,
  "error": "Human readable message",
  "code": "ERROR_CODE"
}
```

Common HTTP Status Codes:
- `200` - Success
- `201` - Created
- `400` - Bad Request / Validation Error
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `500` - Internal Server Error