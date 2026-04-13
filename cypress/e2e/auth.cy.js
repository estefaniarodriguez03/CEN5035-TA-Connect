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

  // 1. Dashboard loads correctly
  it("displays the student dashboard with welcome message", () => {
    cy.contains("Welcome Back").should("be.visible");
    cy.contains("TA Connect").should("be.visible");
  });

  // Dropdown shows course options
  it("displays all four course options in the dropdown", () => {
    cy.get(".course-dropdown option").should("have.length", 4);
  });

  // Weekly schedule shows all five days
  it("displays all five days in the weekly schedule", () => {
    cy.contains("Monday").should("be.visible");
    cy.contains("Tuesday").should("be.visible");
    cy.contains("Wednesday").should("be.visible");
    cy.contains("Thursday").should("be.visible");
    cy.contains("Friday").should("be.visible");
  });

  // Notification icon is visible
  it("displays the notification icon in the navbar", () => {
    cy.get(".notification-icon").should("be.visible");
  });

  // 18. TA names are visible in Today's TA Hours
  it("displays TA names in Today's TA Hours", () => {
    cy.contains("Estefania Rodriguez").should("be.visible");
    cy.contains("Sara Waters").should("be.visible");
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

  // Dashboard loads correctly
  it("displays the TA dashboard with welcome message", () => {
    cy.contains("Welcome Back").should("be.visible");
    cy.contains("TA Connect").should("be.visible");
  });

  // Navbar tabs are visible
  it("displays the correct navigation tabs", () => {
    cy.contains("Dashboard").should("be.visible");
    cy.contains("My Office Hours").should("be.visible");
    cy.contains("Queue").should("be.visible");
  });

  // Today's schedule is visible
  it("displays today's office hour schedule cards", () => {
    cy.contains("11:00 AM - 1:00 PM").should("be.visible");
    cy.contains("3:30 PM - 4:30 PM").should("be.visible");
  });

  // Selecting an office hour enables the start queue button
  it("enables the start queue button after selecting an office hour", () => {
    cy.contains("11:00 AM - 1:00 PM").click();
    cy.contains("Start Office Hours Live Queue").should("not.be.disabled");
  });

  // Selecting a different office hour also enables the button
  it("enables the start queue button when the second office hour is selected", () => {
    cy.contains("3:30 PM - 4:30 PM").click();
    cy.contains("Start Office Hours Live Queue").should("not.be.disabled");
  });

  // Stats grid is visible
  it("displays all stat cards on the dashboard", () => {
    cy.contains("Students Helped Today").should("be.visible");
    cy.contains("Avg Wait Time").should("be.visible");
    cy.contains("Current Queue Length").should("be.visible");
    cy.contains("Longest Wait Time").should("be.visible");
    cy.contains("Most Common Topic").should("be.visible");
    cy.contains("Avg Session Duration").should("be.visible");
  });

  // Weekly office hours sidebar is visible
  it("displays the weekly office hours sidebar", () => {
    cy.contains("This Week's Office Hours").should("be.visible");
    cy.contains("Monday").should("be.visible");
    cy.contains("Wednesday").should("be.visible");
    cy.contains("Friday").should("be.visible");
  });

  // Live queue view loads after starting queue
  it("shows the live queue view after starting the queue", () => {
    cy.contains("11:00 AM - 1:00 PM").click();
    cy.contains("Start Office Hours Live Queue").click();
    cy.contains("Live Queue").should("be.visible");
  });

  // Queue controls are visible in live queue view
  it("displays queue controls in the live queue view", () => {
    cy.contains("11:00 AM - 1:00 PM").click();
    cy.contains("Start Office Hours Live Queue").click();
    cy.contains("Pause Queue").should("be.visible");
    cy.contains("Close Queue").should("be.visible");
  });

  // Queue stats are visible in live queue view
  it("displays queue stats in the live queue view", () => {
    cy.contains("11:00 AM - 1:00 PM").click();
    cy.contains("Start Office Hours Live Queue").click();
    cy.contains("Total in Queue").should("be.visible");
    cy.contains("Longest Wait").should("be.visible");
    cy.contains("Avg Wait Time").should("be.visible");
  });

  // Pause queue toggles to resume
  it("toggles to Resume Queue after clicking Pause Queue", () => {
    cy.contains("11:00 AM - 1:00 PM").click();
    cy.contains("Start Office Hours Live Queue").click();
    cy.contains("Pause Queue").click();
    cy.contains("Resume Queue").should("be.visible");
  });

  // Resume queue toggles back to pause
  it("toggles back to Pause Queue after clicking Resume Queue", () => {
    cy.contains("11:00 AM - 1:00 PM").click();
    cy.contains("Start Office Hours Live Queue").click();
    cy.contains("Pause Queue").click();
    cy.contains("Resume Queue").click();
    cy.contains("Pause Queue").should("be.visible");
  });

  // Close queue returns to dashboard
  it("returns to the dashboard view after closing the queue", () => {
    cy.contains("11:00 AM - 1:00 PM").click();
    cy.contains("Start Office Hours Live Queue").click();
    cy.contains("Close Queue").click();
  });
});

describe("TA Dashboard — My Office Hours Tab", () => {
  beforeEach(() => {
    cy.visit("/login");
    cy.get('input[placeholder="Email"]').type("ta@ufl.edu");
    cy.get('input[placeholder="Password"]').type("correctpassword");
    cy.get('button[type="submit"]').click();
    cy.url().should("include", "/ta");
  });

  it("navigates to My Office Hours tab when clicked", () => {
    cy.contains("My Office Hours").click();
    cy.contains("Manage your office hour schedule and availability").should("be.visible");
  });


  it("displays the Add Office Hours button on the My Office Hours tab", () => {
    cy.contains("My Office Hours").click();
    cy.contains("+ Add Office Hours").should("be.visible");
  });

  it("opens the Add Office Hours modal when button is clicked", () => {
    cy.contains("My Office Hours").click();
    cy.contains("+ Add Office Hours").click();
    cy.contains("Add Office Hours").should("be.visible");
    cy.get('input[type="time"]').should("have.length", 2);
  });

  it("adds a new office hour and displays it in the schedule", () => {
    cy.contains("My Office Hours").click();
    cy.contains("+ Add Office Hours").click();

    cy.get("select.oh-modal-input").select("Tuesday");
    cy.get('input[type="time"]').first().type("10:00");
    cy.get('input[type="time"]').last().type("11:00");
    cy.get('input[placeholder="e.g., COP3530 - Data Structures"]').type("COP3530 - Data Structures");
    cy.contains("button", "Add Schedule").click();

    cy.contains("10:00 AM – 11:00 AM").should("be.visible");
  });

  it("shows a success toast after adding office hours", () => {
    cy.contains("My Office Hours").click();
    cy.contains("+ Add Office Hours").click();

    cy.get('input[type="time"]').first().type("14:00");
    cy.get('input[type="time"]').last().type("15:00");
    cy.get('input[placeholder="e.g., COP3530 - Data Structures"]').type("COP4600 - Operating Systems");
    cy.contains("button", "Add Schedule").click();

    cy.contains("Office hours added successfully").should("be.visible");
  });

  it("shows an error toast when required fields are missing", () => {
    cy.contains("My Office Hours").click();
    cy.contains("+ Add Office Hours").click();
    cy.contains("button", "Add Schedule").click();
    cy.contains("Please fill in all fields").should("be.visible");
  });

  it("opens the edit modal pre-filled when Edit is clicked", () => {
    cy.contains("My Office Hours").click();
    cy.contains("Edit").first().click();
    cy.contains("Edit Office Hours").should("be.visible");
    cy.get('input[type="time"]').first().should("not.have.value", "");
  });

  it("updates an office hour and shows success toast", () => {
    cy.contains("My Office Hours").click();
    cy.contains("Edit").first().click();

    cy.get('input[type="time"]').last().clear().type("14:00");
    cy.contains("button", "Update Schedule").click();

    cy.contains("Office hours updated successfully").should("be.visible");
  });

  it("deletes an office hour and shows success toast", () => {
    cy.contains("My Office Hours").click();
    cy.contains("Delete").first().click();
    cy.contains("Office hours deleted").should("be.visible");
  });

  it("shows empty state for days with no office hours", () => {
    cy.contains("My Office Hours").click();
    cy.contains("No office hours scheduled").should("be.visible");
  });

  it("navigates back to Dashboard tab from My Office Hours", () => {
    cy.contains("My Office Hours").click();
    cy.contains("Manage your office hour schedule and availability").should("be.visible");
    cy.contains("Dashboard").click();
    cy.contains("Welcome Back").should("be.visible");
    cy.contains("Manage your office hour schedule and availability").should("not.exist");
  });
});

describe("Toast Notifications", () => {
  it("shows toast notifications on the student dashboard", () => {
    cy.visit("/login");
    cy.get('input[placeholder="Email"]').type("student@ufl.edu");
    cy.get('input[placeholder="Password"]').type("correctpassword");
    cy.get('button[type="submit"]').click();
    cy.url().should("include", "/student");

    // Select a course with no active queue to trigger the error toast
    cy.get(".course-dropdown").select(
      "COP4020 – Programming Languages – Raghav Nanjappan (2:00 PM - 4:00 PM)"
    );
    cy.contains("Join Queue").click();
    cy.contains("The TA has not opened the queue for this office hours yet").should("be.visible");
  });
});