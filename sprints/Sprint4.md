# TA Connect Sprint 4

## Video Link
* **Front-End + Project Overview:** https://youtu.be/u9qu_gYQq-Y
* **Back-End:** https://youtu.be/oxNVwSq2I7o

## Detail Work Completed in Sprint 4
### Front-End
- Updated all office hours and course data with live API data from backend
- Connected My Office Hours Page to backend to create, edit, and delete office hours
- Fixed real-time sync between My Office Hours Page and TA Dashboard
- Implemented student schedule with public courses endpoint
- Created profile page for TA to update name and add courses
- Created profile page for students to update name and add courses
- Fixed weekly schedule grid to correctly filter slots by day of week
- Fixed course color display
- Designed Zoom session popup

### Back-End
- Built API route for TA to add course
- Added TA-course relationship schema
- Created endpoint to start session with Zoom link
- Created API to fetch session history
- Created session Log Model
- Handled invalid state transitions
- Handled empty queue cases
- Standardized error response format
- Audited all endpoints for role checks
- Implemented global role middleware
- Included ETA in queue state API response
- Implemented ETA calculation utility
- Added average session duration field

## List Frontend Unit Tests
### Cypress - Sprint 2
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
 
### Cypress - Sprint 3
- TA Dashboard — My Office Hours Tab
  - **Navigates to My Office Hours tab when clicked:** Verifies that clicking the My Office Hours tab opens the office hours management view
  - **Displays the Add Office Hours button on the My Office Hours tab:** Verifies that the + Add Office Hours button is visible on the My Office Hours page
  - **Opens the Add Office Hours modal when button is clicked:** Verifies that clicking + Add Office Hours opens the modal with two time input fields
  - **Adds a new office hour and displays it in the schedule:** Verifies that a newly added office hour appears in the schedule
  - **Shows a success toast after adding office hours:** Verifies that a success notification appears after adding office hours
  - **Shows an error toast when required fields are missing:** Verifies that an error notification appears if the Add Schedule form is submitted incomplete
  - **Opens the edit modal pre-filled when Edit is clicked:** Verifies that clicking Edit opens the Edit Office Hours modal with existing values populated
  - **Updates an office hour and shows success toast:** Verifies that editing an office hour updates the schedule and shows a success notification
  - **Deletes an office hour and shows success toast:** Verifies that deleting an office hour removes it and shows a success notification
  - **Shows empty state for days with no office hours:** Verifies that a no office hours scheduled message appears when no office hours exist for a day
  - **Navigates back to Dashboard tab from My Office Hours:** Verifies that clicking Dashboard returns to the TA dashboard view

- Toast Notifications
  - **Shows toast notifications on the student dashboard:** Verifies that an error toast appears when a student tries to join a queue that has not been opened yet

- Student Dashboard — Queue Interaction
  - **Does not display a status banner for an open queue:** Verifies that no queue status banner is shown when the selected queue is open
  - **Does not show queue status banners while the student is in the queue:** Verifies that queue status banners are hidden after the student joins the queue
  - **Displays the Estimated Wait Time card:** Verifies that the Estimated Wait Time card is visible on the student dashboard
  - **Displays student count in the wait time card:** Verifies that the current number of students in queue is shown in the wait time card
  - **Navigates to the My Courses page when My Courses tab is clicked:** Verifies that clicking My Courses navigates to the My Courses page

### Cypress - Sprint 4
- Student Dashboard
  - **Displays the course dropdown:** Verifies that the course selection dropdown is visible on the student dashboard
  - **Displays the Estimated Wait Time card:** Verifies that the Estimated Wait Time card is visible on the student dashboard
  - **Displays the weekly schedule section:** Verifies that the This Week's Office Hours section is visible
  - **Navigates to the My Courses page when My Courses tab is clicked:** Verifies that clicking My Courses navigates to /student/my-courses
  - **Navigates to the profile page when the profile icon is clicked:** Verifies that clicking the profile icon navigates to the profile page

- Student Dashboard — Queue Interaction
  - **Shows helper text when no course is selected:** Verifies that the Select a course to see the queue message is displayed by default
  - **Join Queue button is disabled when no course is selected:** Verifies that the Join Queue button cannot be clicked until a course is selected

- Student — My Courses Page
  - **Displays the My Courses page header:** Verifies that the My Courses - Office Hours heading is visible
  - **Displays the Add Office Hours button:** Verifies that the + Add Office Hours button is visible
  - **Displays weekly and today view toggle buttons:** Verifies that the Weekly and Today toggle buttons are visible
  - **Displays filter dropdowns:** Verifies that the Filter by Course and Filter by TA dropdowns are visible
  - **Switches to Today view when clicked:** Verifies that clicking Today shows the Today's Office Hours section
  - **Opens the Add Office Hours modal when button is clicked:** Verifies that clicking + Add Office Hours opens the modal
  - **Shows course dropdown as first step in modal:** Verifies that the Course label and course dropdown are the first step in the modal
  - **Shows TA dropdown after selecting a course:** Verifies that selecting a course in the modal reveals the TA dropdown
  - **Add to Schedule button is disabled when no slot is selected:** Verifies that the Add to Schedule button cannot be clicked until a time slot is selected
  - **Navigates back to dashboard when Dashboard tab is clicked:** Verifies that clicking Dashboard returns to /student

- Student Profile Page
  - **Displays the student's name:** Verifies that the student's name is visible in the profile card
  - **Displays the student's email:** Verifies that the student's email address is visible
  - **Displays the Major field:** Verifies that the Major label is visible
  - **Displays the Year field:** Verifies that the Year label is visible
  - **Displays the Courses section:** Verifies that the Courses section is visible
  - **Displays the Edit Profile button:** Verifies that the Edit Profile button is visible
  - **Displays the Log Out button:** Verifies that the Log Out button is visible
  - **Displays the navbar with Dashboard and My Courses tabs:** Verifies that the Dashboard and My Courses navigation tabs are visible
  - **Navigates to the dashboard when Dashboard tab is clicked:** Verifies that clicking Dashboard navigates to /student
  - **Navigates to My Courses when My Courses tab is clicked:** Verifies that clicking My Courses navigates to /student/my-courses
  - **Logs out when Log Out is clicked:** Verifies that clicking Log Out navigates to the login page
  - **Enters edit mode when Edit Profile is clicked:** Verifies that clicking Edit Profile shows the Save and Cancel buttons
  - **Shows name input field in edit mode:** Verifies that a name text input is visible in edit mode
  - **Shows email as disabled in edit mode:** Verifies that the email field is read-only in edit mode
  - **Shows major input field in edit mode:** Verifies that a major text input is visible in edit mode
  - **Shows year dropdown in edit mode:** Verifies that the year dropdown is visible in edit mode
  - **Year dropdown contains all expected options:** Verifies that the year dropdown contains Freshman, Sophomore, Junior, Senior, Graduate, and PhD
  - **Shows the Add Course button in edit mode:** Verifies that the + Add Course button is visible in edit mode
  - **Cancels edit mode when Cancel is clicked:** Verifies that clicking Cancel returns to view mode and hides the Save button
  - **Restores original name when Cancel is clicked after editing:** Verifies that the original name is restored when edits are cancelled
  - **Updates the name successfully and shows success toast:** Verifies that saving a new name shows a Profile updated successfully! notification
  - **Can select a different year in edit mode:** Verifies that the year dropdown can be changed to Senior
  - **Can update the major field:** Verifies that the major input accepts new text
  - **Opens the Add Course modal when + Add Course is clicked:** Verifies that clicking + Add Course opens the Add Course modal
  - **Displays the course selection dropdown in the Add Course modal:** Verifies that the course dropdown is visible in the modal
  - **Closes the Add Course modal when X is clicked:** Verifies that clicking the X button closes the modal
  - **Adds a course from the modal and shows it in the courses list:** Verifies that selecting and confirming a course adds it to the profile courses list
  - **Does not show already enrolled courses in the Add Course modal dropdown:** Verifies that already enrolled courses do not appear as options in the modal
  - **Can remove a course in edit mode:** Verifies that clicking Remove on a course reduces the course count

- TA Profile Page
  - **Displays the profile page header:** Verifies that the Welcome to your Profile! heading is visible
  - **Displays the user's email:** Verifies that the TA's email address is visible
  - **Displays the Edit Profile button:** Verifies that the Edit Profile button is visible
  - **Displays the Log Out button:** Verifies that the Log Out button is visible
  - **Enters edit mode when Edit Profile is clicked:** Verifies that clicking Edit Profile shows the Save and Cancel buttons
  - **Shows a name input field in edit mode:** Verifies that a name input field is visible in edit mode
  - **Shows email as disabled in edit mode:** Verifies that the email is shown in a read-only disabled field in edit mode
  - **Shows the Add Course button in edit mode:** Verifies that the + Add Course button is visible in edit mode
  - **Cancels edit mode when Cancel is clicked:** Verifies that clicking Cancel returns to view mode and hides the Save button
  - **Updates the username successfully:** Verifies that saving a new username shows a Profile updated successfully! notification
  - **Opens the Add Course modal in edit mode:** Verifies that clicking + Add Course opens the Add Course modal
  - **Logs out when Log Out is clicked:** Verifies that clicking Log Out navigates to the login page
  
### Unit Tests (Vitest + React Testing Library) - Sprint 2

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

 ### Unit Tests (Vitest + React Testing Library) - Sprint 3

- My Courses Page
  - **Renders my courses page with navbar and weekly view by default:** Verifies that the My Courses page loads with the navbar and Weekly view selected by default
  - **Switches to today view when clicking Today button:** Verifies that clicking the Today button activates the Today view and displays today's office hours
  - **Opens add office hours modal when clicking add button:** Verifies that clicking Add Office Hours opens the modal with the expected form fields
  - **Closes modal when clicking X button:** Verifies that clicking the close button dismisses the Add Office Hours modal
  - **Adds office hours to weekly view with form submission:** Verifies that submitting the Add Office Hours form adds office hours to the weekly view
  - **Deletes office hour from weekly view:** Verifies that deleting an office hour removes it from the weekly view
  - **Shows no queue today button for courses with no office hours:** Verifies that courses without office hours today display a No queue today button
  - **Changes office hour to no office hours today when deleted from today view:** Verifies that deleting an office hour from Today view updates the course to show no office hours today
  - **Displays join queue button for courses with office hours today:** Verifies that courses with office hours today display a Join Queue button
  - **Navigates back to student dashboard when clicking dashboard nav:** Verifies that clicking Dashboard navigates back to the student dashboard

- My Office Hours Page
  - **Switches to My Office Hours tab when clicked:** Verifies that clicking My Office Hours opens the office hours management view
  - **Switches back to Dashboard tab from My Office Hours:** Verifies that clicking Dashboard returns to the TA dashboard from the My Office Hours tab
  - **Renders the page header and Add Office Hours button:** Verifies that the My Office Hours page header, description, and Add Office Hours button are visible
  - **Renders all seven days of the week:** Verifies that all seven weekday sections are displayed on the page
  - **Shows empty state message for days with no office hours:** Verifies that days without scheduled office hours display the empty state message
  - **Renders existing office hours fetched from the backend:** Verifies that existing office hours and course information are displayed after loading
  - **Opens Add Office Hours modal when button is clicked:** Verifies that clicking + Add Office Hours opens the Add Office Hours modal
  - **Closes Add modal when Cancel is clicked:** Verifies that clicking Cancel closes the Add Office Hours modal
  - **Shows error toast if required fields are missing on add:** Verifies that submitting the add form with missing required fields shows an error toast
  - **Calls createOfficeHour:** Verifies that submitting valid office hour information triggers office hour creation
  - **Deletes an office hour and shows success toast:** Verifies that deleting an office hour removes it and shows a success toast
  - **Opens edit modal pre-filled with existing office hour data:** Verifies that clicking Edit opens the Edit Office Hours modal with existing values populated
  - **Updates an office hour and shows success toast on edit submission:** Verifies that editing an office hour updates it and shows a success toast

### Unit Tests (Vitest + React Testing Library) - Sprint 4

- Student Dashboard Page
  - **Renders the course dropdown:** Verifies that the course selection dropdown is rendered on the student dashboard
  - **Populates dropdown with schedule entries from API:** Verifies that listStudentSchedule is called on mount and the dropdown is present
  - **Shows placeholder when no schedule entries exist:** Verifies that a placeholder message appears when the student has no courses added
  - **Auto-selects course from navigation state:** Verifies that navigating from My Courses with an autoSelectLabel pre-selects the correct course and clears navigation state
  - **Join Queue button is disabled when no course is selected:** Verifies that the Join Queue button is disabled until a course is selected
  - **Joins queue and shows real-time section:** Verifies that selecting a course, waiting for the active queue poll, and clicking Join Queue shows the Real-Time Queue Status section
  - **Leaves queue and hides real-time section:** Verifies that clicking Cancel & Leave Queue hides the Real-Time Queue Status section and calls the leave queue API

- TA Dashboard Page
  - **Loads TA courses from API on mount:** Verifies that listMyTACourses is called when the TA dashboard mounts
  - **Loads office hours from API on mount:** Verifies that listOfficeHoursByTA is called with the TA's user ID when the dashboard mounts
  - **Displays Send Announcement button on closed dashboard:** Verifies that the Send Announcement button is visible on the closed dashboard state
  - **Calls startSession when Start Session is clicked on first queued student:** Verifies that clicking Start Session calls the startSession API with the correct queue and student IDs
  - **Shows session modal with Zoom details after starting session:** Verifies that after starting a session the modal appears showing the student name and Zoom meeting ID

- My Office Hours Page
  - **Displays existing office hours:** Verifies that loaded office hours are displayed with the correct time
  - **Shows course dropdown in modal with linked courses:** Verifies that the course dropdown in the Add Office Hours modal shows the TA's linked courses
  - **Disables Add Office Hours button when no courses are linked:** Verifies that the + Add Office Hours button is disabled when the TA has no linked courses

- My Courses Page
  - **Renders my courses page with weekly view by default:** Verifies that the My Courses page loads with the This Week's Office Hours section and Weekly button active
  - **Loads and displays student schedule from API:** Verifies that listStudentSchedule is called and the schedule data is displayed
  - **Switches to today view when clicking Today button:** Verifies that clicking Today switches to the Today's Office Hours view
  - **Opens Add Office Hours modal when clicking add button:** Verifies that clicking + Add Office Hours opens the modal with the Course step visible
  - **Loads student courses when modal is opened:** Verifies that getStudentCourses is called when the modal opens and courses appear as options
  - **Shows TA dropdown after selecting a course:** Verifies that selecting a course in the modal triggers listOfficeHoursByCourse and shows the TA step
  - **Add to Schedule button is disabled when no slot is selected:** Verifies that the Add to Schedule button is disabled until a time slot is chosen
  - **Closes modal when Cancel is clicked:** Verifies that clicking Cancel removes the Add to Schedule button from the DOM
  - **Calls removeFromStudentSchedule when × delete button is clicked:** Verifies that clicking the delete button calls removeFromStudentSchedule with the correct entry ID
  - **Navigates to /student when Dashboard nav is clicked:** Verifies that clicking Dashboard navigates to the student dashboard
  - **Navigates to /student with autoSelectLabel when Join Queue is clicked in today view:** Verifies that clicking Join Queue in the Today view navigates to /student with the correct autoSelectLabel in navigation state
  
## List Backend Unit Tests
### Backend Go Tests (`backend/api_e2e_test.go`) - Sprint 2

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

### Backend Go Tests (`backend/api_e2e_test.go`) — Sprint 3

**Helper:** `readSSEEventUntil` — reads the SSE stream until a named `event:` line appears and returns the JSON from the following `data:` line (full queue event envelope: `type`, `queue_id`, `payload`). Used by several notification SSE tests below.

- **`TestQueueState_PATCH_JoinBlockedWhenPausedOrClosed`**
  - **PATCH queue state (TA)**: `PATCH /api/queues/{id}/state` with `{"status":"paused"}` → **200 OK**
  - **Join while paused**: `POST /api/queues/{id}/join` → **409 Conflict** with error `queue is paused`
  - **PATCH closed**: `{"status":"closed"}` → join again → **409** with `queue is closed`
  - **Re-open**: `{"status":"open"}` → join → **201 Created**
  - **Wrong method**: `POST` to `/state` → **405 Method Not Allowed**

- **`TestSSE_EmitsQueueStateChanged`**
  - **SSE + state change**: connects to `GET /api/queues/{id}/events`, then TA `PATCH /api/queues/{id}/state` to `paused`
  - **Expectation**: stream receives `QUEUE_STATE_CHANGED` with payload `previous_status: open`, `status: paused`

- **`TestSSE_EmitsStudentUpNextAfterJoin`**
  - **SSE + join**: opens events stream, student joins an empty queue (becomes position 1)
  - **Expectation**: `STUDENT_UP_NEXT` envelope with `payload.student_id` matching the joiner and `payload.position` **1**

- **`TestSSE_EmitsAnnouncementSent`**
  - **SSE + announcement**: opens events stream, owning TA `POST /api/queues/{id}/announcement` with a message → **201 Created**
  - **Expectation**: `ANNOUNCEMENT_SENT` with same `message` and correct `ta_id` / `queue_id`

- **`TestPostAnnouncement_HappyPath`**
  - **POST announcement**: owning TA posts `{ "message": "..." }` → **201 Created**
  - **Response body**: includes `id`, `queue_id`, `message`, `created_at`

- **`TestPostAnnouncement_StudentForbidden`**
  - **Student cannot announce**: `POST /api/queues/{id}/announcement` with student token → **403 Forbidden**

- **`TestPostAnnouncement_NonOwningTAForbidden`**
  - **Non-owner TA**: another TA (not the queue owner) posts an announcement → **403 Forbidden**

- **`TestCreateOfficeHour_HappyPath`**
  - **Create office hour (TA)**: `POST /api/office-hours` with valid body → **201 Created** and persisted fields

- **`TestCreateOfficeHour_UnauthorizedAndStudentForbidden`**
  - **No token**: `POST /api/office-hours` → **401 Unauthorized**
  - **Student role**: same with student token → **403 Forbidden**

- **`TestCreateOfficeHour_OverlapRejected`**
  - **Overlapping slot**: second create for same TA/course/day with overlapping times → **409 Conflict**

- **`TestUpdateOfficeHour_HappyPath`**
  - **Update (TA owner)**: `PUT /api/office-hours/{id}` → **200 OK** and updated fields

- **`TestUpdateOfficeHour_OverlapRejected`**
  - **Update would overlap** another slot → **409 Conflict**

- **`TestUpdateOfficeHour_NotFoundAndForbidden`**
  - **Unknown id**: `PUT` → **404 Not Found**
  - **Wrong TA**: another TA updates someone else’s row → **403 Forbidden**

- **`TestDeleteOfficeHour_HappyPath`**
  - **Delete (TA owner)**: `DELETE /api/office-hours/{id}` → **204 No Content**

- **`TestDeleteOfficeHour_UnauthorizedNotFoundForbidden`**
  - **No token**: `DELETE` → **401**
  - **Unknown id**: → **404**
  - **Wrong TA**: → **403**

- **`TestListOfficeHoursByTA_HappyPathAndEmpty`**
  - **List by TA**: `GET /api/office-hours/ta/{ta_id}` → **200 OK** with expected rows; empty TA returns empty list

- **`TestListOfficeHoursByTA_InvalidID`**
  - **Invalid TA id** in path → **400 Bad Request**

- **`TestListOfficeHoursByCourse_HappyPath`**
  - **List by course**: `GET /api/office-hours/course/{course_id}` → **200 OK** with expected office hours

- **`TestListOfficeHoursByCourse_InvalidID`**
  - **Invalid course id** in path → **400 Bad Request**
 
### Backend Go Tests (`backend/tests/`) — Sprint 4

**Helper:** `newTestServer`, `setupTestDB`, `doJSON`, `parseAuthUser`, `stdErrorBody` (`backend/tests/helpers_test.go`) — shared migrated DB + chi router for HTTP assertions.

- **`TestAuthMiddleware_Missing_Returns401WithCode`**
  - **No token** on protected route: `POST /api/queues` → **401 Unauthorized** with JSON **`code`: `auth_required`**

- **`TestAuthMiddleware_InvalidToken_Returns401WithCode`**
  - **Invalid Bearer token**: same route → **401** with **`code`: `auth_required`**

- **`TestRoleMiddleware_StudentOnTARoute_Returns403WithCodeAndRequiredRole`**
  - **Student** calls TA-only `POST /api/queues` → **403 Forbidden** with **`code`: `role_forbidden`** and **`details.required_role`: `ta`**

- **`TestRoleMiddleware_TAOnStudentRoute_Returns403WithCodeAndRequiredRole`**
  - **TA** calls student-only `POST /api/queues/{id}/join` → **403** with **`role_forbidden`** and **`required_role`: `student`**

- **`TestRoleMiddleware_TACannotLeaveQueue`**
  - **TA** calls `POST /api/queues/{id}/leave` → **403** with **`role_forbidden`** / **`required_role`: `student`**

- **`TestOwnership_NonOwningTAGetsStructured403_OnNext`**
  - **Non-owner TA** calls `POST /api/queues/{id}/next` → **403** with **`code`: `not_queue_owner`**

- **`TestOwnership_NonOwningTAGetsStructured403_OnUpdateState`**
  - **Non-owner TA** calls `PATCH /api/queues/{id}/state` → **403** with **`not_queue_owner`**

- **`TestOwnership_NonOwningTAGetsStructured403_OnOfficeHourUpdate`**
  - **Non-owner TA** calls `PUT /api/office-hours/{id}` → **403** with **`code`: `not_resource_owner`**

- **`TestConcurrentJoins_PositionsUniqueAndContiguous`**
  - **Concurrent joins**: many students join in parallel → all **201**; **`position`** values are **unique** and cover **1..N**

- **`TestDuplicateJoin_UniqueConstraintReturns409`**
  - **Second join** same student → **409 Conflict**, message **`already in queue`** (DB uniqueness path)

- **`TestConcurrentServeNext_NeverDoubleServesSameStudent`**
  - **Concurrent `/next`**: exactly **N** successes for **N** waiting students; **no duplicate `student_id`** served; extras → **404**

- **`TestTACourses_AddAndList`**
  - **TA** `POST /api/ta/courses` with **`code`/`name`** → **201**; idempotent repeat **`course_id`** → **201**; **`GET /api/ta/courses`** → **200** with linked course

- **`TestTACourses_StudentForbidden`**
  - **Student** `POST /api/ta/courses` → **403 Forbidden**

- **`TestSessionHistory_Unauthorized`**
  - **`GET /api/session-history`** without token → **401 Unauthorized**

- **`TestSessionHistory_TAandStudent_SeeRelevantRows`**
  - After **two `/next`** calls on a two-student queue: **TA** history lists **one** completed session for first student; **served student** sees **one** row; **waiting student** sees **none**

- **`TestSessionLog_WrittenOnSecondNext`**
  - Two students in line: **first `/next`** leaves **`session_logs` empty**; **second `/next`** persists **one** row for the first student

- **`TestQueueClose_FlushesLastSessionLog`**
  - Student **in session** (`/next` once): **`PATCH` closed** → **`session_logs`** has **one** row; **`queue_entries`** cleared

- **`TestQueueClose_ClearsLineWithoutSessionLog`**
  - Student **waiting only** (no `/next`): **`PATCH` closed** → **no `session_logs`** rows

- **`TestQueueState_InvalidTransitionIdempotent`**
  - **`closed` → `paused`** → **409** with **`invalid_state_transition`** / **`details.allowed`**; **`open` → `open`** PATCH → **200** (idempotent)

- **`TestGetQueueResponse_IncludesETAMetadata`**
  - **`GET /api/queues/{id}`** includes **`average_session_duration_seconds`**, **`ta_average_session_duration_seconds`**, **`estimated_wait_time_seconds`**, per-entry **`estimated_wait_seconds`**; single waiter at front → **`is_empty`** false and front ETA **0**

- **`TestGetQueue_EmptyQueue_IsEmptyTrue`**
  - Fresh queue with **no joins**: **`GET /api/queues/{id}`** → **`is_empty`: true**, **`entries`** empty

## Updated Documentation for Backend API 

The backend is an **HTTP API** in **Go** with **chi** and **PostgreSQL** (`backend/internal/routes/routes.go`, `backend/cmd/server/main.go`, `go.mod`). Most routes use **JSON** request/response bodies. Queues follow **resource-style** paths (`/api/queues`, `/api/queues/{id}`), while **auth** and queue flows mostly use **POST** command paths (`/api/login`, `/api/register`, `/join`, `/leave`, `/next`) plus **`PATCH /api/queues/{id}/state`** — common for web apps, but not a strict REST-only design. Catalog and history surface as **`GET /api/courses`**, TA **`/api/ta/courses`**, and **`GET /api/session-history`**; protected handlers sit in **TA** vs **student** route groups with role middleware. **Live updates** use **Server-Sent Events (SSE)** on **`GET /api/queues/{id}/events`** (`Content-Type: text/event-stream`), not a JSON response body.

### Run the server

1. Create a Postgres database (e.g. `officehours`).
2. In **`backend/`**, copy `.env.example` → `.env` and set the DB connection string and **`JWT_SECRET`**.
3. Run `go mod tidy` then `go run ./cmd/server`. The process listens on **`PORT`** (default **`8080`**). Smoke test: **`GET /health`** → `200` `{"status":"ok"}` if the DB is reachable.

### Conventions

- **Base URL:** `http://localhost:<PORT>` (replace `8080` if you set `PORT`).
- **Bodies / errors:** Most endpoints use JSON. **SSE** (`/events`) uses **`text/event-stream`**, not JSON for the overall response. Wrong HTTP method on a handler → **`405`**. Failures usually look like `{"error":"<message>"}`; many DB failures return **`500`** with `"database error"`.
- **Structured errors (Sprint 4):** Auth, role, ownership, invalid queue-state transitions, and many validation failures use **`backend/internal/httperr`** JSON **`{ "code", "message", "details" }`** — overlapping with legacy **`{"error":"..."}`** on some older handlers (see Sprint 4 middleware tests for **`auth_required`**, **`role_forbidden`**, **`not_queue_owner`**, **`not_resource_owner`**, **`invalid_state_transition`**).
- **CORS:** `Access-Control-Allow-Methods: GET, POST, OPTIONS`; headers allowed include `Authorization` (see `corsMiddleware` in `routes.go`).
- **CORS (complete list):** `corsMiddleware` also allows **PUT**, **PATCH**, **DELETE**, and **OPTIONS** preflight (`routes.go`).

### Authentication

After **`POST /api/register`** or **`POST /api/login`**, responses include a **`token`** string. Send it on protected routes as:

`Authorization: Bearer <token>`

The JWT carries **`user_id`**, **`email`**, and **`role`** (`student` or `ta`) — see `backend/internal/auth/jwt.go`. Missing token, malformed header, or invalid/expired token → **`401`** with `{"error":"authorization required"}`.

On TA-/student-group routes guarded by **`RequireAuth`** / **`RequireRole`**, **`401`** responses may instead use the structured envelope with **`code`: `auth_required`** (see **`TestAuthMiddleware_*`**).

**Routes that require a valid Bearer token:**

| Route | Extra rule (from code) |
|--------|-------------------------|
| `POST /api/queues` | Caller must have **`role: ta`** (`403` otherwise). |
| `POST /api/queues/{id}/join` | **`role: student`** (middleware); **`403`** for TA. |
| `POST /api/queues/{id}/leave` | **`role: student`** (middleware); **`403`** for TA. |
| `POST /api/queues/{id}/next` | **`role: ta`** and JWT **`user_id`** must match the queue’s **`ta_id`** (`403` otherwise). |
| `PATCH /api/queues/{id}/state` | **`role: ta`** and JWT **`user_id`** must match the queue’s **`ta_id`** (`403` otherwise). |
| `POST /api/queues/{id}/announcement` | **`role: ta`** and JWT **`user_id`** must match the queue’s **`ta_id`** (`403` otherwise). |
| `POST /api/office-hours` | **`role: ta`** (`403` for students). |
| `PUT /api/office-hours/{id}` | **`role: ta`** and must own the row (`403` otherwise). |
| `DELETE /api/office-hours/{id}` | **`role: ta`** and must own the row (`403` otherwise). |
| `POST /api/ta/courses` | **`role: ta`** (`403` for students). |
| `GET /api/ta/courses` | **`role: ta`**. |
| `DELETE /api/ta/courses/{id}` | **`role: ta`** · removes link for **`course_id`** path · **`404`** if not linked. |
| `GET /api/session-history` | Any authenticated **`ta`** or **`student`** · rows filtered per role (see **`TestSessionHistory_TAandStudent_SeeRelevantRows`**). |

**Note (Sprint 4):** **`POST /api/queues/{id}/join`** and **`POST .../leave`** are **student-only** at the middleware layer — a TA bearer token gets **`403`** **`role_forbidden`** (**`TestRoleMiddleware_TACannotLeaveQueue`**).

**Public (no token):** **`GET /api/courses`** (courses with ≥1 TA), **`GET /api/office-hours/ta/{ta_id}`**, **`GET /api/office-hours/course/{course_id}`**, and queue read/SSE routes such as **`GET /api/queues/{id}`**, **`GET /api/queues/active`**, **`GET /api/queues/{id}/events`**.

### Endpoints

Path parameter **`{id}`** is the numeric queue id (`chi` route `/api/queues/{id}`). If **`{id}`** is not a valid integer, **`GET /api/queues/{id}`**, **join**, **leave**, **next**, and **events** return **`400`** `{"error":"invalid queue id"}`.

| Method | Path | What it does | Success |
|--------|------|----------------|---------|
| `GET` | `/health` | Ping database | **`200`** `{"status":"ok"}` · **`503`** if ping fails (`{"status":"unavailable","error":"database"}`) |
| `POST` | `/api/register` | Create user | **`200`** JSON with `token` and `user` `{ id, username, email, role }`. **`400`** validation / bad role · **`409`** duplicate username or email |
| `POST` | `/api/login` | Login | **`200`** same shape as register · **`401`** bad email/password |
| `GET` | `/api/courses` | Courses that have ≥1 TA | **`200`** `{ "courses": [ { id, code, name, color }, ... ] }` |
| `POST` | `/api/queues` | TA starts a queue | **`201`** `{ id, course_id, ta_id, status, created_at }` — **`status`** is **`open`**. Optional body: `{ "course_id": 0 }`; omitting the body ⇒ **`course_id` is 0** |
| `GET` | `/api/queues/{id}` | Queue + waiting students | **`200`** `{ id, course_id, ta_id, status, created_at, entries: [...] }`. Each entry: `id`, `queue_id`, `student_id`, `position`, `joined_at`, `username`. Ordered by **position**, then **joined_at**. **`404`** unknown queue |
| `POST` | `/api/queues/{id}/join` | Authenticated user joins | **`201`** `{ id, queue_id, position, joined_at }`. Queue must exist and **`status`** must be **`open`**. Duplicate student in same queue → **`409`**. Triggers SSE (below) |
| `POST` | `/api/queues/{id}/leave` | Authenticated user leaves | **`204`** empty body; positions renumbered. **`404`** if no row was removed (`not in queue`) |
| `POST` | `/api/queues/{id}/next` | Owning TA removes front of line | **`200`** `{ "queue_id", "status": "in_session", "student": { …entry } }`. **`404`** no queue or empty queue · **`409`** queue not **`open`**. Triggers SSE |
| `PATCH` | `/api/queues/{id}/state` | Owning TA sets `open` / `paused` / `closed` | **`200`** `{ id, status }`. Emits **`QUEUE_STATE_CHANGED`** on SSE · invalid transition **`409`** (**`invalid_state_transition`**) · closing may flush **`session_logs`** / clear entries (see Sprint 4 summary above) |
| `POST` | `/api/queues/{id}/session` | Owning TA starts persisted Zoom session for student **in_session** | **`201`** session JSON with Zoom fields · provisioning runs before DB insert (**`backend/internal/queue/session.go`**) |
| `POST` | `/api/queues/{id}/announcement` | Owning TA posts `{ "message": "..." }` (stored + SSE) | **`201 Created`** body includes `id`, `queue_id`, `message`, `created_at`. **`403`** wrong role or non-owner TA |
| `POST` | `/api/ta/courses` | TA links self to catalog course (`course_id` **or** `code` + optional `name`/`color`) | **`201`** `{ ta_id, course }` · idempotent link (**`TestTACourses_AddAndList`**) · structured **`400`** if neither reference provided |
| `GET` | `/api/ta/courses` | TA lists linked courses | **`200`** `{ "courses": [...] }` |
| `DELETE` | `/api/ta/courses/{id}` | TA removes link by **`course_id`** | **`204`** · **`404`** if not linked |
| `GET` | `/api/session-history` | Completed sessions visible to caller | **`200`** `{ "sessions": [...] }` · **`401`** without token (**`TestSessionHistory_Unauthorized`**) |

**Queue snapshot (`GET /api/queues/{id}`):** Response also includes **`is_empty`**, **`average_session_duration_seconds`**, **`ta_average_session_duration_seconds`**, **`estimated_wait_time_seconds`**, and each entry may include **`estimated_wait_seconds`** (see **`TestGetQueueResponse_IncludesETAMetadata`**, **`TestGetQueue_EmptyQueue_IsEmptyTrue`**).

**Queue `status` in the database** can be `open`, `paused`, or `closed` (`migrate.go`). **`POST /api/queues`** creates **`open`** queues; **join** and **next** require **`open`** (join returns **409** when paused or closed). The owning TA can change status via **`PATCH /api/queues/{id}/state`** with JSON `{"status":"open"|"paused"|"closed"}` (see queue handler tests above).

### Real-time updates (SSE)

**`GET /api/queues/{id}/events`** — **Server-Sent Events**, no auth.

1. Response headers include `Content-Type: text/event-stream`.
2. First line is a comment: `: connected` (keeps the stream open).
3. Later lines look like: `event: <TYPE>` then `data: <JSON>` (blank line between events).

The **`data`** line is JSON for the **full queue event** (`type`, `queue_id`, optional `payload`) — see `StreamQueueEvents` in `backend/internal/queue/events.go`. Events with no extra fields omit `payload` in JSON.

**`event` names (non-exhaustive):** `STUDENT_JOINED` · `STUDENT_LEFT` · `STUDENT_SERVED` · `STUDENT_UP_NEXT` · `ANNOUNCEMENT_SENT` · `QUEUE_UPDATED` · `QUEUE_STATE_CHANGED`.

Join / leave / **next** / queue state changes / announcements publish the corresponding events after the database change succeeds (where applicable).

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

From **`backend/`** with Postgres and `.env` configured: **`go test -v ./tests`** (see **`backend/tests/`**).
