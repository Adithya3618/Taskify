/// <reference types="cypress" />

/**
 * Reset browser storage so board tests start from a predictable state.
 * Auth and board-owner keys are re-applied in each test via visit `onBeforeLoad`.
 */
beforeEach(() => {
  cy.clearAllCookies();
  cy.clearAllLocalStorage();
  cy.clearAllSessionStorage();
});
