# Backend

Go API server with PostgreSQL.

## Prerequisites

**Go** 1.21
**PostgreSQL** 18.1

## 1. Install PostgreSQL

Download Installer online on PostgresSQL website

## 2. Create a database

PostgreSQL uses a default superuser `postgres`. Create a database for the app:

```bash
psql -U postgres

# Inside psql:
CREATE DATABASE officehours;
\q
```

Use a different name if you want it is set in `.env`

## 3. Configure connection

Copy the example env file and set your database credentials:

```bash
cp .env.example .env
```

Values should be set as needed.

If you created a different database use that values instead.

## 4. Run the server

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

```
server listening on :8080
```

### Test PostgreSQL connection

localhost:8080/health pings the database

## 5. Auth (login / register)

Set `JWT_SECRET` in your `.env` (see `.env.example`). The server uses it to sign JWTs.

### Test with curl

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
