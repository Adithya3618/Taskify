# Sprint 3 — Taskify Project

## Team Member Assignments

| Developer | GitHub Issue | Feature |
|-----------|-------------|---------|
| saisreejachava (Sreeja) | #70, #71, #72, #73 | Google OAuth, Task Deadlines & Priority UI, Labels/Tags, Notification Center |
| nandhan (Jyothi Nandhan Repaka) | #67, #68, #69 | Task Enhancements API, Task Comments API, Subtasks / Checklists API |
| meghana21-arch (Sai Meghana Barla) | #74, #75, #76, #77 | Project Member Management UI, Task Comments UI, Subtasks/Checklists UI, Activity History UI |
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

## Backend Work Completed — Nandhan (`nandhan/backend`)

### Issue #67 — Task Enhancements API

- Extended the `tasks` schema with first-class backend fields for:
  - `deadline`
  - `priority`
  - `assigned_to`
- Updated `Task` and task request/response handling so deadline, priority, assignee, subtask count, and completed subtask count are returned by the API
- Added priority validation in `TaskService` with normalized allowed values:
  - `low`
  - `medium`
  - `high`
  - `urgent`
- Added normalization logic so blank values can be cleared and invalid priorities are rejected with `ErrInvalidTaskPriority`
- Updated task create and update flows to persist these fields in SQLite
- Enhanced task read queries to return:
  - `subtask_count`
  - `completed_count`
- Added backend tests in `backend/internal/testcases/task_enhancements_test.go` covering:
  - create / read / update / clear task enhancements
  - invalid priority rejection in the service layer
  - invalid priority handling in the task controller

### Issue #68 — Task Comments: DB Schema & CRUD API

- Added a new `comments` table in the backend database schema with:
  - `task_id`
  - `user_id`
  - `author_name`
  - `content`
  - `created_at`
  - `updated_at`
- Added indexes for comment lookups by:
  - `task_id`
  - `user_id`
- Implemented `CommentService` with full CRUD behavior:
  - create comment
  - list comments for a task
  - update comment
  - delete comment
- Enforced task-level access checks before reading or creating comments
- Added author-name resolution from the authenticated user record
- Added validation for empty comment content using `ErrCommentContentRequired`
- Added comment routes:
  - `POST /api/tasks/{id}/comments`
  - `GET /api/tasks/{id}/comments`
  - `PATCH /api/comments/{id}`
  - `DELETE /api/comments/{id}`

### Issue #69 — Subtasks / Checklists: DB Schema & CRUD API

- Added a new `subtasks` table in the backend database schema with:
  - `task_id`
  - `title`
  - `is_completed`
  - `position`
  - `created_at`
  - `updated_at`
- Added indexes for subtask lookup and ordering:
  - `idx_subtasks_task`
  - `idx_subtasks_task_position`
- Implemented `SubtaskService` with full CRUD behavior:
  - create subtask
  - list subtasks for a task
  - update subtask title / completion / position
  - delete subtask
- Used SQL transactions for safe subtask reordering and deletion compaction
- Added validation for:
  - empty subtask titles
  - invalid insert positions
  - invalid move positions
- Added subtask routes:
  - `POST /api/tasks/{id}/subtasks`
  - `GET /api/tasks/{id}/subtasks`
  - `PATCH /api/subtasks/{id}`
  - `DELETE /api/subtasks/{id}`

### Backend Files Updated for These Issues

- `backend/internal/database/database.go`
- `backend/internal/models/models.go`
- `backend/internal/controllers/task_controller.go`
- `backend/internal/controllers/comment_controller.go`
- `backend/internal/controllers/subtask_controller.go`
- `backend/internal/services/task_service.go`
- `backend/internal/services/comment_service.go`
- `backend/internal/services/subtask_service.go`
- `backend/internal/routes/routes.go`
- `backend/internal/testcases/task_enhancements_test.go`

---

## Frontend Work Completed — Sai Meghana Barla (meghana21-arch)

### Issue #74 — Project Member Management UI

- Added a dedicated **Members** tab inside the project settings modal so member management stays within the existing board/project workflow
- Implemented **search / invite by name or email** with an **Add Member** action:
  - supports searching known users already available in the frontend session
  - also accepts direct email entry for invites when no cached match is available
- Rendered the current project member roster with:
  - avatar / initials fallback
  - display name
  - email
  - role badge
- Added **owner-only remove controls** so only the project owner can remove members from the project
- Added a **confirmation step before member deletion** and a short **Undo window** after removal so accidental deletes can be reversed
- Added **member count** to the project card on the home page so project access size is visible at a glance
- Updated the frontend member service / API integration layer to:
  - fetch project members
  - add members
  - remove members
  - merge backend responses with cached frontend member state so the member list still appears after page refresh even when backend persistence is incomplete
- Added duplicate-member validation so inviting a user who is already in the project shows a clear inline error instead of silently duplicating the row
- Member add / remove flows update the UI immediately without requiring a manual refresh

---

### Issue #75 — Task Comments UI

- Added a **comments section at the bottom of the task detail modal** on the board view
- Mirrored the same comment experience into the **planner task detail modal** so comments behave consistently across both task surfaces
- Added a **textarea composer** with a **Post** button for creating new comments
- Prevented empty submissions by trimming whitespace and disabling / rejecting blank comment content on the frontend
- Rendered comment entries with:
  - author avatar
  - author name
  - comment body
  - relative / formatted timestamp
- Added **author-only edit and delete actions** so only the person who wrote a comment can modify it
- Implemented **inline editing** with Save / Cancel actions
- Added **delete confirmation** before removing a comment
- Added **auto-scroll to the latest comment** when the task detail opens and when a new comment is posted, so the newest discussion stays in view
- Added a **comment icon + comment count** on task cards so discussion activity is visible from the board / planner surfaces
- Updated frontend comment API service calls for create / list / update / delete and kept the UI in sync immediately after each action without a full page reload

---

### Issue #76 — Subtasks / Checklists UI

- Added a **checklist / subtasks section** inside the task detail modal
- Added a lightweight composer so users can create new checklist items by:
  - pressing **Enter**
  - clicking the **Add** button
- Displayed checklist items in their stored order and rendered each item with:
  - title
  - completion checkbox
  - delete action
- Toggling a checkbox updates completion state immediately in the modal and on the task card without page refresh
- Added delete confirmation for checklist item removal to match the rest of the task-detail UX
- Added a **progress summary on task cards** so cards show how many subtasks are complete (for example, `1/3 done`)
- Mirrored checklist support into the **planner modal**, keeping board and planner task details aligned
- Updated subtasks API service calls for create / list / update / delete and kept frontend state synchronized immediately after every change

---

### Issue #77 — Activity History UI

- Added an **Activity** tab inside project settings for project owners / managers
- Restricted visibility of the Activity tab so non-owners do not see the activity history surface
- Implemented an activity feed that renders each entry with:
  - activity icon
  - human-readable action description
  - relative timestamp
- Rendered entries in **reverse chronological order** so the newest activity appears first
- Added filters for:
  - team member
  - date range
- Filter changes update the activity list immediately without a full page reload
- Added **Load more** style pagination so larger history lists can be browsed without overwhelming the initial settings view
- Added an **empty state** when no activity exists for the project yet
- Updated frontend activity service / API calls and local state handling so the activity feed stays consistent with the project settings UI

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

## Frontend Cypress E2E Tests — 142 total

> Includes all 70 tests from Sprint 2 plus 72 new tests for Sprint 3.

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

### `project-members.cy.ts` — 10 tests *(Sprint 3 — NEW — Issue #74)*
| Test |
|------|
| opens the settings modal on the Members tab from the top bar button |
| keeps the Add Member button disabled until a search value is entered |
| shows matching known users while typing in the member search field |
| adds a member when a search result is clicked |
| adds a member from a direct email entry with the Add Member button |
| shows a clear error when trying to add a duplicate member |
| hides remove buttons for non-owners |
| keeps a member in the list when removal is cancelled |
| shows an undo toast after removing a member and restores the member when Undo is clicked |
| finalizes the member removal after the undo window expires |

### `task-comments.cy.ts` — 11 tests *(Sprint 3 — NEW — Issue #75)*
| Test |
|------|
| shows the board comment count on the task card |
| shows the board empty state when a task has no comments |
| keeps the board Post button disabled until comment text is entered |
| shows edit and delete actions only for the board comment author |
| cancels a board inline comment edit without saving changes |
| auto-scrolls to the latest board comment when task details open |
| shows the planner comment count on the task chip |
| keeps the planner Post button disabled until text is entered |
| shows author-only actions in the planner comment list |
| keeps the planner comment when deletion is cancelled |
| posts a planner comment from the Post button and updates the count |

### `checklists.cy.ts` — 9 tests *(Sprint 3 — NEW — Issue #76)*
| Test |
|------|
| shows the board empty checklist state when no subtasks exist |
| keeps the board Add button disabled until a checklist title is entered |
| adds a board checklist item when the Add button is clicked |
| updates the board task progress bar fill after a checklist item is checked |
| shows the board delete confirmation with the selected checklist title |
| shows the planner empty checklist state when no subtasks exist |
| keeps the planner Add button disabled until a checklist title is entered |
| adds a planner checklist item from the Add button |
| deletes a planner checklist item from the task modal |

### `activity-history.cy.ts` — 7 tests *(Sprint 3 — NEW — Issue #77)*
| Test |
|------|
| opens the Activity tab and shows the project activity heading |
| renders activity icons and relative timestamps for entries |
| lists project members in the activity member filter dropdown |
| updates the activity list when the member filter changes |
| clears the date and member filters when Reset is clicked |
| loads more activity entries when Load more is clicked |
| shows the empty state when the project has no activity |

---

### `board.cy.ts` — 13 tests *(Sprint 3 — NEW — Issues #75, #76, #77)*
| Test |
|------|
| shows checklist items in order and renders progress on the task card |
| adds a checklist item from the task detail modal |
| toggles a checklist item without refreshing and updates progress immediately |
| deletes a checklist item and shrinks the progress summary |
| keeps a checklist item when delete is cancelled |
| keeps checklist changes visible after switching to planner |
| posts a comment from task details |
| edits an existing comment |
| cancels comment deletion from the confirm modal |
| deletes a comment after confirming |
| shows the Activity tab only for the owner |
| renders newest activity first, filters by member/date, and paginates with load more |
| shows an empty state when no activity has been logged |

### `planner.cy.ts` — 1 test *(Sprint 3 — NEW — Issue #76)*
| Test |
|------|
| shows and updates checklist items inside the planner task modal |

---

## Backend Unit Tests (Go) — Sprint 3 New Tests

All backend tests are in `backend/internal/testcases/` and run with:
```bash
cd backend
go test -v ./internal/testcases/...
```

See [Sprint2.md](Sprint2.md) for the full list of Sprint 2 backend test cases.

### `comment_test.go` — 17 tests *(Sprint 3 — NEW — Issue #68)*

| Test | Function under test |
|------|---------------------|
| `TestCommentService_CreateComment_Success` | `CreateComment()` happy path |
| `TestCommentService_CreateComment_EmptyContentRejected` | `ErrCommentContentRequired` |
| `TestCommentService_CreateComment_InvalidTaskRejected` | task ownership check |
| `TestCommentService_GetCommentsByTask_ReturnsAll` | `GetCommentsByTask()` ordering |
| `TestCommentService_GetCommentsByTask_EmptyWhenNone` | empty list result |
| `TestCommentService_UpdateComment_Success` | `UpdateComment()` happy path |
| `TestCommentService_UpdateComment_EmptyContentRejected` | `ErrCommentContentRequired` on update |
| `TestCommentService_UpdateComment_OtherUserCannotEdit` | `ErrCommentNotFoundOrAccessDenied` |
| `TestCommentService_DeleteComment_Success` | `DeleteComment()` removes entry |
| `TestCommentService_DeleteComment_OtherUserCannotDelete` | access control on delete |
| `TestCommentService_ContentTrimmed` | whitespace normalization |
| `TestCommentController_CreateComment_Unauthorized` | 401 without user context |
| `TestCommentController_GetCommentsByTask_Unauthorized` | 401 without user context |
| `TestCommentController_UpdateComment_Unauthorized` | 401 without user context |
| `TestCommentController_DeleteComment_Unauthorized` | 401 without user context |
| `TestCommentController_CreateComment_EmptyContentReturnsBadRequest` | 400 for blank content |
| `TestCommentController_CRUD_RoundTrip` | Create → List → Update → Delete → Verify |
| `TestNewCommentController` | constructor returns non-nil |

### `subtask_test.go` — 20 tests *(Sprint 3 — NEW — Issue #69)*

| Test | Function under test |
|------|---------------------|
| `TestSubtaskService_CreateSubtask_Success` | `CreateSubtask()` happy path |
| `TestSubtaskService_CreateSubtask_EmptyTitleRejected` | `ErrSubtaskTitleRequired` |
| `TestSubtaskService_CreateSubtask_InvalidTaskRejected` | task ownership check |
| `TestSubtaskService_GetSubtasksByTask_OrderedByPosition` | ordering by position ASC |
| `TestSubtaskService_GetSubtasksByTask_EmptyWhenNone` | empty list result |
| `TestSubtaskService_UpdateSubtask_Title` | title update |
| `TestSubtaskService_UpdateSubtask_ToggleCompletion` | is_completed toggle |
| `TestSubtaskService_UpdateSubtask_EmptyTitleRejected` | `ErrSubtaskTitleRequired` on update |
| `TestSubtaskService_UpdateSubtask_NotFoundReturnsError` | `ErrSubtaskNotFoundOrAccessDenied` |
| `TestSubtaskService_DeleteSubtask_Success` | delete + position compaction |
| `TestSubtaskService_DeleteSubtask_NotFoundReturnsError` | `ErrSubtaskNotFoundOrAccessDenied` |
| `TestSubtaskService_Reorder_MovingSubtask` | position reordering logic |
| `TestSubtaskService_TitleTrimmed` | whitespace normalization |
| `TestSubtaskController_CreateSubtask_Unauthorized` | 401 without user context |
| `TestSubtaskController_GetSubtasksByTask_Unauthorized` | 401 without user context |
| `TestSubtaskController_UpdateSubtask_Unauthorized` | 401 without user context |
| `TestSubtaskController_DeleteSubtask_Unauthorized` | 401 without user context |
| `TestSubtaskController_CreateSubtask_EmptyTitleReturnsBadRequest` | 400 for blank title |
| `TestSubtaskController_CRUD_RoundTrip` | Create → List → Update → Delete → Verify |
| `TestNewSubtaskController` | constructor returns non-nil |

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
| POST | `/api/projects/:id/stages/:stageId/tasks` | Create a task with title, description, position, deadline, priority, and assigned user |
| GET | `/api/tasks/:id` | Get task by ID |
| PUT | `/api/tasks/:id` | Update task title, description, position, deadline, priority, and assigned user |
| DELETE | `/api/tasks/:id` | Delete task |
| PUT | `/api/tasks/:id/move` | Move a task to another stage / position |

### Comment Endpoints (JWT required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/tasks/:id/comments` | Create a comment for a task |
| GET | `/api/tasks/:id/comments` | List comments for a task |
| PATCH | `/api/comments/:id` | Update an existing comment |
| DELETE | `/api/comments/:id` | Delete a comment |

### Subtask Endpoints (JWT required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/tasks/:id/subtasks` | Create a subtask / checklist item |
| GET | `/api/tasks/:id/subtasks` | List subtasks for a task |
| PATCH | `/api/subtasks/:id` | Update subtask title, completion, or position |
| DELETE | `/api/subtasks/:id` | Delete a subtask / checklist item |

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

## Commit History — Nandhan (`nandhan/backend`)

Key backend work merged for Sprint 3:

- Task enhancements API for deadline, priority, and assignee support
- Task comments schema and CRUD API
- Subtasks / checklist schema, ordering logic, and CRUD API
- Merge-conflict resolution in `backend/internal/database/database.go` to preserve both Sprint 3 backend feature sets

---

## Additional Work Completed — (User's Contributions)

### 1. Label System Implementation

Backend implementation of a complete Labels/Tags system for tasks within projects:

- **Database Schema**: Created `labels` table (project-scoped with name, color, created_by, created_at) and `task_labels` join table with proper foreign keys and indexes
- **Models**: Added `Label`, `TaskLabel`, `LabelResponse`, and `TaskLabelResponse` structs to `models.go`
- **Repository Layer**: Implemented `LabelRepository` with CRUD operations for labels and `TaskLabelRepository` for task-label assignments
- **Service Layer**: Created `LabelService` and `TaskLabelService` with business logic, validation, and activity logging
- **Controller & Routes**: Added `LabelController` and `TaskLabelController` with endpoints for managing labels and assigning/removing labels from tasks
- **Activity Logging**: Integrated with `ActivityService` to log `label_created`, `label_deleted`, `label_assigned`, and `label_removed` events

**Reference**: [`plans/label_system_implementation.md`](plans/label_system_implementation.md)

---

### 2. Notifications System Implementation

Backend notifications system providing real-time and email notifications for important project events:

- **Database Schema**: Created `notifications` table with user_id, type, message, is_read, related_entity_type, related_entity_id, and created_at fields with optimized indexes
- **Repository**: Implemented `NotificationRepository` with CreateNotification, GetUserNotifications, MarkAsRead, MarkAllAsRead, GetUnreadCount, and duplicate prevention methods
- **Service**: Created `NotificationService` with NotifyMemberAdded, NotifyTaskAssigned, NotifyDeadlineNear methods plus paginated retrieval and read management
- **Controller**: Added `NotificationController` with GET `/api/notifications`, PATCH `/api/notifications/:id/read`, and PATCH `/api/notifications/read-all` endpoints
- **Background Job**: Implemented `StartDeadlineChecker()` with 15-minute interval for automatic deadline reminder notifications
- **Email Integration**: Non-blocking email notifications via async channel pattern

**Reference**: [`plans/notifications_system_implementation.md`](plans/notifications_system_implementation.md)

---

### 3. Activity History Async/Event-Based Extension

Extended the Activity History system to use an asynchronous, event-driven architecture:

- **Event Types**: Created `ActivityEvent` struct with ProjectID, UserID, UserName, Action, EntityType, EntityID, and Details fields
- **Event Bus**: Implemented buffered channel-based `EventBus` with configurable buffer size for non-blocking event publishing
- **Worker Pool**: Added concurrent worker pool pattern with WaitGroup for graceful shutdown
- **Event Handler**: Created `ActivityEventHandler` that processes events and delegates to `ActivityService`
- **Performance Benefits**: Eliminates latency impact on main requests, preserves database connection pool, prevents error propagation, and improves scalability under high load

**Reference**: [`Taskify/backend/docs/ACTIVITY_HISTORY_BONUS.md`](Taskify/backend/docs/ACTIVITY_HISTORY_BONUS.md)

---

### 4. Project Members MVP-Ready Improvements

Production-quality improvements to the Project Members feature:

- **Single Source of Truth**: `project_members` is the authoritative source for ownership; `projects.owner_id` is synced from it
- **Transaction Safety**: All critical operations (AddMember, RemoveMember) wrapped in transactions with `FOR UPDATE` locks
- **Middleware with Role Injection**: `ProjectAccessMiddleware` and `OwnerOnlyMiddleware` inject role into context for handlers
- **Pagination**: `GetMembers` supports `?page=1&limit=20` with total count in response
- **Standardized Errors**: Consistent JSON response format with `success`, `error`, and `code` fields
- **Audit Trail**: Automatic logging to `activity_logs` table for member_added and member_removed events
- **Permission Helpers**: Role-based `CanEditProject()` and `CanManageMembers()` functions
- **Proper Indexing**: Optimized indexes for membership checks, user queries, and activity logs

**Reference**: [`Taskify/backend/docs/PROJECT_MEMBERS_BONUS.md`](Taskify/backend/docs/PROJECT_MEMBERS_BONUS.md)

---

### Test Cases Written for These Features

Comprehensive unit tests were created in the [`Taskify/backend/internal/testcases/`](Taskify/backend/internal/testcases/) directory to verify the implementation of all 4 features:

#### 1. Label System Tests — [`label_service_test.go`](Taskify/backend/internal/testcases/label_service_test.go) (18 test cases)

| Test | Description |
|------|-------------|
| `TestNewLabelService` | Verifies LabelService initialization |
| `TestLabelRepository_CreateLabel_Success` | Tests successful label creation |
| `TestLabelRepository_CreateLabel_DefaultColor` | Tests default color assignment |
| `TestLabelRepository_CreateLabel_DuplicateName` | Tests duplicate name prevention |
| `TestLabelRepository_GetLabelByID` | Tests fetching a single label |
| `TestLabelRepository_GetLabelsByProject` | Tests fetching all project labels |
| `TestLabelRepository_DeleteLabel` | Tests label deletion |
| `TestTaskLabelRepository` | Tests task-label assignment and removal |
| `TestTaskLabelRepository_GetTaskIDsByLabel` | Tests finding tasks by label |
| `TestLabelService_NoProjectAccess` | Tests access denial for non-members |

#### 2. Notifications System Tests — [`notification_service_test.go`](Taskify/backend/internal/testcases/notification_service_test.go) (18 test cases)

| Test | Description |
|------|-------------|
| `TestNewNotificationService` | Verifies NotificationService initialization |
| `TestNotificationRepository_CreateNotification` | Tests notification creation |
| `TestNotificationRepository_GetUserNotifications` | Tests paginated notification retrieval |
| `TestNotificationRepository_MarkAsRead` | Tests marking single notification as read |
| `TestNotificationRepository_MarkAsRead_WrongUser` | Tests access control for mark as read |
| `TestNotificationRepository_MarkAllAsRead` | Tests bulk mark as read |
| `TestNotificationRepository_GetUnreadCount` | Tests unread count tracking |
| `TestNotificationRepository_CheckDuplicateDeadlineNotification` | Tests duplicate prevention |
| `TestNotificationService_NotifyMemberAdded_SelfNotification` | Tests self-notification prevention |
| `TestNotificationService_NotifyTaskAssigned_SelfNotification` | Tests self-notification prevention |

#### 3. Activity History Tests — [`activity_service_test.go`](Taskify/backend/internal/testcases/activity_service_test.go) (14 test cases)

| Test | Description |
|------|-------------|
| `TestNewActivityService` | Verifies ActivityService initialization |
| `TestActivityRepository_CreateActivityLog` | Tests activity log creation |
| `TestActivityRepository_GetActivityLogsByProject` | Tests project activity retrieval |
| `TestActivityRepository_GetActivityLogsByProject_Pagination` | Tests pagination |
| `TestActivityRepository_GetActivityLogsByProject_DateFilter` | Tests date range filtering |
| `TestActivityService_LogActivity` | Tests central activity logging |
| `TestActivityService_LogLabelCreated/Deleted/Assigned/Removed` | Tests label activity logging |
| `TestActivityRepository_ConcurrentLogging` | Tests concurrent/event-based logging |

#### 4. Project Members Tests — [`project_member_service_test.go`](Taskify/backend/internal/testcases/project_member_service_test.go) (18 test cases)

| Test | Description |
|------|-------------|
| `TestNewProjectMemberServicePM` | Verifies ProjectMemberService initialization |
| `TestProjectMemberRepository_AddMember` | Tests member addition |
| `TestProjectMemberRepository_AddMember_Duplicate` | Tests duplicate prevention |
| `TestProjectMemberRepository_RemoveMember` | Tests member removal |
| `TestProjectMemberRepository_RemoveMember_CannotRemoveOwner` | Tests owner protection |
| `TestProjectMemberRepository_IsOwner` | Tests ownership check |
| `TestProjectMemberRepository_IsMember` | Tests membership check |
| `TestProjectMemberRepository_GetMembersPaginated` | Tests paginated member retrieval |
| `TestProjectMemberService_CreateInvite` | Tests invite creation |
| `TestProjectMemberService_AcceptInviteByID` | Tests invite acceptance |
| `TestProjectMember_CanEditProject/CanManageMembers` | Tests permission helpers |
| `TestInvite_Expiry` | Tests invite expiration |



