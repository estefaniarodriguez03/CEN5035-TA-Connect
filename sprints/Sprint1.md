# TA Connect Sprint 1

## User Stories

* As a student, I want to join a real-time queue for office hours to track my position.
* As a student, I want the application to show me an estimated wait time for my turn so that I can decide to stay in the queue or leave.
* As a student, I want to receive push notifications for when it is about to be my turn so that I am informed about my position and can be ready.
* As a student, I want to view the TA’s office hour schedule so that I know what sessions are available.
* As a student, I want the ability to cancel my spot in the queue if my plans change.
* As a student, I want to be able to view the office hours for all my courses in a centralized area for easy access.
* As a TA, I want to see a real-time, labeled queue of the students waiting to attend my office hour session so that I can see the order in which they arrived.
* As a TA, I want to launch an office hour session with the next student in line with a single click so that I can get to work efficiently.
* As a TA, I want to receive push notifications when students join or leave the queue so I can be aware of the current session status.
* As a TA, I want to update my personal office hour schedule in the application so that students can see the correct times.
* As a TA, I want to send announcements such as running late or extending time to everyone in the queue so students are kept up to date.
* As a TA, I want the ability to pause or close the queue so I can finish on time.

## What Issues We Planned to Address

### Front-End

* Create authentication page with login and registration options
* Create the "Home" landing page for a student
* Create the "Home" landing page for a TA

### Back-End

* Establish Go to Postgres connection 
* Create database schema for users and office hours
* Create user authentication credentials to allow login and registration

## Which Ones Were Successfully Completed

### Front-End (Completed)

* Create authentication page with login and registration options
  * Created landing page for application including a welcome screen with options to login and register
  * Connected frontend to backend database to 
* Create the "Home" landing page for a student
* Create the "Home" landing page for a TA
  
### Back-End (Completed)

* Establish Go to Postgres connection
  * Connected backend to Postgres database for frontend to access
* Create database schema for users
  * Created user schema including username, email, password, and role
  * Created office hours scheme including TA name, email, and availability
* Create user authentication credentials to allow login and registration
  * Created user authentication credentials for student and TA role for different types of login

## Which Ones Didn't and Why?

### Front-End (Not Completed)

* User login and registration currently only allow local accounts as opposed to GatorLink. Other forms of authentication will be added later.
* Validation of user role currently does not exist. We will need to determine how to tackle this problem in upcoming sprints.

### Back-End (Not Completed)

* More information here …

## Front-End and Back-End Demo Links

**Front End:** [insert video link]  

Record a video with both members demoing your frontend work. Use a mocked up backend, if necessary.

**Back End:** [insert video link]

Record a video with both members demoing your backend work. Use the command line or Postman.

