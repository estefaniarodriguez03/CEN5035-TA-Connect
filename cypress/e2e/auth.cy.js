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

  it("displays the student dashboard", () => {
    cy.contains("TA Connect").should("be.visible");
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

  it("displays the TA dashboard", () => {
    cy.contains("TA Connect").should("be.visible");
  });

  it("navigates to the queue when Start Office Hours Live Queue is clicked", () => {
    cy.contains("Start Office Hours Live Queue").click();
    cy.contains("Queue Management").should("be.visible");
  });
});