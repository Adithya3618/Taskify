# Sprint 3 — Taskify Project

## Team Member Assignments

| Developer | GitHub Issue | Feature |
|-----------|-------------|---------|
| saisreejachava (Sreeja) | #70, #71, #72, #73 | Google OAuth, Task Deadlines & Priority UI, Labels/Tags, Notification Center |
| *(teammate 2)* | #74–#77 | *(backend features)* |
| *(teammate 3)* | #78–#81 | *(backend/frontend features)* |
| *(teammate 4)* | #82–#85 | *(backend/frontend features)* |

---

## Frontend Work Completed — Sreeja (saisreejachava)

### Issue #70 — Google OAuth Integration (End-to-End)

- Added **"Sign in with Google"** button to the Login page with a loading state and a dedicated `googleError` field so errors appear *below* the Google button (not above the email form)
- Added **"Sign up with Google"** button to the Signup page using the same flow
- Implemented `startGoogleLogin()` in `AuthService`:
  - Makes a GET to `/api/auth/google/login`
  - status 0 / CORS error → Google redirect occurred → silently redirect
  - status 503 or 500 → backend not configured → surface error to user
- Implemented `handleGoogleToken()` in `AuthService`:
  - Accepts `{ token, name, email, id }` from the backend redirect query params
  - Stores JWT and user session in `localStorage`
- Created `GoogleCallbackComponent` at `/auth/google/callback`:
  - Reads `?token=` params set by the Go backend redirect
  - Calls `handleGoogleToken()` then navigates to `/boards`
  - Also handles `?error=` param for graceful error display
- Modified `backend/internal/auth/controller/auth_controller.go`:
  - `GoogleCallback` now redirects to `FRONTEND_URL/auth/google/callback?token=…&name=…&email=…&id=…` instead of returning JSON
- Configured `backend/.env` with `GOOGLE_REDIRECT_URL` and `FRONTEND_URL` pointing to the Angular dev server

**Key fix:** Separated `googleError` from the shared `error` field so Google OAuth errors never appear above the email/password form.

---

### Issue #71 — Task Deadlines & Dynamic Priority Escalation

- Added **due date field** (date picker) to the task creation inline form (`+ Add details`)
- Added **due date field** to the task detail/edit modal
- Due date is stored in the task description metadata block (`\n---\ndue:YYYY-MM-DD`) via `buildCardDescription` / `parseCardMeta` in `utils/task-card-meta.ts`
- Due date chip is shown on task cards with colour classes:
  - `.due-overdue` — past deadline (red)
  - `.due-today` — due today (amber)
- **Dynamic priority escalation** (display-only — stored priority is never changed):

  | Stored Priority | Condition | Displayed As |
  |----------------|-----------|--------------|
  | Any | Overdue | Urgent |
  | Low / Medium | Due today | High |
  | Low | Due tomorrow | Medium |
  | Any | No deadline / future | Stored value |

- `getEffectivePriority(task)` computes the display priority at render time
- All priority badges, filter chips, and card CSS classes use the effective priority
- Renamed "Critical" → "Urgent" throughout the UI for consistency with escalation language
- Task detail modal shows a tooltip indicating when a priority has been escalated

---

### Issue #72 — Labels / Tags UI

- **Label Manager modal** accessible from a "Labels" button in the board topbar:
  - Create a label (name + auto-assigned colour from a palette)
  - Delete a label
  - Labels are scoped per project and persisted in `localStorage` (key: `taskify.labels.<projectId>`)
- **Assign labels to existing tasks** — label picker (chip style) in the task detail modal
  - Clicking a chip toggles the label on/off for that task
  - Assignments persisted in `localStorage` (key: `taskify.taskLabels.<projectId>`)
- **Assign labels when creating a task** — label picker chips shown inside the `+ Add details` section of the inline task form
- **Label chips on task cards** — coloured chips shown above the meta row; an `+ Add label` chip appears if the project has labels but the task has none
- **Label filter in the filter panel** — "All" chip plus one chip per project label; selecting a label shows only tasks that have that label; integrated into `hasActiveFilters` and `clearFilters()`
- Fixed a CSS conflict where the modal's generic `label` element style overrode the `.labelCheckItem` class — resolved by replacing `<label>` wrapper elements with `<div>`s

---

### Issue #73 — Notification Center UI

- Created `AppNotification` model (`notification.model.ts`):

  ```typescript
  export type NotificationType = 'project_invite' | 'task_assigned' | 'deadline_reminder';
  export interface AppNotification {
    id: string; type: NotificationType; message: string;
    is_read: boolean; created_at: string; link?: string;
  }
  ```

- Created `NotificationService` (Angular signal-based, `notification.service.ts`):
  - `notifications` — readonly signal backed by `localStorage` key `taskify.notifications`
  - `unreadCount` — computed signal (count of unread)
  - `add(type, message, link?)` — deduplicates by message+unread, keeps max 50
  - `markRead(id)` / `markAllRead()` — update signal + persist
  - `checkDeadlines(tasks, projectId)` — called on board load; generates `deadline_reminder` notifications for overdue / due-today / due-tomorrow tasks
  - `timeAgo(isoDate)` — relative time formatter for the UI ("just now", "5m ago", "3h ago", "2d ago")

- Created `NotificationBellComponent` (standalone, `notification-bell/`):
  - Bell icon button with animated unread count badge (capped at 9+)
  - Dropdown panel with notification list, type icon (⏰ / 📋 / 🔔), message, relative timestamp
  - "Mark all read" button shown only when there are unread notifications
  - Empty state ("You're all caught up!") when no notifications exist
  - Click-outside backdrop closes the panel
  - Clicking a notification marks it read and navigates to its `link` if set

- Added `<app-notification-bell>` to:
  - `BoardComponent` topbar (between theme toggle and profile menu)
  - `HomeComponent` navbar (between theme toggle and profile menu)

- `BoardComponent.loadTasks()` calls `notificationService.checkDeadlines()` passing tasks with parsed due dates from `parseCardMeta(t.description).due`

---

## How to Run the App

```bash
# Backend
cd backend
cp .env.example .env   # fill in GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, DB_*
go run cmd/server/main.go

# Frontend
cd frontend
npm install
npm start              # http://localhost:4200
```

---

## Testing

### How to run Karma unit tests (frontend)
```bash
cd frontend
npm test                                              # watch mode
npm test -- --watch=false --browsers=ChromeHeadless  # single headless run
```

### How to run Cypress E2E tests
```bash
cd frontend
# Terminal 1
npm start

# Terminal 2 — interactive
npm run cy:open

# Or headless
npm run cy:run
```

### How to run backend Go tests
```bash
cd backend
go test -v ./internal/testcases/...
```

---

## Frontend Unit Tests (Karma / Jasmine) — 99 total

> Includes all 75 tests from Sprint 2 plus 24 new tests for Sprint 3.

### `ThemeService` — 8 tests *(Sprint 2)*
| Test | Function under test |
|------|---------------------|
| should default to dark mode when no saved preference | `isDark` getter |
| should initialise to light mode when localStorage has "light" | constructor / `_apply` |
| should toggle from dark to light | `toggle()` |
| should toggle from light back to dark | `toggle()` |
| should persist "light" in localStorage after toggling to light | `toggle()` |
| should persist "dark" in localStorage after toggling back to dark | `toggle()` |
| should set data-theme="dark" on the html element in dark mode | `_apply()` |
| should set data-theme="light" on the html element after toggling | `_apply()` |

### `AuthService` — 12 tests *(Sprint 2)*
| Test | Function under test |
|------|---------------------|
| should return false when no token is stored | `isAuthenticated()` |
| should return true when a token exists in localStorage | `isAuthenticated()` |
| should return null when no token is set | `getToken()` |
| should return the token stored in localStorage | `getToken()` |
| should return null when no session is stored | `getCurrentUser()` |
| should return the parsed user from localStorage | `getCurrentUser()` |
| should return null when session contains invalid JSON | `getCurrentUser()` |
| should remove token and session from localStorage on logout | `logout()` |
| should return null from updateCurrentUser when no user is logged in | `updateCurrentUser()` |
| should merge patch into the current user and persist it | `updateCurrentUser()` |
| should store token and session after successful login | `login()` |
| should store token and session after successful registration | `register()` |

### `BoardComponent` — 22 tests *(Sprint 2)*
| Test | Function under test |
|------|---------------------|
| should start with filter panel closed | initial state |
| toggleFilterPanel() should open the filter panel | `toggleFilterPanel()` |
| toggleFilterPanel() should close the filter panel when called again | `toggleFilterPanel()` |
| hasActiveFilters should be false when no filters are set | `hasActiveFilters` |
| hasActiveFilters should be true when filterPriority is set | `hasActiveFilters` |
| hasActiveFilters should be true when filterDue is set | `hasActiveFilters` |
| clearFilters() should reset filterPriority and filterDue | `clearFilters()` |
| getFilteredTasks() should return all tasks when no filters are active | `getFilteredTasks()` |
| getFilteredTasks() should filter tasks by priority | `getFilteredTasks()` |
| getFilteredTasks() should return empty array when no tasks match priority filter | `getFilteredTasks()` |
| getFilteredTasks() should filter by "none" due date (tasks with no due date) | `getFilteredTasks()` |
| openShareModal() should set showShareModal to true | `openShareModal()` |
| openShareModal() should reset shareLinkCopied to false | `openShareModal()` |
| closeShareModal() should set showShareModal to false | `closeShareModal()` |
| closeShareModal() should reset shareLinkCopied to false | `closeShareModal()` |
| boardUrl should return the current window location href | `boardUrl` getter |
| toggleBoardSwitcher() should open the board switcher | `toggleBoardSwitcher()` |
| toggleBoardSwitcher() should close the board switcher when called again | `toggleBoardSwitcher()` |
| switchBoard() should navigate to the new board | `switchBoard()` |
| switchBoard() should not navigate when selecting the current board | `switchBoard()` |
| switchBoard() should close the board switcher | `switchBoard()` |
| goBack() should navigate to /boards | `goBack()` |

### `PlannerBoardComponent` — 30 tests *(Sprint 2)*
| Test | Function / behavior under test |
|------|----------------------------------|
| should create | initial load, `loading`, `projectId` |
| should open Scheduled panel when its header button is clicked | `scheduledOpen` |
| should close Scheduled panel when header button is clicked again | toggle closed |
| toggleScheduledOpen() should flip scheduledOpen | `toggleScheduledOpen()` |
| should open No due date panel when its header button is clicked | `noDueOpen` |
| toggleNoDueOpen() should flip noDueOpen | `toggleNoDueOpen()` |
| unscheduledFiltered should include all no-due tasks when filter is null | `unscheduledFiltered` |
| unscheduledFiltered should only include tasks from selected list | `noDueListFilterStageId` |
| should show list filter select when No due date is open and stages exist | `#noDueListFilter` |
| clicking previous month in main header should call prevMonth | header nav |
| clicking next month in main header should call nextMonth | header nav |
| clicking Today in main header should reset view to current month | `.btn-today` |
| clicking first mini nav button should go to previous month | `.btn-mini-nav` |
| clicking second mini nav button should go to next month | `.btn-mini-nav` |
| openMonthYearPicker should show dialog and set draft | `openMonthYearPicker()` |
| clicking main month label button should open month/year picker | `#planner-cal-title` |
| closeMonthYearPicker should hide dialog | `closeMonthYearPicker()` |
| applyMonthYear should update viewMonth for valid draft | `applyMonthYear()` |
| extraTasksOnSameDateLabel should return +1 task for two tasks | `extraTasksOnSameDateLabel()` |
| extraTasksOnSameDateLabel should return +2 tasks for three tasks | `extraTasksOnSameDateLabel()` |
| calendarDayToggleLabel should show +N tasks when multiple tasks and collapsed | `calendarDayToggleLabel()` |
| tasksForCalendarCell should return one task until expanded | `tasksForCalendarCell()` |
| visibleScheduledTasks should show one task until group expanded | `visibleScheduledTasks()` |
| toggleBoardSwitcher should flip showBoardSwitcher | `toggleBoardSwitcher()` |
| clicking theme toggle button should call themeService.toggle | `.themeToggle` |
| goBack should navigate to /boards | `goBack()` |
| onDayCellClick on current-month empty cell should open add-task modal | `onDayCellClick()` |
| onDayCellClick on other-month cell should change view month | `onDayCellClick()` |
| formatDueLabel should format YYYY-MM-DD | `formatDueLabel()` |
| should redirect to login when user is missing | unauthenticated → `/login` |

### `AppComponent` — 3 tests *(Sprint 2)*
| Test | What it checks |
|------|----------------|
| should create the app | component instantiation |
| should have the 'taskify' title | `title` property |
| should render a router-outlet | template structure |

### `NotificationService` — 24 tests *(Sprint 3 — NEW)*
| Test | Function under test |
|------|---------------------|
| should start with an empty notifications list when localStorage is empty | `load()` / signal init |
| should load persisted notifications from localStorage on creation | `load()` |
| should return empty list when localStorage contains invalid JSON | `load()` error handling |
| should report unreadCount of 0 when no notifications exist | `unreadCount` computed |
| should compute unreadCount for multiple unread notifications | `unreadCount` computed |
| should decrease unreadCount after a notification is marked read | `unreadCount` + `markRead()` |
| should add a notification with correct type, message, and is_read=false | `add()` |
| should generate a unique id for each notification | `add()` |
| should persist the new notification to localStorage | `add()` |
| should prepend new notifications so the latest appears first | `add()` |
| should not add a duplicate if an unread notification with the same message exists | `add()` dedup |
| should allow re-adding the same message once the original has been read | `add()` dedup |
| should store the optional link when provided | `add()` |
| should cap the notifications list at 50 entries | `add()` sliding window |
| should mark only the targeted notification as read | `markRead()` |
| should persist the read state after markRead | `markRead()` |
| should mark all notifications as read and set unreadCount to 0 | `markAllRead()` |
| should persist the fully-read state to localStorage | `markAllRead()` |
| should add an "overdue" notification for a task with a past deadline | `checkDeadlines()` |
| should add a "due today" notification for a task due today | `checkDeadlines()` |
| should add a "due tomorrow" notification for a task due tomorrow | `checkDeadlines()` |
| should not add a notification for tasks with no deadline | `checkDeadlines()` |
| should not add a notification for a task due more than one day away | `checkDeadlines()` |
| should set the notification link to the correct /board/:id URL | `checkDeadlines()` |
| should include the task title in the notification message | `checkDeadlines()` |
| should return "just now" for a timestamp less than 1 minute ago | `timeAgo()` |
| should return minutes ago for a timestamp 30 minutes ago | `timeAgo()` |
| should return hours ago for a timestamp 3 hours ago | `timeAgo()` |
| should return days ago for a timestamp 2 days ago | `timeAgo()` |

### `NotificationBellComponent` — 12 tests *(Sprint 3 — NEW)*
| Test | Function under test |
|------|---------------------|
| should create the component | instantiation |
| should start with the dropdown panel closed | `isOpen` initial state |
| toggle() should open the panel | `toggle()` |
| toggle() should close the panel when called again | `toggle()` |
| close() should set isOpen to false | `close()` |
| markAllRead() should delegate to notifService and set unreadCount to 0 | `markAllRead()` |
| onNotificationClick() should mark the notification as read | `onNotificationClick()` |
| onNotificationClick() should close the panel | `onNotificationClick()` |
| onNotificationClick() should navigate to the notification link when present | `onNotificationClick()` |
| onNotificationClick() should not navigate when the notification has no link | `onNotificationClick()` |
| iconForType() should return clock emoji for deadline_reminder | `iconForType()` |
| iconForType() should return clipboard emoji for task_assigned | `iconForType()` |
| iconForType() should return bell emoji for project_invite | `iconForType()` |

---

## Frontend Cypress E2E Tests — 105 total

> Includes all 70 tests from Sprint 2 plus 35 new tests for Sprint 3.

### Sprint 2 E2E tests — 70 tests *(carried forward)*

- Welcome page (`welcome.cy.ts`) — 6 tests
- Login page (`login.cy.ts`) — 6 tests
- Signup page (`signup.cy.ts`) — 7 tests
- Board (`board.cy.ts`) — 51 tests (checkboxes, collapse, due date, filters, add task/stage, modal, delete, topbar)
- Planner (`planner.cy.ts`) — N tests

### `google-oauth.cy.ts` — 10 tests *(Sprint 3 — NEW — Issue #70)*
| Test |
|------|
| shows a "Sign in with Google" button on the login page |
| Google button is enabled by default |
| shows a loading state on the Google button while the request is in flight |
| shows an error message below the Google button when the backend returns 503 |
| shows an error message below the Google button when the backend returns 500 |
| error message appears below the Google button, not above the email form |
| has a divider separating email login from Google login |
| shows a "Sign up with Google" button on the signup page |
| shows an error below the Google button on signup when backend returns 503 |
| has a divider separating email signup from Google signup |
| redirects to /boards after handling a token in the callback URL |

### `labels.cy.ts` — 14 tests *(Sprint 3 — NEW — Issues #71 & #72)*
| Test |
|------|
| shows a Labels button in the board topbar |
| opens the Label Manager modal when the Labels button is clicked |
| closes the Label Manager modal with the X button |
| shows an input field for the new label name |
| creates a new label and shows it in the label list |
| clears the input after adding a label |
| can delete a label from the manager |
| shows a label section in the task detail modal |
| shows label chips for selection in the task detail modal |
| shows label picker in the add task form when expanded |
| shows a Label section in the filter panel |
| All labels chip is active by default in the filter panel |
| shows "Urgent" priority badge for a task with Low priority that is overdue |
| shows "High" effective priority for a Medium-priority task due today |
| shows the due-overdue class on the due date label for an overdue task |
| shows the due-today class on the due date label for a task due today |
| task detail modal priority selector reflects dynamic priority for overdue task |

### `notifications.cy.ts` — 18 tests *(Sprint 3 — NEW — Issue #73)*
| Test |
|------|
| shows the notification bell button in the board topbar |
| shows no unread badge when there are no notifications |
| shows an unread badge with count when there is an unread notification |
| caps the badge at 9+ when there are more than 9 unread notifications |
| opens the notification panel when the bell is clicked |
| closes the panel when clicking the bell again |
| closes the panel when clicking the backdrop |
| shows "You're all caught up!" when there are no notifications |
| shows the notification message in the panel |
| shows "Mark all read" button when there are unread notifications |
| removes the unread badge after clicking Mark all read |
| shows clock emoji for deadline_reminder type notifications |
| auto-generates an overdue notification when board loads with an overdue task |
| auto-generates a "due today" notification on board load |
| does NOT add a notification for a task due next week |
| shows the notification bell on the home/boards page |
| opens the notification panel from the home page |
| shows empty state on home page when no notifications exist |
| shows unread badge on home page when notification was seeded |

---

## Backend Unit Tests (Go) — 100+ tests *(Sprint 2 — carried forward)*

All backend tests are in `backend/internal/testcases/` and run with:
```bash
cd backend
go test -v ./internal/testcases/...
```

See [Sprint2.md](Sprint2.md) for the full list of backend test cases.

---

## Updated Backend API Documentation

### Auth Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/auth/register` | None | Register with name/email/password |
| POST | `/api/auth/login` | None | Login, returns JWT |
| POST | `/api/auth/forgot-password` | None | Send OTP to email |
| POST | `/api/auth/verify-otp` | None | Verify OTP, returns reset token |
| POST | `/api/auth/reset-password` | None | Reset password with token |
| GET | `/api/auth/google/login` | None | Redirect to Google OAuth consent screen |
| GET | `/api/auth/google/callback` | None | OAuth callback → redirects to `FRONTEND_URL/auth/google/callback?token=…` |
| POST | `/api/auth/google/id-token` | None | Login via Google ID token (mobile) |
| GET | `/api/auth/me` | JWT | Get current authenticated user |

### Project Endpoints (JWT required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/projects` | List all projects |
| POST | `/api/projects` | Create a project |
| GET | `/api/projects/:id` | Get project by ID |
| PUT | `/api/projects/:id` | Update project |
| DELETE | `/api/projects/:id` | Delete project |

### Stage Endpoints (JWT required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/projects/:id/stages` | List stages for a project |
| POST | `/api/projects/:id/stages` | Create a stage |
| GET | `/api/stages/:id` | Get stage by ID |
| PUT | `/api/stages/:id` | Update stage |
| DELETE | `/api/stages/:id` | Delete stage |

### Task Endpoints (JWT required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/projects/:id/stages/:stageId/tasks` | List tasks in a stage |
| POST | `/api/projects/:id/stages/:stageId/tasks` | Create a task |
| GET | `/api/tasks/:id` | Get task by ID |
| PUT | `/api/tasks/:id` | Update task (title, description, position) |
| DELETE | `/api/tasks/:id` | Delete task |

> **Note:** Due date, priority, labels, and notes are encoded in the task `description` field using the metadata block format:
> ```
> <user description>
> ---
> due:YYYY-MM-DD
> priority:Low|Medium|High|Urgent
> notes:<free text>
> ```

### Message Endpoints (JWT required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/projects/:id/messages` | Get messages for a project |
| POST | `/api/projects/:id/messages` | Post a message |
| DELETE | `/api/messages/:id` | Delete a message |

---

## Commit History — Sreeja (saisreejachava)

Key commits on branch `sreeja/frontendDev`:

- `Google OAuth end-to-end, sign up with Google, graceful error handling and dynamic priority escalation - Closes #70 #71`
- `feat: Google OAuth frontend integration - Closes #70`
- Labels/Tags UI — Closes #72
- Notification Center UI (bell, service, deadline reminders, home page) — Closes #73
- Sprint 3 Cypress E2E tests and Karma unit tests
