 
# TA Connect Sprint 2
## Video Link
* **Front-End:** ...
* **Back-End:** https://youtu.be/8WT-iecgn4k

## Detail work you've completed in Sprint 2

### Front-End
* Student Dasboard: Implemented Join Queue button functionality
* Student Dasboard: Implemented real-time position display when in queue
* Student Dasboard: Implemented Cancel Spot button functionality
* TA Dasboard: Implemented display of Real-Time Queue on dashboard
* TA Dasboard: Implemented Serve Next Button functionality

### Back-End
* Created Queue and QueueEntry data models
* Implemented real-time queue updates
* Implemented serve next student API
* Implemented get queue state API
* Implemented leave queue API
* Implemented join queue API
<p>&nbsp;</p>

## List unit tests and Cypress tests for frontend
### Cypress
- Visit TA Connect Home Page
  - **Visits the TA Connect home page:** Verifies that the app loads successfully at the root URL

- Login Page
  - **Displays the login form:** Verifies that the email input, password input, and login button are visible
  - **Shows an error for invalid credentials:** Verifies that an error message appears when incorrect credentials are submitted
  - **Has a link to the register page:** Verifies that clicking the Register link navigates to /register

- Register Page
  - **Displays the registration form:** Verifies that the name, email, password inputs and role dropdown are visible
  - **Can select TA role:** Verifies that the role dropdown can be changed to TA
  - **Defaults to student role:** Verifies that the role dropdown defaults to Student
  - **Has a link back to the login page:** Verifies that clicking the Login link in the footer navigates back to /login

- Student Dashboard
  - **Displays the student dashboard with welcome message:** Verifies that correct student dashboard is visible after login
  - **Displays all four course options in the dropdown:** Verifies that the course selection dropdown contains all available course options
  - **Displays all five days in the weekly schedule:** Verifies that Monday through Friday are all visible in the weekly schedule
  - **Displays the notification icon in the navbar:** Verifies that the notification icon is visible in the navigation bar
  - **Displays TA names in Today's TA Hours:** Verifies that TA names are visible in the Today's TA Hours section

- TA Dashboard
  - **Displays the TA dashboard with welcome message:** Verifies that correct TA dashboard is visible after login
  - **Displays the correct navigation tabs:** Verifies that Dashboard, My Office Hours, and Queue tabs are visible
  - **Displays today's office hour schedule cards:** Verifies that both office hour time slots are visible on the dashboard
  - **Enables the start queue button after selecting an office hour:** Verifies that selecting the first time slot enables the Start Queue button
  - **Enables the start queue button when the second office hour is selected:** Verifies that selecting the second time slot also enables the Start Queue button
  - **Displays all stat cards on the dashboard:** Verifies that all six stat cards are visible on the dashboard
  - **Displays the weekly office hours sidebar:** Verifies that the sidebar shows available office hours
  - **Shows the live queue view after starting the queue:** Verifies that clicking Start Queue loads the Live Queue view
  - **Displays queue controls in the live queue view:** Verifies that the Pause Queue and Close Queue buttons are visible
  - **Displays queue stats in the live queue view:** Verifies that Total in Queue, Longest Wait, and Average Wait Time stats are visible
  - **Toggles to Resume Queue after clicking Pause Queue:** Verifies that the queue status toggles to paused when Pause Queue is clicked
  - **Toggles back to Pause Queue after clicking Resume Queue:** Verifies that the queue status toggles back to open when Resume Queue is clicked
  - **Returns to the dashboard view after closing the queue:** Verifies that clicking Close Queue returns to the TA dashboard

### Unit Tests (Vitest + React Testing Library)

- Login Page
  - **Renders login form fields and submit button:** Verifies that the email input, password input, and login button are rendered
  - **Shows registration-success message when redirected from register:** Verifies that the success message appears when navigated from the register page
  - **Logs in a student and navigates to student dashboard:** Verifies that a successful student login calls the auth login and navigates to /student
  - **Logs in a TA and navigates to TA dashboard:** Verifies that a successful TA login calls the auth login and navigates to /ta
  - **Shows login error when API call fails:** Verifies that an error message is shown when the login API call fails

- Register Page
  - **Renders registration form with default student role:** Verifies that all form fields are rendered properly and the role defaults to Student
  - **Registers successfully and redirects to login with register flag:** Verifies that a successful registration shows a success message and redirects to /login
  - **Shows specific backend registration error message:** Verifies that a backend error message is displayed when registration fails with a known error
  - **Shows fallback registration error message:** Verifies that an error message is shown when registration fails

- Student Dashboard Page
  - **Renders student greeting and queue controls:** Verifies that the welcome message, course dropdown, and Join Queue button are rendered
  - **Shows error when no active queue exists for selected office hour:** Verifies that an error is shown when the TA has not opened the queue yet
  - **Joins queue, disables selectors, and shows real-time section:** Verifies that joining the queue disables the dropdown and shows the Real-Time Queue Status section
  - **Leaves queue and hides real-time section:** Verifies that leaving the queue hides the Real-Time Queue Status section and re-enables controls
  - **Auto-exits queue view when queue refresh no longer contains current student:** Verifies that the student is automatically removed from the queue view when they are no longer in the queue

- TA Dashboard Page
  - **Renders closed dashboard and requires selecting office-hour time first:** Verifies that the start queue button is disabled before an office hour is selected
  - **Renders welcome message with the TA username:** Verifies that the welcome message displays the TA's username
  - **Renders all closed-state stat cards:** Verifies that all six stat cards are visible on the dashboard
  - **Enables start button after selecting an office-hour card:** Verifies that clicking an office hour card enables the Start Queue button
  - **Opens live queue and calls backend queue creation flow:** Verifies that starting the queue calls the backend API and shows the Live Queue view
  - **Pauses and resumes the queue through backend status API:** Verifies that pausing and resuming the queue calls the backend status update API correctly
  - **Closes queue and returns to closed dashboard state:** Verifies that closing the queue calls the backend and returns to the dashboard view
  - **Opens announcement modal from dashboard and sends announcement:** Verifies that the announcement modal opens, accepts input, and sends the announcement
  - **Calls next endpoint when starting session on first queued student:** Verifies that clicking Start Session calls the nextQueueStudent API with the correct queue ID

### Backend Go tests (`backend/api_e2e_test.go`)

- **`TestRegisterAndLoginStudentAndTA`**
  - **Register student**: `POST /api/register` → expects **200 OK** and user role `student`
  - **Register TA**: `POST /api/register` → expects **200 OK** and user role `ta`
  - **Login student**: `POST /api/login` → expects **200 OK**
  - **Login TA**: `POST /api/login` → expects **200 OK**

- **`TestRegister_DuplicateUsernameOrEmail`**
  - **Duplicate email**: second `POST /api/register` with same email → expects **409 Conflict**
  - **Duplicate username**: second `POST /api/register` with same username → expects **409 Conflict**

- **`TestRegister_InvalidRole_Returns400`**
  - **Invalid role**: `POST /api/register` with role not in `{student, ta}` → expects **400 Bad Request**

- **`TestLogin_WrongPassword_Returns401`**
  - **Wrong password**: `POST /api/login` with incorrect password → expects **401 Unauthorized**

- **`TestProtectedEndpoints_MissingToken_Return401`**
  - **Create queue without token**: `POST /api/queues` → expects **401 Unauthorized**
  - **Join/Next without token**: calls `/join` and `/next` without `Authorization`

- **`TestQueueLifecycle_HappyPath`**
  - **Create queue (TA)**: `POST /api/queues` → expects **201 Created**
  - **Join queue (student)**: `POST /api/queues/{id}/join` → expects **201 Created**
  - **Get queue shows entry**: `GET /api/queues/{id}` → expects **200 OK** and `entries` length is **1**
  - **Serve next student (TA)**: `POST /api/queues/{id}/next` → expects **200 OK**
  - **Get queue shows empty**: `GET /api/queues/{id}` → expects `entries` length is **0**

- **`TestRoleEnforcement_StudentCannotCreateQueueOrNext`**
  - **Student cannot create queue**: `POST /api/queues` with student token → expects **403 Forbidden**
  - **Student cannot advance queue**: `POST /api/queues/{id}/next` with student token → expects **403 Forbidden**

- **`TestQueueJoinLeave_Errors`**
  - **Join non-existent queue**: `POST /api/queues/999999/join` → expects **404 Not Found**
  - **Leave non-existent queue**: `POST /api/queues/999999/leave` → expects **404 Not Found**

- **`TestQueueEdgeCases_DuplicateJoin_LeaveNotInQueue_NextEmpty`**
  - **Duplicate join**: joining the same queue twice → expects **409 Conflict**
  - **Leave when not in queue**: `POST /api/queues/{id}/leave` as a student not in the queue → expects **404 Not Found**
  - **Next on empty queue**: after serving the only student, calling `/next` again → expects **404 Not Found**

- **`TestSSE_EmitsStudentJoinedEvent`**
  - **SSE smoke test**: connects to `GET /api/queues/{id}/events`, then triggers a join
  - **Expectation**: receives `event: STUDENT_JOINED` on the SSE stream

## Add documentation for your backend API

The backend is an **HTTP API** in **Go** with **chi** and **PostgreSQL** (`backend/internal/routes/routes.go`, `backend/cmd/server/main.go`, `go.mod`). Most routes use **JSON** request/response bodies. Queues follow **resource-style** paths (`/api/queues`, `/api/queues/{id}`), while **auth** and **queue actions** use **POST** “command” paths (`/api/login`, `/api/register`, `/join`, `/leave`, `/next`) — common for web apps, but not a strict REST-only design. **Live updates** use **Server-Sent Events (SSE)** on **`GET /api/queues/{id}/events`** (`Content-Type: text/event-stream`), not a JSON response body.

### Run the server

1. Create a Postgres database (e.g. `officehours`).
2. In **`backend/`**, copy `.env.example` → `.env` and set the DB connection string and **`JWT_SECRET`**.
3. Run `go mod tidy` then `go run ./cmd/server`. The process listens on **`PORT`** (default **`8080`**). Smoke test: **`GET /health`** → `200` `{"status":"ok"}` if the DB is reachable.

### Conventions

- **Base URL:** `http://localhost:<PORT>` (replace `8080` if you set `PORT`).
- **Bodies / errors:** Most endpoints use JSON. **SSE** (`/events`) uses **`text/event-stream`**, not JSON for the overall response. Wrong HTTP method on a handler → **`405`**. Failures usually look like `{"error":"<message>"}`; many DB failures return **`500`** with `"database error"`.
- **CORS:** `Access-Control-Allow-Methods: GET, POST, OPTIONS`; headers allowed include `Authorization` (see `corsMiddleware` in `routes.go`).

### Authentication

After **`POST /api/register`** or **`POST /api/login`**, responses include a **`token`** string. Send it on protected routes as:

`Authorization: Bearer <token>`

The JWT carries **`user_id`**, **`email`**, and **`role`** (`student` or `ta`) — see `backend/internal/auth/jwt.go`. Missing token, malformed header, or invalid/expired token → **`401`** with `{"error":"authorization required"}`.

**Routes that require a valid Bearer token:**

| Route | Extra rule (from code) |
|--------|-------------------------|
| `POST /api/queues` | Caller must have **`role: ta`** (`403` otherwise). |
| `POST /api/queues/{id}/join` | Any authenticated user. |
| `POST /api/queues/{id}/leave` | Any authenticated user. |
| `POST /api/queues/{id}/next` | **`role: ta`** and JWT **`user_id`** must match the queue’s **`ta_id`** (`403` otherwise). |

All other routes listed below do **not** require a token.

### Endpoints

Path parameter **`{id}`** is the numeric queue id (`chi` route `/api/queues/{id}`). If **`{id}`** is not a valid integer, **`GET /api/queues/{id}`**, **join**, **leave**, **next**, and **events** return **`400`** `{"error":"invalid queue id"}`.

| Method | Path | What it does | Success |
|--------|------|----------------|---------|
| `GET` | `/health` | Ping database | **`200`** `{"status":"ok"}` · **`503`** if ping fails (`{"status":"unavailable","error":"database"}`) |
| `POST` | `/api/register` | Create user | **`200`** JSON with `token` and `user` `{ id, username, email, role }`. **`400`** validation / bad role · **`409`** duplicate username or email |
| `POST` | `/api/login` | Login | **`200`** same shape as register · **`401`** bad email/password |
| `POST` | `/api/queues` | TA starts a queue | **`201`** `{ id, course_id, ta_id, status, created_at }` — **`status`** is **`open`**. Optional body: `{ "course_id": 0 }`; omitting the body ⇒ **`course_id` is 0** |
| `GET` | `/api/queues/{id}` | Queue + waiting students | **`200`** `{ id, course_id, ta_id, status, created_at, entries: [...] }`. Each entry: `id`, `queue_id`, `student_id`, `position`, `joined_at`, `username`. Ordered by **position**, then **joined_at**. **`404`** unknown queue |
| `POST` | `/api/queues/{id}/join` | Authenticated user joins | **`201`** `{ id, queue_id, position, joined_at }`. Queue must exist and **`status`** must be **`open`**. Duplicate student in same queue → **`409`**. Triggers SSE (below) |
| `POST` | `/api/queues/{id}/leave` | Authenticated user leaves | **`204`** empty body; positions renumbered. **`404`** if no row was removed (`not in queue`) |
| `POST` | `/api/queues/{id}/next` | Owning TA removes front of line | **`200`** `{ "queue_id", "status": "in_session", "student": { …entry } }`. **`404`** no queue or empty queue · **`409`** queue not **`open`**. Triggers SSE |

**Queue `status` in the database** can be `open`, `paused`, or `closed` (`migrate.go`). **Current API behavior:** **`POST /api/queues`** only creates **`open`** queues; **join** and **next** require **`open`**. There is no HTTP route in this repo to set `paused` / `closed` yet.

### Real-time updates (SSE)

**`GET /api/queues/{id}/events`** — **Server-Sent Events**, no auth.

1. Response headers include `Content-Type: text/event-stream`.
2. First line is a comment: `: connected` (keeps the stream open).
3. Later lines look like: `event: <TYPE>` then `data: <JSON>` (blank line between events).

The **`data`** line is **only** the JSON for the event **`payload`** field (see `StreamQueueEvents` in `backend/internal/queue/events.go`). **`QUEUE_UPDATED`** is sent with **no payload** ⇒ **`data: null`**.

**`event` names:** `STUDENT_JOINED` · `STUDENT_LEFT` · `STUDENT_SERVED` · `QUEUE_UPDATED`.

Join / leave / **next** each publish the corresponding events after the database change succeeds.

### Example: register and login (curl)

```bash
curl -s -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","email":"alice@example.com","password":"secret123","role":"student"}'

curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"secret123"}'
```

Use the `"token"` from the response on a protected call (example: TA creates a queue):

```bash
curl -s -X POST http://localhost:8080/api/queues \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{"course_id":0}'
```

### Tests

From **`backend/`** with Postgres and `.env` configured: **`go test -v .`** (see **`api_e2e_test.go`**).
