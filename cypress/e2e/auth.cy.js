describe('Visit TA Connect Home Page', () => {
  it('Visits the TA Connect home page', () => {
    cy.visit('/')
  })
})

describe("Login Page", () => {
  beforeEach(() => {
    cy.visit("/login");
  });

  it("Displays the login form", () => {
    cy.get('input[placeholder="Email"]').should("be.visible");
    cy.get('input[placeholder="Password"]').should("be.visible");
    cy.get('button[type="submit"]').should("contain", "Login");
  });

  it("shows an error for invalid credentials", () => {
    cy.get('input[placeholder="Email"]').type("wrong@ufl.edu");
    cy.get('input[placeholder="Password"]').type("wrongpassword");
    cy.get('button[type="submit"]').click();
    cy.contains("Login failed").should("be.visible");
  });

  it("has a link to the register page", () => {
    cy.contains("Register").click();
    cy.url().should("include", "/register");
  });
});

describe("Register Page", () => {
  beforeEach(() => {
    cy.visit("/register");
  });

  it("displays the registration form", () => {
    cy.get('input[placeholder="Name"]').should("be.visible");
    cy.get('input[placeholder="Email"]').should("be.visible");
    cy.get('input[placeholder="Password"]').should("be.visible");
    cy.get("select").should("be.visible");
  });

  it("can select TA role", () => {
    cy.get("select").select("ta");
    cy.get("select").should("have.value", "ta");
  });

  it("defaults to student role", () => {
    cy.get("select").should("have.value", "student");
  });

  it("has a link back to the login page", () => {
    cy.get('.auth-footer a').click();
    cy.url().should("include", "/login");
  });
});

describe("Student Dashboard", () => {
  beforeEach(() => {
    cy.visit("/login");
    cy.get('input[placeholder="Email"]').type("student@ufl.edu");
    cy.get('input[placeholder="Password"]').type("correctpassword");
    cy.get('button[type="submit"]').click();
    cy.url().should("include", "/student");
  });

  it("displays the student dashboard with welcome message", () => {
    cy.contains("Welcome Back").should("be.visible");
    cy.contains("TA Connect").should("be.visible");
  });

  it("displays the course dropdown", () => {
    cy.get(".course-dropdown").should("be.visible");
  });

  it("displays all five days in the weekly schedule", () => {
    cy.contains("Monday").should("be.visible");
    cy.contains("Tuesday").should("be.visible");
    cy.contains("Wednesday").should("be.visible");
    cy.contains("Thursday").should("be.visible");
    cy.contains("Friday").should("be.visible");
  });

  it("displays the notification icon in the navbar", () => {
    cy.get(".notification-icon").should("be.visible");
  });

  it("displays the Estimated Wait Time card", () => {
    cy.contains("Estimated Wait Time").should("be.visible");
  });

  it("displays the weekly schedule section", () => {
    cy.contains("This Week's Office Hours").should("be.visible");
  });

  it("navigates to the My Courses page when My Courses tab is clicked", () => {
    cy.contains("My Courses").click();
    cy.url().should("include", "/student/my-courses");
  });

  it("navigates to the profile page when the profile icon is clicked", () => {
    cy.get(".profile-icon").click();
    cy.url().should("include", "/profile");
  });
});

describe("Student Dashboard — Queue Interaction", () => {
  beforeEach(() => {
    cy.visit("/login");
    cy.get('input[placeholder="Email"]').type("student@ufl.edu");
    cy.get('input[placeholder="Password"]').type("correctpassword");
    cy.get('button[type="submit"]').click();
    cy.url().should("include", "/student");
  });

  it("displays the Estimated Wait Time card", () => {
    cy.contains("Estimated Wait Time").should("be.visible");
  });

  it("shows helper text when no course is selected", () => {
    cy.contains("Select a course to see the queue").should("be.visible");
  });

  it("Join Queue button is disabled when no course is selected", () => {
    cy.contains("Join Queue").should("be.disabled");
  });

  it("navigates to My Courses page when My Courses tab is clicked", () => {
    cy.contains("My Courses").click();
    cy.url().should("include", "/student/my-courses");
  });
});

describe("Student — My Courses Page", () => {
  beforeEach(() => {
    cy.visit("/login");
    cy.get('input[placeholder="Email"]').type("student@ufl.edu");
    cy.get('input[placeholder="Password"]').type("correctpassword");
    cy.get('button[type="submit"]').click();
    cy.url().should("include", "/student");
    cy.contains("My Courses").click();
    cy.url().should("include", "/student/my-courses");
  });

  it("displays the My Courses page header", () => {
    cy.contains("My Courses - Office Hours").should("be.visible");
  });

  it("displays the Add Office Hours button", () => {
    cy.contains("+ Add Office Hours").should("be.visible");
  });

  it("displays weekly and today view toggle buttons", () => {
    cy.contains("Weekly").should("be.visible");
    cy.contains("Today").should("be.visible");
  });

  it("displays filter dropdowns", () => {
    cy.contains("Filter by Course").should("be.visible");
    cy.contains("Filter by TA").should("be.visible");
  });

  it("switches to Today view when clicked", () => {
    cy.contains("Today").click();
    cy.contains("Today's Office Hours").should("be.visible");
  });

  it("opens the Add Office Hours modal when button is clicked", () => {
    cy.contains("+ Add Office Hours").click();
    cy.contains("Add Office Hours").should("be.visible");
  });

  it("shows course dropdown as first step in modal", () => {
    cy.contains("+ Add Office Hours").click();
    cy.contains("Course").should("be.visible");
    cy.get(".form-select").first().should("be.visible");
  });

  it("shows TA dropdown after selecting a course", () => {
    cy.contains("+ Add Office Hours").click();
    cy.get(".form-select").first().select(1);
    cy.contains("TA").should("be.visible");
  });

  it("Add to Schedule button is disabled when no slot is selected", () => {
    cy.contains("+ Add Office Hours").click();
    cy.contains("Add to Schedule").should("be.disabled");
  });

  it("navigates back to dashboard when Dashboard tab is clicked", () => {
    cy.contains("Dashboard").click();
    cy.url().should("include", "/student");
  });
});

describe("Student Profile Page", () => {
  beforeEach(() => {
    cy.visit("/login");
    cy.get('input[placeholder="Email"]').type("student@ufl.edu");
    cy.get('input[placeholder="Password"]').type("correctpassword");
    cy.get('button[type="submit"]').click();
    cy.url().should("include", "/student");
    cy.get(".profile-icon").click();
    cy.url().should("include", "/profile");
  });

  // View mode
  it("displays the profile page header", () => {
    cy.contains("Welcome to your Profile!").should("be.visible");
  });

  it("displays the student's name", () => {
    cy.get(".profile-name").should("be.visible");
  });

  it("displays the student's email", () => {
    cy.contains("student@ufl.edu").should("be.visible");
  });

  it("displays the Major field", () => {
    cy.contains("Major:").should("be.visible");
  });

  it("displays the Year field", () => {
    cy.contains("Year:").should("be.visible");
  });

  it("displays the Courses section", () => {
    cy.contains("Courses:").should("be.visible");
  });

  it("displays the Edit Profile button", () => {
    cy.contains("Edit Profile").should("be.visible");
  });

  it("displays the Log Out button", () => {
    cy.contains("Log Out").should("be.visible");
  });

  it("displays the navbar with Dashboard and My Courses tabs", () => {
    cy.contains("Dashboard").should("be.visible");
    cy.contains("My Courses").should("be.visible");
  });

  it("navigates to the dashboard when Dashboard tab is clicked", () => {
    cy.contains("Dashboard").click();
    cy.url().should("include", "/student");
  });

  it("navigates to My Courses when My Courses tab is clicked", () => {
    cy.contains("My Courses").click();
    cy.url().should("include", "/student/my-courses");
  });

  it("logs out when Log Out is clicked", () => {
    cy.contains("Log Out").click();
    cy.url().should("include", "/login");
  });

  // Edit mode
  it("enters edit mode when Edit Profile is clicked", () => {
    cy.contains("Edit Profile").click();
    cy.contains("Save").should("be.visible");
    cy.contains("Cancel").should("be.visible");
  });

  it("shows name input field in edit mode", () => {
    cy.contains("Edit Profile").click();
    cy.get('input[placeholder="Enter your name"]').should("be.visible");
  });

  it("shows email as disabled in edit mode", () => {
    cy.contains("Edit Profile").click();
    cy.get('input[placeholder="Email (read-only)"]').should("be.disabled");
  });

  it("shows major input field in edit mode", () => {
    cy.contains("Edit Profile").click();
    cy.get('input[placeholder="Enter your major"]').should("be.visible");
  });

  it("shows year dropdown in edit mode", () => {
    cy.contains("Edit Profile").click();
    cy.get(".profile-form-select").should("be.visible");
  });

  it("year dropdown contains all expected options", () => {
    cy.contains("Edit Profile").click();
    cy.get(".profile-form-select").should("contain", "Freshman");
    cy.get(".profile-form-select").should("contain", "Sophomore");
    cy.get(".profile-form-select").should("contain", "Junior");
    cy.get(".profile-form-select").should("contain", "Senior");
    cy.get(".profile-form-select").should("contain", "Graduate");
    cy.get(".profile-form-select").should("contain", "PhD");
  });

  it("shows the Add Course button in edit mode", () => {
    cy.contains("Edit Profile").click();
    cy.contains("+ Add Course").should("be.visible");
  });

  it("cancels edit mode when Cancel is clicked", () => {
    cy.contains("Edit Profile").click();
    cy.contains("Cancel").click();
    cy.contains("Edit Profile").should("be.visible");
    cy.contains("Save").should("not.exist");
  });

  it("restores original name when Cancel is clicked after editing", () => {
    cy.get(".profile-name").invoke("text").then((originalName) => {
      cy.contains("Edit Profile").click();
      cy.get('input[placeholder="Enter your name"]').clear().type("TempName");
      cy.contains("Cancel").click();
      cy.get(".profile-name").should("contain", originalName.trim());
    });
  });

  it("updates the name successfully and shows success toast", () => {
    cy.contains("Edit Profile").click();
    cy.get('input[placeholder="Enter your name"]').clear().type("Updated Student");
    cy.contains("Save").click();
    cy.contains("Profile updated successfully!").should("be.visible");
  });

  it("can select a different year in edit mode", () => {
    cy.contains("Edit Profile").click();
    cy.get(".profile-form-select").select("Senior");
    cy.get(".profile-form-select").should("have.value", "Senior");
  });

  it("can update the major field", () => {
    cy.contains("Edit Profile").click();
    cy.get('input[placeholder="Enter your major"]').clear().type("Computer Engineering");
    cy.get('input[placeholder="Enter your major"]').should("have.value", "Computer Engineering");
  });

  // Add Course modal
  it("opens the Add Course modal when + Add Course is clicked", () => {
    cy.contains("Edit Profile").click();
    cy.contains("+ Add Course").click();
    cy.contains("Add Course").should("be.visible");
  });

  it("displays the course selection dropdown in the Add Course modal", () => {
    cy.contains("Edit Profile").click();
    cy.contains("+ Add Course").click();
    cy.get(".modal-select-blue").should("be.visible");
  });

  it("closes the Add Course modal when X is clicked", () => {
    cy.contains("Edit Profile").click();
    cy.contains("+ Add Course").click();
    cy.contains("Add Course").should("be.visible");
    cy.get(".modal-close-btn-white").click();
    cy.get(".modal-select-blue").should("not.exist");
  });

  it("adds a course from the modal and shows it in the courses list", () => {
    cy.contains("Edit Profile").click();
    cy.contains("+ Add Course").click();
    cy.get(".modal-select-blue").then(($select) => {
      if ($select.find("option").length > 0) {
        cy.get(".modal-select-blue").select(0);
        cy.get(".modal-add-btn-blue").click();
        cy.get(".profile-courses-list").should("not.be.empty");
      }
    });
  });

  it("does not show already enrolled courses in the Add Course modal dropdown", () => {
    cy.contains("Edit Profile").click();
    cy.get(".profile-courses-list .profile-course-text").then(($courses) => {
      if ($courses.length > 0) {
        const enrolledCourse = $courses.first().text().trim();
        cy.contains("+ Add Course").click();
        cy.get(".modal-select-blue").should("not.contain", enrolledCourse);
      }
    });
  });

  it("can remove a course in edit mode", () => {
    cy.contains("Edit Profile").click();
    cy.get(".profile-remove-course-btn").then(($btns) => {
      if ($btns.length > 0) {
        const initialCount = $btns.length;
        cy.get(".profile-remove-course-btn").first().click();
        cy.get(".profile-remove-course-btn").should("have.length", initialCount - 1);
      }
    });
  });
});

describe("TA Dashboard", () => {
  beforeEach(() => {
    cy.visit("/login");
    cy.get('input[placeholder="Email"]').type("ta@ufl.edu");
    cy.get('input[placeholder="Password"]').type("correctpassword");
    cy.get('button[type="submit"]').click();
    cy.url().should("include", "/ta");
  });

  it("displays the TA dashboard with welcome message", () => {
    cy.contains("Welcome Back").should("be.visible");
    cy.contains("TA Connect").should("be.visible");
  });

  it("displays the correct navigation tabs", () => {
    cy.contains("Dashboard").should("be.visible");
    cy.contains("My Office Hours").should("be.visible");
    cy.contains("Queue").should("be.visible");
  });

  it("displays the start queue button as disabled when no office hour is selected", () => {
    cy.contains("Select Time to Start Live Queue").should("be.disabled");
  });

  it("displays all stat cards on the dashboard", () => {
    cy.contains("Students Helped Today").should("be.visible");
    cy.contains("Avg Wait Time").should("be.visible");
    cy.contains("Current Queue Length").should("be.visible");
    cy.contains("Longest Wait Time").should("be.visible");
    cy.contains("Most Common Topic").should("be.visible");
    cy.contains("Avg Session Duration").should("be.visible");
  });

  it("displays the weekly office hours sidebar", () => {
    cy.contains("This Week's Office Hours").should("be.visible");
  });

  it("displays the Send Announcement button", () => {
    cy.contains("Send Announcement").should("be.visible");
  });
});

describe("TA Dashboard — Live Queue", () => {
  beforeEach(() => {
    cy.visit("/login");
    cy.get('input[placeholder="Email"]').type("ta@ufl.edu");
    cy.get('input[placeholder="Password"]').type("correctpassword");
    cy.get('button[type="submit"]').click();
    cy.url().should("include", "/ta");
  });

  it("enables the start queue button after selecting an office hour", () => {
    cy.get(".schedule-card").first().click();
    cy.contains("Start Office Hours Live Queue").should("not.be.disabled");
  });

  it("shows the live queue view after starting the queue", () => {
    cy.get(".schedule-card").first().click();
    cy.contains("Start Office Hours Live Queue").click();
    cy.contains("Live Queue").should("be.visible");
  });

  it("displays queue controls in the live queue view", () => {
    cy.get(".schedule-card").first().click();
    cy.contains("Start Office Hours Live Queue").click();
    cy.contains("Pause Queue").should("be.visible");
    cy.contains("Close Queue").should("be.visible");
  });

  it("displays queue stats in the live queue view", () => {
    cy.get(".schedule-card").first().click();
    cy.contains("Start Office Hours Live Queue").click();
    cy.contains("Total in Queue").should("be.visible");
    cy.contains("Longest Wait").should("be.visible");
    cy.contains("Avg Wait Time").should("be.visible");
  });

  it("shows empty queue state when no students are waiting", () => {
    cy.get(".schedule-card").first().click();
    cy.contains("Start Office Hours Live Queue").click();
    cy.contains("No students in queue").should("be.visible");
  });

  it("toggles to Resume Queue after clicking Pause Queue", () => {
    cy.get(".schedule-card").first().click();
    cy.contains("Start Office Hours Live Queue").click();
    cy.contains("Pause Queue").click();
    cy.contains("Resume Queue").should("be.visible");
  });

  it("toggles back to Pause Queue after clicking Resume Queue", () => {
    cy.get(".schedule-card").first().click();
    cy.contains("Start Office Hours Live Queue").click();
    cy.contains("Pause Queue").click();
    cy.contains("Resume Queue").click();
    cy.contains("Pause Queue").should("be.visible");
  });

  it("returns to the dashboard view after closing the queue", () => {
    cy.get(".schedule-card").first().click();
    cy.contains("Start Office Hours Live Queue").click();
    cy.contains("Close Queue").click();
    cy.contains("Welcome Back").should("be.visible");
  });
});

describe("TA Dashboard — My Office Hours Tab", () => {
  beforeEach(() => {
    cy.visit("/login");
    cy.get('input[placeholder="Email"]').type("ta@ufl.edu");
    cy.get('input[placeholder="Password"]').type("correctpassword");
    cy.get('button[type="submit"]').click();
    cy.url().should("include", "/ta");
    cy.contains("My Office Hours").click();
  });

  it("navigates to My Office Hours tab when clicked", () => {
    cy.contains("Manage your office hour schedule and availability").should("be.visible");
  });

  it("displays the Add Office Hours button", () => {
    cy.contains("+ Add Office Hours").should("be.visible");
  });

  it("displays all seven days of the week", () => {
    cy.contains("Sunday").should("be.visible");
    cy.contains("Monday").should("be.visible");
    cy.contains("Tuesday").should("be.visible");
    cy.contains("Wednesday").should("be.visible");
    cy.contains("Thursday").should("be.visible");
    cy.contains("Friday").should("be.visible");
    cy.contains("Saturday").should("be.visible");
  });

  it("shows empty state for days with no office hours", () => {
    cy.contains("No office hours scheduled").should("be.visible");
  });

  it("opens the Add Office Hours modal when button is clicked", () => {
    cy.contains("+ Add Office Hours").click();
    cy.contains("Add Office Hours").should("be.visible");
    cy.get('input[type="time"]').should("have.length", 2);
  });

  it("shows course dropdown in the modal", () => {
    cy.contains("+ Add Office Hours").click();
    cy.contains("Course").should("be.visible");
    cy.get("select.oh-modal-input").should("be.visible");
  });

  it("Add Schedule button is disabled when required fields are missing", () => {
    cy.contains("+ Add Office Hours").click();
    cy.contains("Add Schedule").should("be.disabled");
  });

  it("opens the edit modal pre-filled when Edit is clicked", () => {
    cy.contains("Edit").first().click();
    cy.contains("Edit Office Hours").should("be.visible");
    cy.get('input[type="time"]').first().should("not.have.value", "");
  });

  it("deletes an office hour and shows success toast", () => {
    cy.contains("Delete").first().click();
    cy.contains("Office hours deleted").should("be.visible");
  });

  it("navigates back to Dashboard tab from My Office Hours", () => {
    cy.contains("Dashboard").click();
    cy.contains("Welcome Back").should("be.visible");
    cy.contains("Manage your office hour schedule and availability").should("not.exist");
  });
});

describe("TA Profile Page", () => {
  beforeEach(() => {
    cy.visit("/login");
    cy.get('input[placeholder="Email"]').type("ta@ufl.edu");
    cy.get('input[placeholder="Password"]').type("correctpassword");
    cy.get('button[type="submit"]').click();
    cy.url().should("include", "/ta");
    cy.get(".profile-icon").click();
    cy.url().should("include", "/profile");
  });

  it("displays the profile page header", () => {
    cy.contains("Welcome to your Profile!").should("be.visible");
  });

  it("displays the user's email", () => {
    cy.contains("ta@ufl.edu").should("be.visible");
  });

  it("displays the Edit Profile button", () => {
    cy.contains("Edit Profile").should("be.visible");
  });

  it("displays the Log Out button", () => {
    cy.contains("Log Out").should("be.visible");
  });

  it("enters edit mode when Edit Profile is clicked", () => {
    cy.contains("Edit Profile").click();
    cy.contains("Save").should("be.visible");
    cy.contains("Cancel").should("be.visible");
  });

  it("shows a name input field in edit mode", () => {
    cy.contains("Edit Profile").click();
    cy.get(".profile-form-input").first().should("be.visible");
  });

  it("shows email as disabled in edit mode", () => {
    cy.contains("Edit Profile").click();
    cy.get(".profile-form-email-disabled").should("be.visible");
  });

  it("shows the Add Course button in edit mode", () => {
    cy.contains("Edit Profile").click();
    cy.contains("+ Add Course").should("be.visible");
  });

  it("cancels edit mode when Cancel is clicked", () => {
    cy.contains("Edit Profile").click();
    cy.contains("Cancel").click();
    cy.contains("Edit Profile").should("be.visible");
    cy.contains("Save").should("not.exist");
  });

  it("updates the username successfully", () => {
    cy.contains("Edit Profile").click();
    cy.get(".profile-form-input").first().clear().type("UpdatedTAName");
    cy.contains("Save").click();
    cy.contains("Profile updated successfully!").should("be.visible");
  });

  it("opens the Add Course modal in edit mode", () => {
    cy.contains("Edit Profile").click();
    cy.contains("+ Add Course").click();
    cy.contains("Add Course").should("be.visible");
  });

  it("logs out when Log Out is clicked", () => {
    cy.contains("Log Out").click();
    cy.url().should("include", "/login");
  });
});

describe("Toast Notifications", () => {
  it("shows error toast when joining a queue that is not open", () => {
    cy.visit("/login");
    cy.get('input[placeholder="Email"]').type("student@ufl.edu");
    cy.get('input[placeholder="Password"]').type("correctpassword");
    cy.get('button[type="submit"]').click();
    cy.url().should("include", "/student");
    cy.get(".course-dropdown").then(($dropdown) => {
      if ($dropdown.find("option").length > 1) {
        cy.get(".course-dropdown").select(1);
        cy.contains("Join Queue").click();
        cy.contains("The TA has not opened the queue yet").should("be.visible");
      }
    });
  });
});