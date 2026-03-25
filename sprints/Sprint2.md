 
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

### Unit Tests
* ...

## List unit tests for backend
* ...

## Add documentation for your backend API 
Go API server with PostgreSQL.

### Prerequisites

**Go** 1.21
**PostgreSQL** 18.1

#### 1. Install PostgreSQL

Download Installer online on PostgresSQL website

### 2. Create a database

PostgreSQL uses a default superuser `postgres`. Create a database for the app:

```bash
psql -U postgres

# Inside psql:
CREATE DATABASE officehours;
\q
```

Use a different name if you want it is set in `.env`

### 3. Configure connection

Copy the example env file and set your database credentials:

```bash
cp .env.example .env
```

Values should be set as needed.

If you created a different database use that values instead.

### 4. Run the server

From the `backend` directory:

First run this for dependecies:

```bash
go mod tidy
```

Then run this to run the database server:

```bash
go run ./cmd/server
```

You should see this:

```bash
server listening on :8080
```

### Test PostgreSQL connection

`localhost:8080/health` pings the database

## 5. Auth (login / register)

Set `JWT_SECRET` in your `.env` (see `.env.example`). The server uses it to sign JWTs.

#### Test with curl (macOS / Linux)

**Register** (creates a user; role is `student` or `ta`):

```bash
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","email":"alice@example.com","password":"secret123","role":"student"}'
```

**Login**:

```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"secret123"}'
```

Both return JSON with `token` (JWT) and `user` (id, username, email, role). Use the token in the `Authorization: Bearer <token>` header for protected endpoints later.

#### Windows (PowerShell)

**Health**:

```powershell
curl.exe "http://localhost:8080/health"
```

**Register** (`curl.exe`):

```powershell
curl.exe -X POST "http://localhost:8080/api/register" `
  -H "Content-Type: application/json" `
  -d "{\"username\":\"alice\",\"email\":\"alice@example.com\",\"password\":\"secret123\",\"role\":\"student\"}"
```

**Login** (`curl.exe`):

```powershell
curl.exe -X POST "http://localhost:8080/api/login" `
  -H "Content-Type: application/json" `
  -d "{\"email\":\"alice@example.com\",\"password\":\"secret123\"}"
```

**Register / Login** (native PowerShell):

```powershell
$registerBody = @{
  username = "alice"
  email    = "alice@example.com"
  password = "secret123"
  role     = "student"
} | ConvertTo-Json

Invoke-RestMethod -Method POST `
  -Uri "http://localhost:8080/api/register" `
  -ContentType "application/json" `
  -Body $registerBody

$loginBody = @{
  email    = "alice@example.com"
  password = "secret123"
} | ConvertTo-Json

$loginResponse = Invoke-RestMethod -Method POST `
  -Uri "http://localhost:8080/api/login" `
  -ContentType "application/json" `
  -Body $loginBody

$TOKEN = $loginResponse.token
```

Then use the token for protected endpoints:

```powershell
$headers = @{ Authorization = "Bearer $TOKEN" }
# Example:
# Invoke-RestMethod -Method POST -Uri "http://localhost:8080/api/queues/2/next" -Headers $headers
```
