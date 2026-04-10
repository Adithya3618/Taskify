/// <reference types="cypress" />

import { visitBoard, makeTask, taskDescriptionWithMeta, STAGE_ID } from '../support/board-stubs';

const NOTIF_KEY = 'taskify.notifications';

/** Seeds a notification into localStorage before the page loads. */
function seedNotification(
  win: Cypress.AUTWindow,
  overrides: {
    id?: string;
    type?: string;
    message?: string;
    is_read?: boolean;
    created_at?: string;
    link?: string;
  } = {}
) {
  const notif = {
    id: overrides.id ?? `notif-${Date.now()}-abc`,
    type: overrides.type ?? 'task_assigned',
    message: overrides.message ?? 'You have been assigned a task',
    is_read: overrides.is_read ?? false,
    created_at: overrides.created_at ?? new Date().toISOString(),
    link: overrides.link,
  };
  const existing = JSON.parse(win.localStorage.getItem(NOTIF_KEY) ?? '[]');
  win.localStorage.setItem(NOTIF_KEY, JSON.stringify([notif, ...existing]));
}

// ── Notification Bell on Board (#73) ───────────────────────────────────────

describe('Notifications — bell button in board topbar', () => {
  it('shows the notification bell button in the board topbar', () => {
    visitBoard();
    cy.get('app-notification-bell .bellBtn').should('be.visible');
  });

  it('shows no unread badge when there are no notifications', () => {
    visitBoard();
    cy.get('app-notification-bell .badge').should('not.exist');
  });

  it('shows an unread badge with count when there is an unread notification', () => {
    visitBoard({
      tasksByStageId: { 100: [], 101: [] },
      skipTaskCardAssert: true,
    });
    // Seed a notification before visiting — use onBeforeLoad approach via custom visit
    cy.visit('/board/1', {
      onBeforeLoad(win) {
        win.localStorage.setItem('taskify.auth.token', 'e2e-token');
        win.localStorage.setItem(
          'taskify.auth.session',
          JSON.stringify({ name: 'E2E User', email: 'e2e@test.com' })
        );
        win.localStorage.setItem('taskify.board.owners', JSON.stringify({ '1': 'e2e@test.com' }));
        seedNotification(win, { message: 'Task assigned to you', is_read: false });
      },
    });
    cy.get('.board-content', { timeout: 15000 }).should('be.visible');
    cy.get('app-notification-bell .badge').should('be.visible').and('contain', '1');
  });

  it('caps the badge at 9+ when there are more than 9 unread notifications', () => {
    cy.visit('/board/1', {
      onBeforeLoad(win) {
        win.localStorage.setItem('taskify.auth.token', 'e2e-token');
        win.localStorage.setItem(
          'taskify.auth.session',
          JSON.stringify({ name: 'E2E User', email: 'e2e@test.com' })
        );
        win.localStorage.setItem('taskify.board.owners', JSON.stringify({ '1': 'e2e@test.com' }));
        for (let i = 0; i < 12; i++) {
          seedNotification(win, { id: `notif-${i}`, message: `Notification ${i}`, is_read: false });
        }
      },
    });
    cy.get('.board-content', { timeout: 15000 }).should('be.visible');
    cy.get('app-notification-bell .badge').should('contain', '9+');
  });
});

// ── Notification dropdown panel (#73) ─────────────────────────────────────

describe('Notifications — dropdown panel', () => {
  beforeEach(() => {
    cy.intercept('GET', '**/api/projects/1', { id: 1, name: 'E2E Board 1', description: '', created_at: new Date().toISOString(), updated_at: new Date().toISOString() });
    cy.intercept('GET', '**/api/projects/1/stages', []);
    cy.intercept('GET', '**/api/projects', []);
  });

  it('opens the notification panel when the bell is clicked', () => {
    visitBoard({ tasksByStageId: { 100: [], 101: [] }, skipTaskCardAssert: true });
    cy.get('app-notification-bell .bellBtn').click();
    cy.get('app-notification-bell .notifPanel').should('be.visible');
  });

  it('closes the panel when clicking the bell again', () => {
    visitBoard({ tasksByStageId: { 100: [], 101: [] }, skipTaskCardAssert: true });
    cy.get('app-notification-bell .bellBtn').click();
    cy.get('app-notification-bell .notifPanel').should('be.visible');
    cy.get('app-notification-bell .bellBtn').click();
    cy.get('app-notification-bell .notifPanel').should('not.exist');
  });

  it('closes the panel when clicking the backdrop', () => {
    visitBoard({ tasksByStageId: { 100: [], 101: [] }, skipTaskCardAssert: true });
    cy.get('app-notification-bell .bellBtn').click();
    cy.get('app-notification-bell .notifBackdrop').click({ force: true });
    cy.get('app-notification-bell .notifPanel').should('not.exist');
  });

  it('shows "You\'re all caught up!" when there are no notifications', () => {
    visitBoard({ tasksByStageId: { 100: [], 101: [] }, skipTaskCardAssert: true });
    cy.get('app-notification-bell .bellBtn').click();
    cy.get('app-notification-bell').contains("You're all caught up!").should('be.visible');
  });

  it('shows the notification message in the panel', () => {
    cy.visit('/board/1', {
      onBeforeLoad(win) {
        win.localStorage.setItem('taskify.auth.token', 'e2e-token');
        win.localStorage.setItem(
          'taskify.auth.session',
          JSON.stringify({ name: 'E2E User', email: 'e2e@test.com' })
        );
        win.localStorage.setItem('taskify.board.owners', JSON.stringify({ '1': 'e2e@test.com' }));
        seedNotification(win, { message: '"Design task" is due today.', type: 'deadline_reminder' });
      },
    });
    cy.get('.board-content', { timeout: 15000 }).should('be.visible');
    cy.get('app-notification-bell .bellBtn').click();
    cy.get('app-notification-bell').contains('"Design task" is due today.').should('be.visible');
  });

  it('shows "Mark all read" button when there are unread notifications', () => {
    cy.visit('/board/1', {
      onBeforeLoad(win) {
        win.localStorage.setItem('taskify.auth.token', 'e2e-token');
        win.localStorage.setItem(
          'taskify.auth.session',
          JSON.stringify({ name: 'E2E User', email: 'e2e@test.com' })
        );
        win.localStorage.setItem('taskify.board.owners', JSON.stringify({ '1': 'e2e@test.com' }));
        seedNotification(win, { message: 'Unread notification' });
      },
    });
    cy.get('.board-content', { timeout: 15000 }).should('be.visible');
    cy.get('app-notification-bell .bellBtn').click();
    cy.get('app-notification-bell').contains('Mark all read').should('be.visible');
  });

  it('removes the unread badge after clicking Mark all read', () => {
    cy.visit('/board/1', {
      onBeforeLoad(win) {
        win.localStorage.setItem('taskify.auth.token', 'e2e-token');
        win.localStorage.setItem(
          'taskify.auth.session',
          JSON.stringify({ name: 'E2E User', email: 'e2e@test.com' })
        );
        win.localStorage.setItem('taskify.board.owners', JSON.stringify({ '1': 'e2e@test.com' }));
        seedNotification(win, { message: 'A notification to clear' });
      },
    });
    cy.get('.board-content', { timeout: 15000 }).should('be.visible');
    cy.get('app-notification-bell .badge').should('be.visible');
    cy.get('app-notification-bell .bellBtn').click();
    cy.get('app-notification-bell').contains('Mark all read').click();
    cy.get('app-notification-bell .badge').should('not.exist');
  });

  it('shows clock emoji for deadline_reminder type notifications', () => {
    cy.visit('/board/1', {
      onBeforeLoad(win) {
        win.localStorage.setItem('taskify.auth.token', 'e2e-token');
        win.localStorage.setItem(
          'taskify.auth.session',
          JSON.stringify({ name: 'E2E User', email: 'e2e@test.com' })
        );
        win.localStorage.setItem('taskify.board.owners', JSON.stringify({ '1': 'e2e@test.com' }));
        seedNotification(win, { type: 'deadline_reminder', message: '"Task X" is overdue!' });
      },
    });
    cy.get('.board-content', { timeout: 15000 }).should('be.visible');
    cy.get('app-notification-bell .bellBtn').click();
    cy.get('app-notification-bell .notifIcon').first().should('contain', '⏰');
  });
});

// ── Deadline-triggered notifications (#73) ─────────────────────────────────

describe('Notifications — deadline reminders auto-generated on board load', () => {
  it('auto-generates an overdue notification when board loads with an overdue task', () => {
    cy.clock(new Date(2026, 5, 15, 12, 0, 0).getTime(), ['Date']);
    visitBoard({
      tasksByStageId: {
        100: [
          makeTask(
            9001,
            STAGE_ID,
            'Past due task',
            taskDescriptionWithMeta('Desc', { due: '2026-06-10' })
          ),
        ],
        101: [],
      },
    });
    cy.get('app-notification-bell .badge').should('be.visible');
    cy.get('app-notification-bell .bellBtn').click();
    cy.get('app-notification-bell').contains('overdue').should('be.visible');
  });

  it('auto-generates a "due today" notification on board load', () => {
    cy.clock(new Date(2026, 5, 15, 12, 0, 0).getTime(), ['Date']);
    visitBoard({
      tasksByStageId: {
        100: [
          makeTask(
            9002,
            STAGE_ID,
            'Due today task',
            taskDescriptionWithMeta('Desc', { due: '6/15/2026' })
          ),
        ],
        101: [],
      },
    });
    cy.get('app-notification-bell .bellBtn').click();
    cy.get('app-notification-bell').contains('due today').should('be.visible');
  });

  it('does NOT add a notification for a task due next week', () => {
    cy.clock(new Date(2026, 5, 15, 12, 0, 0).getTime(), ['Date']);
    visitBoard({
      tasksByStageId: {
        100: [
          makeTask(
            9003,
            STAGE_ID,
            'Future task',
            taskDescriptionWithMeta('Desc', { due: '2026-07-01' })
          ),
        ],
        101: [],
      },
    });
    cy.get('app-notification-bell .badge').should('not.exist');
  });
});

// ── Notification bell on Home page (#73) ───────────────────────────────────

describe('Notifications — bell on Home/Boards page', () => {
  beforeEach(() => {
    cy.intercept('GET', '**/api/projects', []).as('getProjects');
    cy.visit('/boards', {
      onBeforeLoad(win) {
        win.localStorage.setItem('taskify.auth.token', 'e2e-token');
        win.localStorage.setItem(
          'taskify.auth.session',
          JSON.stringify({ name: 'E2E User', email: 'e2e@test.com' })
        );
        win.localStorage.setItem('taskify.board.owners', JSON.stringify({}));
      },
    });
  });

  it('shows the notification bell on the home/boards page', () => {
    cy.get('app-notification-bell').should('exist');
    cy.get('app-notification-bell .bellBtn').should('be.visible');
  });

  it('opens the notification panel from the home page', () => {
    cy.get('app-notification-bell .bellBtn').click();
    cy.get('app-notification-bell .notifPanel').should('be.visible');
  });

  it('shows empty state on home page when no notifications exist', () => {
    cy.get('app-notification-bell .bellBtn').click();
    cy.get('app-notification-bell').contains("You're all caught up!").should('be.visible');
  });

  it('shows unread badge on home page when notification was seeded', () => {
    cy.intercept('GET', '**/api/projects', []).as('getProjects2');
    cy.visit('/boards', {
      onBeforeLoad(win) {
        win.localStorage.setItem('taskify.auth.token', 'e2e-token');
        win.localStorage.setItem(
          'taskify.auth.session',
          JSON.stringify({ name: 'E2E User', email: 'e2e@test.com' })
        );
        win.localStorage.setItem('taskify.board.owners', JSON.stringify({}));
        seedNotification(win, { message: 'A home page notification' });
      },
    });
    cy.get('app-notification-bell .badge').should('be.visible');
  });
});
