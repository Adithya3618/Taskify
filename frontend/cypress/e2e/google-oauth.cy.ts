/// <reference types="cypress" />

// ── Helpers ────────────────────────────────────────────────────────────────

function visitLogin() {
  cy.visit('/login');
}

function visitSignup() {
  cy.visit('/signup');
}

// ── Login page — Google OAuth UI (#70) ─────────────────────────────────────

describe('Google OAuth — Login page', () => {
  beforeEach(() => visitLogin());

  it('shows a "Sign in with Google" button on the login page', () => {
    cy.contains('button', 'Sign in with Google').should('be.visible');
  });

  it('Google button is enabled by default', () => {
    cy.contains('button', 'Sign in with Google').should('not.be.disabled');
  });

  it('shows a loading state on the Google button while the request is in flight', () => {
    cy.intercept('GET', '/api/auth/google/login', (req) => {
      req.on('response', (res) => { res.setDelay(3000); });
      req.reply({ statusCode: 200, body: '' });
    }).as('googleLogin');

    cy.contains('button', 'Sign in with Google').click();
    cy.contains('button', /signing in|loading/i).should('exist');
  });

  it('shows an error message below the Google button when the backend returns 503', () => {
    cy.intercept('GET', '/api/auth/google/login', {
      statusCode: 503,
      body: { error: 'OAuth not configured' },
    }).as('googleLoginFail');

    cy.contains('button', 'Sign in with Google').click();
    cy.wait('@googleLoginFail');

    // Error should appear below the Google button, not above the login form
    cy.contains('Google sign-in is not available right now').should('be.visible');
  });

  it('shows an error message below the Google button when the backend returns 500', () => {
    cy.intercept('GET', '/api/auth/google/login', {
      statusCode: 500,
      body: { error: 'Internal error' },
    }).as('googleLoginError');

    cy.contains('button', 'Sign in with Google').click();
    cy.wait('@googleLoginError');

    cy.contains('Google sign-in is not available right now').should('be.visible');
  });

  it('error message appears below the Google button, not above the email form', () => {
    cy.intercept('GET', '/api/auth/google/login', { statusCode: 503, body: {} });

    cy.contains('button', 'Sign in with Google').click();

    // The error for Google should NOT affect the main form's error slot
    cy.get('input[type="email"]').should('not.have.class', 'error');
    cy.contains('Google sign-in is not available right now').should('be.visible');
  });

  it('has a divider separating email login from Google login', () => {
    cy.contains(/or/i).should('be.visible');
  });
});

// ── Signup page — Google OAuth UI (#70) ────────────────────────────────────

describe('Google OAuth — Signup page', () => {
  beforeEach(() => visitSignup());

  it('shows a "Sign up with Google" button on the signup page', () => {
    cy.contains('button', 'Sign up with Google').should('be.visible');
  });

  it('Google button on signup is enabled by default', () => {
    cy.contains('button', 'Sign up with Google').should('not.be.disabled');
  });

  it('shows an error below the Google button on signup when backend returns 503', () => {
    cy.intercept('GET', '/api/auth/google/login', {
      statusCode: 503,
      body: { error: 'OAuth not configured' },
    }).as('googleSignupFail');

    cy.contains('button', 'Sign up with Google').click();
    cy.wait('@googleSignupFail');

    cy.contains('Google sign-in is not available right now').should('be.visible');
  });

  it('has a divider separating email signup from Google signup', () => {
    cy.contains(/or/i).should('be.visible');
  });
});

// ── Google callback page (#70) ─────────────────────────────────────────────

describe('Google OAuth — Callback page', () => {
  it('redirects to /boards after handling a token in the callback URL', () => {
    cy.visit('/auth/google/callback?token=fake-jwt&name=Alice&email=alice%40example.com&id=42');
    cy.url().should('include', '/boards');
    expect(localStorage.getItem('taskify.auth.token')).to.eq('fake-jwt');
  });
});
