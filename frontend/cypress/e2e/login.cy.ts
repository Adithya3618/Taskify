/// <reference types="cypress" />

// E2E tests for the Login page
describe('Login Page', () => {
  beforeEach(() => {
    cy.visit('/login');
  });

  it('should display the login form', () => {
    cy.get('h1').should('contain.text', 'Log in to Taskify');
    cy.get('input[name="email"]').should('be.visible');
    cy.get('input[name="password"]').should('be.visible');
    cy.get('button[type="submit"]').should('be.visible');
  });

  it('should fill in the email field', () => {
    cy.get('input[name="email"]').type('test@example.com');
    cy.get('input[name="email"]').should('have.value', 'test@example.com');
  });

  it('should fill in the password field', () => {
    cy.get('input[name="password"]').type('password123');
    cy.get('input[name="password"]').should('have.value', 'password123');
  });

  it('should show an error when submitting empty credentials', () => {
    cy.get('button[type="submit"]').click();
    cy.get('.error').should('be.visible');
  });

  it('should navigate to signup page via the sign up link', () => {
    cy.get('a[routerlink="/signup"]').click();
    cy.url().should('include', '/signup');
  });

  it('should toggle password visibility', () => {
    cy.get('input[name="password"]').should('have.attr', 'type', 'password');
    cy.get('.eyeBtn').first().click();
    cy.get('input[name="password"]').should('have.attr', 'type', 'text');
  });
});