/// <reference types="cypress" />

// E2E tests for the Login page
describe('Login Page', () => {
  beforeEach(() => {
    cy.visit('/login');
  });

  it('should display the login form', () => {
    cy.get('h1').should('contain.text', 'Log in to Taskify');
    cy.get('[data-testid="login-email"]').should('be.visible');
    cy.get('[data-testid="login-password"]').should('be.visible');
    cy.get('[data-testid="login-submit"]').should('be.visible');
  });

  it('should fill in the email field', () => {
    cy.get('[data-testid="login-email"]').type('test@example.com');
    cy.get('[data-testid="login-email"]').should('have.value', 'test@example.com');
  });

  it('should fill in the password field', () => {
    cy.get('[data-testid="login-password"]').type('password123');
    cy.get('[data-testid="login-password"]').should('have.value', 'password123');
  });

  it('should show an error when submitting empty credentials', () => {
    cy.get('[data-testid="login-submit"]').click();
    cy.get('.error').should('be.visible');
  });

  it('should navigate to signup page via the sign up link', () => {
    cy.get('[data-testid="login-go-signup"]').click();
    cy.url().should('include', '/signup');
  });

  it('should toggle password visibility', () => {
    cy.get('[data-testid="login-password"]').should('have.attr', 'type', 'password');
    cy.get('[data-testid="login-toggle-password"]').click();
    cy.get('[data-testid="login-password"]').should('have.attr', 'type', 'text');
  });
});