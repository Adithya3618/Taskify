/// <reference types="cypress" />

// E2E tests for the Signup page
describe('Signup Page', () => {
  beforeEach(() => {
    cy.visit('/signup');
  });

  it('should display the signup form', () => {
    cy.get('h1').should('contain.text', 'Create your account');
    cy.get('input[name="name"]').should('be.visible');
    cy.get('input[name="email"]').should('be.visible');
    cy.get('input[name="password"]').should('be.visible');
    cy.get('input[name="confirmPassword"]').should('be.visible');
    cy.get('button[type="submit"]').should('be.visible');
  });

  it('should fill in the name field', () => {
    cy.get('input[name="name"]').type('Test User');
    cy.get('input[name="name"]').should('have.value', 'Test User');
  });

  it('should fill in the email field', () => {
    cy.get('input[name="email"]').type('test@example.com');
    cy.get('input[name="email"]').should('have.value', 'test@example.com');
  });

  it('should fill in the full signup form', () => {
    cy.get('input[name="name"]').type('Test User');
    cy.get('input[name="email"]').type('test@example.com');
    cy.get('input[name="password"]').type('password123');
    cy.get('input[name="confirmPassword"]').type('password123');
    cy.get('button[type="submit"]').should('not.be.disabled');
  });

  it('should show error when passwords do not match', () => {
    cy.get('input[name="name"]').type('Test User');
    cy.get('input[name="email"]').type('test@example.com');
    cy.get('input[name="password"]').type('password123');
    cy.get('input[name="confirmPassword"]').type('different456');
    cy.get('button[type="submit"]').click();
    cy.get('.error').should('be.visible');
  });

  it('should navigate to login page via the log in link', () => {
    cy.get('a[routerlink="/login"]').click();
    cy.url().should('include', '/login');
  });

  it('should have a back to home link', () => {
    cy.get('a.backLink').should('be.visible').and('contain.text', 'Back to home');
  });
});
