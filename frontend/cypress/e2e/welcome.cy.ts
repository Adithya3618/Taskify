/// <reference types="cypress" />

// E2E tests for the Welcome (landing) page
describe('Welcome Page', () => {
  beforeEach(() => {
    cy.visit('/');
  });

  it('should load the welcome page and display the brand name', () => {
    cy.get('.brandName').should('be.visible').and('contain.text', 'askify');
  });

  it('should show the Login and Sign up buttons in the navbar', () => {
    cy.get('a.btn.navBtn').should('be.visible').and('contain.text', 'Login');
    cy.get('a.btn.btnPrimary').should('be.visible').and('contain.text', 'Sign up');
  });

  it('should navigate to /login when Login button is clicked', () => {
    cy.get('a.btn.navBtn').click();
    cy.url().should('include', '/login');
  });

  it('should navigate to /signup when Sign up button is clicked', () => {
    cy.visit('/');
    cy.get('a.btn.btnPrimary').click();
    cy.url().should('include', '/signup');
  });

  it('should display the hero heading', () => {
    cy.get('.hero h1').should('be.visible');
  });

  it('should toggle the theme when the theme toggle button is clicked', () => {
    cy.get('button.themeToggle').click();
    cy.get('html').should('have.attr', 'data-theme', 'light');
    cy.get('button.themeToggle').click();
    cy.get('html').should('have.attr', 'data-theme', 'dark');
  });
});
