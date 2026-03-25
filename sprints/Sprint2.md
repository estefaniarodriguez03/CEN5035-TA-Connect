 
# TA Connect Sprint 2
## Video Link
* **Front-End:** ...
* **Back-End:** https://youtu.be/8WT-iecgn4k

## Detail work you've completed in Sprint 2

### Front-End
* ...

### Back-End
* Created Queue and QueueEntry data models
* Implemented real-time queue updates
* Implemented serve next student API
* Implemented get queue state API
* Implemented leave queue API
* Implemented join queue API
<p>&nbsp;</p>

## List unit tests and Cypress tests for frontend
### Cypress Tests
* Visit TA Connect Home Page: Verifies that the app loads successfully at the root URL
* Login Page:
  * Displays the login form: Verifies that the email input, password input, and login button are visible
  * Shows an error for invalid credentials: Verifies that an error message appears when wrong credentials are submitted
  * Has a link to the register page: Verifies that clicking the Register link navigates to /register
* Register Page
  * Displays the registration form: Verifies that the name, email, password inputs and role dropdown are visible
  * Can select TA role: Verifies that the role dropdown can be changed to TA
  * Defaults to student role: Verifies that the role dropdown defaults to Student
  * Has a link back to the login page: Verifies that clicking the Login link navigates back to /login
* Student Dashboard
  * Displays the student dashboard: Verifies that the student dashboard loads
* TA Dashboard
  * Displays the TA dashboard: Verifies that the TA dashboard loads
  * Navigates to the queue when Start Office Hours Live Queue is clicked: Verifies that clicking the Start Office Hours Live Queue button loads the Queue Management page

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
