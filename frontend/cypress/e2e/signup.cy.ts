/// <reference types="cypress" />

// E2E tests for the Signup page
describe('Signup Page', () => {
  beforeEach(() => {
    cy.visit('/signup');
  });

  it('should display the signup form', () => {
    cy.get('h1').should('contain.text', 'Create your account');
    cy.get('[data-testid="signup-name"]').should('be.visible');
    cy.get('[data-testid="signup-email"]').should('be.visible');
    cy.get('[data-testid="signup-password"]').should('be.visible');
    cy.get('[data-testid="signup-confirm-password"]').should('be.visible');
    cy.get('[data-testid="signup-submit"]').should('be.visible');
  });

  it('should fill in the name field', () => {
    cy.get('[data-testid="signup-name"]').type('Test User');
    cy.get('[data-testid="signup-name"]').should('have.value', 'Test User');
  });

  it('should fill in the email field', () => {
    cy.get('[data-testid="signup-email"]').type('test@example.com');
    cy.get('[data-testid="signup-email"]').should('have.value', 'test@example.com');
  });

  it('should fill in the full signup form', () => {
    cy.get('[data-testid="signup-name"]').type('Test User');
    cy.get('[data-testid="signup-email"]').type('test@example.com');
    cy.get('[data-testid="signup-password"]').type('password123');
    cy.get('[data-testid="signup-confirm-password"]').type('password123');
    cy.get('[data-testid="signup-submit"]').should('not.be.disabled');
  });

  it('should show error when passwords do not match', () => {
    cy.get('[data-testid="signup-name"]').type('Test User');
    cy.get('[data-testid="signup-email"]').type('test@example.com');
    cy.get('[data-testid="signup-password"]').type('password123');
    cy.get('[data-testid="signup-confirm-password"]').type('different456');
    cy.get('[data-testid="signup-submit"]').click();
    cy.get('.error').should('be.visible');
  });

  it('should navigate to login page via the log in link', () => {
    cy.get('[data-testid="signup-go-login"]').click();
    cy.url().should('include', '/login');
  });

  it('should have a back to home link', () => {
    cy.get('a.backLink').should('be.visible').and('contain.text', 'Back to home');
  });
});
