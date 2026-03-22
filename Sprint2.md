# Sprint 2 — Frontend Development

**Branch:** `sreeja/frontendDev`

---

## Work Completed

### 1. Dark / Light Theming System
- Introduced `ThemeService` with `localStorage` persistence — theme survives page refresh
- Defined a full set of CSS custom properties (`:root` for dark, `[data-theme="light"]` for light) in `styles.scss`, covering backgrounds, text, borders, cards, inputs, modals, dropdowns, avatars, and nav elements
- Added a theme-toggle button (sun/moon icon) to every page navbar
- Rewrote all component SCSS files to consume CSS variables instead of hardcoded colours:
  - `welcome.component.scss`
  - `features.component.scss`
  - `about.component.scss`
  - `login.component.scss`
  - `signup.component.scss`
  - `board.component.scss`
  - `home.component.scss`

### 2. Board Filter Panel
- Added a sticky filter panel below the board topbar
- Filter by **Priority**: All / Low / Medium / High / Critical
- Filter by **Due date**: Any / Overdue / Today / This week / No date
- Active-filter badge (purple dot) on the Filter button when a filter is applied
- "Clear filters" button resets all active filters
- Fixed a bug where the `'none'` due-date filter incorrectly passed through tasks that have a due date

### 3. Board Share Modal
- Share button opens a modal with the current board URL pre-filled
- One-click **Copy** button uses the Clipboard API; button label changes to "Copied!" for 2.5 seconds

### 4. Board Switcher (Board Selector Dropdown)
- Board title in the topbar is now a clickable dropdown listing all boards owned by the logged-in user
- Active board is highlighted (indigo tint + checkmark)
- Selecting a different board navigates to `/board/:id` via Angular Router — no page reload
- Switched from `route.snapshot` to `route.params` observable so the component reacts to route changes without being destroyed and recreated
- Click-outside backdrop dismisses the dropdown

### 5. Light-Mode Visibility Fixes
- Fixed invisible Login button text in the welcome page navbar in light mode (was white text on white background)
- Fixed brand name and logo link colour in light mode
- Used `:host-context([data-theme="light"])` for Angular-encapsulated component overrides where `[data-theme="light"] .selector` was not being applied reliably

---

## Testing

### How to run unit tests
```bash
cd frontend
npm test               # runs Karma/Jasmine in watch mode
npm test -- --watch=false --browsers=ChromeHeadless   # single headless run
```

### How to run Cypress E2E tests
```bash
cd frontend
# Terminal 1 — start the app
npm start

# Terminal 2 — open Cypress interactively
npm run cy:open

# Or run headlessly
npm run cy:run
```

---

## Unit Tests (Karma / Jasmine) — 45 total

### `ThemeService` — 8 tests
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

### `AuthService` — 12 tests
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

### `BoardComponent` — 22 tests
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
| getFilteredTasks() should filter by "none" due date | `getFilteredTasks()` |
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

### `AppComponent` — 3 tests
| Test | What it checks |
|------|----------------|
| should create the app | component instantiation |
| should have the 'taskify' title | `title` property |
| should render a router-outlet | template structure |

---

## Cypress E2E Tests — 19 total

### Welcome page (`cypress/e2e/welcome.cy.ts`) — 6 tests
| Test |
|------|
| should load the welcome page and display the brand name |
| should show the Login and Sign up buttons in the navbar |
| should navigate to /login when Login button is clicked |
| should navigate to /signup when Sign up button is clicked |
| should display the hero heading |
| should toggle the theme when the theme toggle button is clicked |

### Login page (`cypress/e2e/login.cy.ts`) — 6 tests
| Test |
|------|
| should display the login form |
| should fill in the email field |
| should fill in the password field |
| should show an error when submitting empty credentials |
| should navigate to signup page via the sign up link |
| should toggle password visibility |

### Signup page (`cypress/e2e/signup.cy.ts`) — 7 tests
| Test |
|------|
| should display the signup form |
| should fill in the name field |
| should fill in the email field |
| should fill in the full signup form |
| should show error when passwords do not match |
| should navigate to login page via the log in link |
| should have a back to home link |

---

## Backend Unit Tests (Go) — ~95 tests

All backend tests are located in `backend/internal/testcases/` and can be run with:
```bash
cd backend
go test -v ./internal/testcases/...
```

### `jwt_service_test.go` — 8 tests
| Test | Function under test |
|------|---------------------|
| valid token generation | `GenerateToken()` |
| empty user ID | `GenerateToken()` |
| empty email | `GenerateToken()` |
| valid token validation | `ValidateToken()` |
| invalid token format | `ValidateToken()` |
| empty token | `ValidateToken()` |
| token with wrong secret | `ValidateToken()` |
| malformed token | `ValidateToken()` |

### `otp_service_test.go` — 6 tests
| Test | Function under test |
|------|---------------------|
| generates 6-digit OTP | `GenerateOTP()` |
| generates unique OTPs | `GenerateOTP()` |
| valid OTP verification | `VerifyOTP()` |
| invalid OTP code | `VerifyOTP()` |
| non-existent email | `VerifyOTP()` |
| valid reset token | `ValidateResetToken()` |

### `auth_service_test.go` — 12 tests
| Test | Function under test |
|------|---------------------|
| valid email formats | `isValidEmail()` |
| invalid email formats | `isValidEmail()` |
| email normalization | `normalizeEmail()` |
| valid registration | validation |
| empty name | validation |
| empty email | validation |
| invalid email format | validation |
| password too short | validation |
| RegisterRequest structure | struct |
| LoginRequest structure | struct |
| AuthResponse structure | struct |
| AuthService errors | error definitions |

### `auth_controller_test.go` — 11 tests
| Test | Function under test |
|------|---------------------|
| returns 400 for invalid JSON (Register) | `Register()` |
| returns 400 for invalid JSON (Login) | `Login()` |
| returns 401 when no user ID in context (GetMe) | `GetMe()` |
| returns 400 for invalid JSON (ForgotPassword) | `ForgotPassword()` |
| returns 400 for empty email | `ForgotPassword()` |
| returns 400 for invalid JSON (VerifyOTP) | `VerifyOTP()` |
| returns 400 for missing fields | `VerifyOTP()` |
| returns 400 for invalid JSON (ResetPassword) | `ResetPassword()` |
| returns 400 for missing fields | `ResetPassword()` |
| creates new controller | `NewAuthController()` |
| AuthService errors | error definitions |

### `middleware_test.go` — 10 tests
| Test | Function under test |
|------|---------------------|
| returns user ID from context | `GetUserID()` |
| returns empty string when not set | `GetUserID()` |
| returns empty string for wrong type | `GetUserID()` |
| returns user email from context | `GetUserEmail()` |
| returns empty string when not set | `GetUserEmail()` |
| allows request with valid token | `JWTAuthMiddleware()` |
| rejects request without Authorization header | `JWTAuthMiddleware()` |
| rejects request with invalid token | `JWTAuthMiddleware()` |
| rejects request with wrong secret | `JWTAuthMiddleware()` |
| writes error response | `writeError()` |

### `models_test.go` — 13 tests
| Test | Function under test |
|------|---------------------|
| User.ToResponse() | user conversion |
| RoleUser constant | constants |
| RoleAdmin constant | constants |
| Project structure | struct |
| Stage structure | struct |
| Task structure | struct |
| Message structure | struct |
| zero Project | zero values |
| zero Stage | zero values |
| zero Task | zero values |
| zero Message | zero values |
| UserResponse fields | struct |
| All User fields | struct |

### `controllers_test.go` — 14 tests
| Test | Function under test |
|------|---------------------|
| CreateProject returns 401 without user | authorization |
| CreateProject returns 400 for invalid JSON | validation |
| GetAllProjects returns 401 without user | authorization |
| CreateStage returns 401 without user | authorization |
| GetStagesByProject returns 401 without user | authorization |
| GetStage returns 401 without user | authorization |
| CreateTask returns 401 without user | authorization |
| GetTasksByStage returns 401 without user | authorization |
| GetTask returns 401 without user | authorization |
| CreateMessage returns 401 without user | authorization |
| GetMessagesByProject returns 401 without user | authorization |
| DeleteMessage returns 401 without user | authorization |
| GetUserID helper returns user ID | helper |
| GetUserID helper returns empty without context | helper |

### `services_test.go` — 4 tests
| Test | Function under test |
|------|---------------------|
| NewProjectService returns non-nil | constructor |
| NewStageService returns non-nil | constructor |
| NewTaskService returns non-nil | constructor |
| NewMessageService returns non-nil | constructor |

### `user_repository_test.go` — 11 tests
| Test | Function under test |
|------|---------------------|
| NewUserRepository returns non-nil | constructor |
| CreateUser inserts user | `CreateUser()` |
| GetUserByEmail returns nil for non-existent | `GetUserByEmail()` |
| GetUserByEmail returns user for existing | `GetUserByEmail()` |
| GetUserByID returns nil for non-existent | `GetUserByID()` |
| GetUserByID returns user for existing | `GetUserByID()` |
| EmailExists returns false for non-existent | `EmailExists()` |
| EmailExists returns true for existing | `EmailExists()` |
| UpdatePassword succeeds | `UpdatePassword()` |
| UpdatePassword fails for non-existent email | `UpdatePassword()` |

### `database_test.go` — 6 tests
| Test | Function under test |
|------|---------------------|
| sql.Open creates connection | database |
| db.Ping succeeds | database |
| exec creates table | SQL |
| exec inserts row | SQL |
| query selects row | SQL |
| transaction support | transaction |
