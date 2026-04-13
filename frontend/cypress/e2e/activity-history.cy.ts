/// <reference types="cypress" />

import {
  makeActivity,
  makeProjectMember,
  visitBoard,
} from '../support/board-stubs';

const ownerMember = makeProjectMember(1, 'owner-1', 'Adithya', 'owner@test.com', 'owner');
const teammateMember = makeProjectMember(1, 'member-2', 'Casey Doe', 'casey@test.com', 'member');

function visitBoardWithActivity() {
  visitBoard({
    owners: { '1': 'owner@test.com' },
    sessionUser: { name: 'Adithya', email: 'owner@test.com', id: 'owner-1' },
    projectMembers: [ownerMember, teammateMember],
    cachedProjectMembers: { 1: [ownerMember, teammateMember] },
    activityByProjectId: {
      1: [
        makeActivity(1, 1, "Adithya moved 'Design login' to Done", {
          userId: 'owner-1',
          userName: 'Adithya',
          createdAt: '2026-04-12T14:00:00.000Z',
        }),
        makeActivity(1, 2, "Casey added label 'Urgent' to 'Fix bugs'", {
          userId: 'member-2',
          userName: 'Casey Doe',
          action: 'label_assigned',
          entityType: 'label',
          createdAt: '2026-04-11T11:00:00.000Z',
        }),
        makeActivity(1, 3, "Adithya commented on 'Launch prep'", {
          userId: 'owner-1',
          userName: 'Adithya',
          action: 'comment_added',
          entityType: 'comment',
          createdAt: '2026-04-10T09:00:00.000Z',
        }),
        makeActivity(1, 4, "Casey joined the project", {
          userId: 'member-2',
          userName: 'Casey Doe',
          action: 'member_joined',
          entityType: 'member',
          createdAt: '2026-04-09T09:00:00.000Z',
        }),
        makeActivity(1, 5, "Adithya updated 'Retro notes'", {
          userId: 'owner-1',
          userName: 'Adithya',
          action: 'task_updated',
          createdAt: '2026-04-08T09:00:00.000Z',
        }),
        makeActivity(1, 6, "Casey deleted 'Old draft'", {
          userId: 'member-2',
          userName: 'Casey Doe',
          action: 'task_deleted',
          createdAt: '2026-04-07T09:00:00.000Z',
        }),
        makeActivity(1, 7, "Adithya added Casey to the board", {
          userId: 'owner-1',
          userName: 'Adithya',
          action: 'member_added',
          entityType: 'member',
          createdAt: '2026-04-06T09:00:00.000Z',
        }),
        makeActivity(1, 8, "Casey added a comment to 'Roadmap'", {
          userId: 'member-2',
          userName: 'Casey Doe',
          action: 'comment_added',
          entityType: 'comment',
          createdAt: '2026-04-05T09:00:00.000Z',
        }),
        makeActivity(1, 9, "Adithya removed label 'Blocked'", {
          userId: 'owner-1',
          userName: 'Adithya',
          action: 'label_removed',
          entityType: 'label',
          createdAt: '2026-04-04T09:00:00.000Z',
        }),
        makeActivity(1, 10, "Casey moved 'Homepage QA' to Doing", {
          userId: 'member-2',
          userName: 'Casey Doe',
          createdAt: '2026-04-03T09:00:00.000Z',
        }),
        makeActivity(1, 11, "Adithya updated 'Board polish'", {
          userId: 'owner-1',
          userName: 'Adithya',
          action: 'task_updated',
          createdAt: '2026-04-02T09:00:00.000Z',
        }),
        makeActivity(1, 12, "Casey created 'Write release notes'", {
          userId: 'member-2',
          userName: 'Casey Doe',
          action: 'task_created',
          createdAt: '2026-04-01T09:00:00.000Z',
        }),
        makeActivity(1, 13, "Adithya archived 'Old roadmap'", {
          userId: 'owner-1',
          userName: 'Adithya',
          action: 'task_deleted',
          createdAt: '2026-03-31T09:00:00.000Z',
        }),
      ],
    },
  });
}

function openActivityTab() {
  cy.contains('button.btn-ghost', 'Settings').click();
  cy.contains('.settingsTab', 'Activity').click();
  cy.contains('.activityPanelHead h4', 'Project activity').should('be.visible');
}

describe('Activity history', () => {
  it('opens the Activity tab and shows the project activity heading', () => {
    visitBoardWithActivity();
    openActivityTab();
    cy.contains('.activityCount', '13 entries').should('be.visible');
  });

  it('renders activity icons and relative timestamps for entries', () => {
    visitBoardWithActivity();
    openActivityTab();
    cy.get('.activityItem').first().find('.activityIcon').should('have.attr', 'data-action');
    cy.get('.activityItem').first().find('.activityContent span').invoke('text').should('not.be.empty');
  });

  it('lists project members in the activity member filter dropdown', () => {
    visitBoardWithActivity();
    openActivityTab();
    cy.get('#activity-member-filter option').then((options) => {
      const labels = [...options].map((option) => option.textContent?.trim());
      expect(labels).to.include('Adithya');
      expect(labels).to.include('Casey Doe');
    });
  });

  it('updates the activity list when the member filter changes', () => {
    visitBoardWithActivity();
    openActivityTab();
    cy.get('#activity-member-filter').select('member-2');
    cy.get('.activityList .activityItem').should('have.length', 6);
    cy.get('.activityList').should('contain', 'Casey added label');
    cy.get('.activityList').should('not.contain', "Adithya moved 'Design login' to Done");
  });

  it('clears the date and member filters when Reset is clicked', () => {
    visitBoardWithActivity();
    openActivityTab();
    cy.get('#activity-member-filter').select('member-2');
    cy.get('#activity-date-from')
      .invoke('val', '2026-04-05')
      .trigger('input', { force: true })
      .trigger('change', { force: true });
    cy.get('#activity-date-to')
      .invoke('val', '2026-04-09')
      .trigger('input', { force: true })
      .trigger('change', { force: true });
    cy.contains('.activityToolbar button', 'Reset').click();
    cy.get('#activity-member-filter').should('have.value', '');
    cy.get('#activity-date-from').should('have.value', '');
    cy.get('#activity-date-to').should('have.value', '');
  });

  it('loads more activity entries when Load more is clicked', () => {
    visitBoardWithActivity();
    openActivityTab();
    cy.get('.activityList .activityItem').should('have.length', 12);
    cy.contains('.activityMore button', 'Load more').click();
    cy.get('.activityList .activityItem').should('have.length', 13);
    cy.get('.activityList').should('contain', "Adithya archived 'Old roadmap'");
  });

  it('shows the empty state when the project has no activity', () => {
    visitBoard({
      owners: { '1': 'owner@test.com' },
      sessionUser: { name: 'Adithya', email: 'owner@test.com', id: 'owner-1' },
      projectMembers: [ownerMember, teammateMember],
      cachedProjectMembers: { 1: [ownerMember, teammateMember] },
      activityByProjectId: { 1: [] },
    });
    openActivityTab();
    cy.contains('.activityEmptyState strong', 'No activity logged yet.').should('be.visible');
  });
});
