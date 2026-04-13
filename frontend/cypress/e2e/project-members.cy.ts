/// <reference types="cypress" />

import {
  makeKnownUser,
  makeProjectMember,
  visitBoard,
} from '../support/board-stubs';

function openMembersSettings() {
  cy.contains('button.btn-ghost', 'Settings').click();
  cy.contains('h3', 'Project settings').should('be.visible');
  cy.contains('.settingsTab', 'Members').click();
  cy.contains('.membersPanelHead h4', 'Current members').should('be.visible');
}

describe('Project members', () => {
  it('opens the settings modal on the Members tab from the top bar button', () => {
    visitBoard();
    cy.contains('button.btn-ghost', 'Settings').click();
    cy.contains('h3', 'Project settings').should('be.visible');
    cy.contains('.settingsTab.settingsTabActive', 'Members').should('be.visible');
    cy.contains('.membersPanelHead', 'Current members').should('be.visible');
  });

  it('keeps the Add Member button disabled until a search value is entered', () => {
    visitBoard();
    openMembersSettings();
    cy.contains('.memberSearchField button', 'Add Member').should('be.disabled');
    cy.get('.memberSearchField input').type('new.person@test.com');
    cy.contains('.memberSearchField button', 'Add Member').should('not.be.disabled');
  });

  it('shows matching known users while typing in the member search field', () => {
    visitBoard({
      knownUsers: [
        makeKnownUser('member-7', 'Taylor Reese', 'taylor@test.com'),
        makeKnownUser('member-8', 'Morgan Hall', 'morgan@test.com'),
      ],
    });

    openMembersSettings();
    cy.get('.memberSearchField input').type('tay');
    cy.contains('.memberSearchResult', 'Taylor Reese').should('be.visible');
    cy.contains('.memberSearchResult', 'Morgan Hall').should('not.exist');
  });

  it('adds a member when a search result is clicked', () => {
    visitBoard({
      knownUsers: [
        makeKnownUser('member-9', 'Taylor Reese', 'taylor@test.com'),
      ],
    });

    openMembersSettings();
    cy.get('.memberSearchField input').type('Taylor');
    cy.contains('.memberSearchResult', 'Taylor Reese').click();
    cy.wait('@addProjectMember').its('request.body').should('deep.include', {
      user_id: 'member-9',
      email: 'taylor@test.com',
    });
    cy.contains('.memberFeedbackSuccess', 'Taylor Reese added to the project.').should('be.visible');
    cy.contains('.memberRow', 'Taylor Reese').should('be.visible');
  });

  it('adds a member from a direct email entry with the Add Member button', () => {
    visitBoard();
    openMembersSettings();
    cy.get('.memberSearchField input').type('new.person@test.com');
    cy.contains('.memberSearchField button', 'Add Member').click();
    cy.wait('@addProjectMember').its('request.body.email').should('eq', 'new.person@test.com');
    cy.contains('.memberRow', 'new.person@test.com').should('be.visible');
  });

  it('shows a clear error when trying to add a duplicate member', () => {
    visitBoard();
    openMembersSettings();
    cy.get('.memberSearchField input').type('casey@test.com');
    cy.contains('.memberSearchField button', 'Add Member').click();
    cy.contains('.memberFeedbackError', 'This member is already on the project.').should('be.visible');
  });

  it('hides remove buttons for non-owners', () => {
    visitBoard({
      owners: { '1': 'owner@test.com' },
      sessionUser: { name: 'Jordan', email: 'member@test.com', id: 'member-3' },
      projectMembers: [
        makeProjectMember(1, 'owner-1', 'Owner User', 'owner@test.com', 'owner'),
        makeProjectMember(1, 'member-3', 'Jordan', 'member@test.com', 'member'),
      ],
      cachedProjectMembers: {
        1: [
          makeProjectMember(1, 'owner-1', 'Owner User', 'owner@test.com', 'owner'),
          makeProjectMember(1, 'member-3', 'Jordan', 'member@test.com', 'member'),
        ],
      },
    });

    openMembersSettings();
    cy.get('.memberRemoveBtn').should('not.exist');
  });

  it('keeps a member in the list when removal is cancelled', () => {
    visitBoard();
    openMembersSettings();
    cy.on('window:confirm', (message) => {
      expect(message).to.contain('Casey Doe');
      return false;
    });
    cy.contains('.memberRow', 'Casey Doe').contains('button', 'Remove').click();
    cy.contains('.memberRow', 'Casey Doe').should('be.visible');
    cy.get('.undoToast').should('not.exist');
  });

  it('shows an undo toast after removing a member and restores the member when Undo is clicked', () => {
    cy.clock();
    visitBoard();
    openMembersSettings();
    cy.on('window:confirm', () => true);
    cy.contains('.memberRow', 'Casey Doe').contains('button', 'Remove').click();
    cy.contains('.undoToast', 'Casey Doe removed.').should('be.visible');
    cy.contains('.memberRow', 'Casey Doe').should('not.exist');
    cy.contains('.undoToast button', 'Undo').click();
    cy.contains('.memberFeedbackSuccess', 'Casey Doe restored.').should('be.visible');
    cy.contains('.memberRow', 'Casey Doe').should('be.visible');
    cy.tick(5000);
    cy.get('@removeProjectMember.all').should('have.length', 0);
  });

  it('finalizes the member removal after the undo window expires', () => {
    cy.clock();
    visitBoard();
    openMembersSettings();
    cy.on('window:confirm', () => true);
    cy.contains('.memberRow', 'Casey Doe').contains('button', 'Remove').click();
    cy.contains('.undoToast', 'Casey Doe removed.').should('be.visible');
    cy.tick(5000);
    cy.wait('@removeProjectMember');
    cy.get('.undoToast').should('not.exist');
    cy.contains('.memberRow', 'Casey Doe').should('not.exist');
    cy.contains('.memberFeedbackSuccess', 'Casey Doe removed from the project.').should('be.visible');
  });
});
