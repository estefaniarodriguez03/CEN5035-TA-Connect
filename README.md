# CEN5035 Group Project: TA Connect

## Project description

TAConnect is a web platform that streamlines the office hours experience of university teaching assistants (TAs) and their students. The application provides a real time virtual queue, estimated wait times, and push notifications to reduce idle time for students. TAs can view, reorder, and launch office hour sessions while students can queue, and view TA schedules once logged in. TAConnect essentially offers an efficient office hours scheduling system that solves frequent problems with current scheduling methods.

## Requirements

| Component | Version / notes |
| --- | --- |
| **PostgreSQL** | Running instance and empty database (e.g. `officehours`) |
| **Go** | **1.21+** (`go.mod`) |
| **Node.js** | **18+** recommended (for Vite / React frontend) |

## Running the backend

1. Create a Postgres database.
2. Copy **`backend/.env.example`** → **`backend/.env`** and set **`DB_*`**, **`JWT_SECRET`**, and optionally Zoom vars.
3. From **`backend/`**:

   ```bash
   go mod tidy
   go run ./cmd/server
   ```

   The API listens on **`PORT`** (default **`8080`**). Migrations run on startup. Smoke check: **`GET http://localhost:8080/health`** → **`200`** when the DB is reachable.

## Running the frontend

1. From **`frontend/`**:

   ```bash
   npm install
   npm run dev
   ```

2. The dev server proxies API traffic to **`http://localhost:8080`** (see **`vite.config.ts`**). If the API runs elsewhere, set **`VITE_API_URL`** when building or in **`.env`** for the frontend.

## Using the application

1. Open the URL printed by Vite (typically **`http://localhost:5173`**).
2. **Register** as a **student** or **TA**, then **log in**.
3. **TA:** manage office hours, link courses, start/pause/close queues, serve students, post announcements.
4. **Student:** pick course/office-hour context, **join** an open queue, view wait-related UI and schedules.

Backend integration tests (needs Postgres + **`.env`**): from **`backend/`**, run **`go test -v ./tests`**.

## Project members

### Front-end engineers

1. Sara Waters  
2. Estefania Rodriguez Da Silva  

### Back-end engineers

1. Raghav Nanjappan  
2. John Spurrier  
