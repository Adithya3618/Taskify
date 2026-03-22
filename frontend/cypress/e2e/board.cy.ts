/// <reference types="cypress" />

/**
 * E2E tests for the board page: task checkboxes, column collapse, and due-date inputs.
 * These are Cypress end-to-end tests (they exercise the real app in a browser), not Angular unit tests.
 */

const PROJECT_ID = 1;
const STAGE_ID = 100;
const TASK_ID = 5001;

const isoNow = () => new Date().toISOString();

function seedBoardAuth(win: Cypress.AUTWindow) {
  win.localStorage.setItem('taskify.auth.token', 'e2e-token');
  win.localStorage.setItem(
    'taskify.auth.session',
    JSON.stringify({ name: 'E2E User', email: 'e2e@test.com' })
  );
  win.localStorage.setItem(
    'taskify.board.owners',
    JSON.stringify({ [String(PROJECT_ID)]: 'e2e@test.com' })
  );
}

function stubBoardRestApi() {
  const project = {
    id: PROJECT_ID,
    name: 'E2E Board',
    description: '',
    created_at: isoNow(),
    updated_at: isoNow(),
  };
  const stage = {
    id: STAGE_ID,
    project_id: PROJECT_ID,
    name: 'To Do',
    position: 0,
    created_at: isoNow(),
    updated_at: isoNow(),
  };
  const task = {
    id: TASK_ID,
    stage_id: STAGE_ID,
    title: 'Test task',
    description: '',
    position: 0,
    completed: false,
    created_at: isoNow(),
    updated_at: isoNow(),
  };

  // One GET handler so `/api/projects` never shadows `/api/projects/1`.
  cy.intercept('GET', '**/api/**', (req) => {
    const path = new URL(req.url).pathname.replace(/\/$/, '');
    const base = `/api/projects/${PROJECT_ID}`;
    if (path === `${base}/stages/${STAGE_ID}/tasks`) {
      req.reply([task]);
      return;
    }
    if (path === `${base}/stages`) {
      req.reply([stage]);
      return;
    }
    if (path === base) {
      req.reply(project);
      return;
    }
    if (path === '/api/projects') {
      req.reply([]);
      return;
    }
    req.continue();
  });

  cy.intercept('PUT', `**/api/tasks/${TASK_ID}`, (req) => {
    req.reply({
      statusCode: 200,
      body: {
        ...task,
        title: req.body.title,
        description: req.body.description,
        position: req.body.position,
        updated_at: isoNow(),
      },
    });
  }).as('updateTask');
}

function visitStubbedBoard() {
  stubBoardRestApi();
  cy.visit(`/board/${PROJECT_ID}`, {
    onBeforeLoad(win) {
      seedBoardAuth(win);
    },
  });
  cy.get('.board-content', { timeout: 15000 }).should('be.visible');
  cy.get('.task-card', { timeout: 10000 }).should('have.length', 1);
}

describe('Board — task checkbox', () => {
  beforeEach(() => {
    visitStubbedBoard();
  });

  it('applies completed styling when the checkbox is checked', () => {
    cy.get('.task-card')
      .first()
      .find('.task-check input[type="checkbox"]')
      .should('not.be.checked')
      .check({ force: true });

    cy.get('.task-card').first().should('have.class', 'task-completed');
    cy.get('.task-card .task-check input[type="checkbox"]').first().should('be.checked');
  });

  it('removes completed styling when the checkbox is unchecked', () => {
    cy.get('.task-card .task-check input[type="checkbox"]').first().check({ force: true });
    cy.get('.task-card').first().should('have.class', 'task-completed');

    cy.get('.task-card .task-check input[type="checkbox"]').first().uncheck({ force: true });
    cy.get('.task-card').first().should('not.have.class', 'task-completed');
    cy.get('.task-card .task-check input[type="checkbox"]').first().should('not.be.checked');
  });
});

describe('Board — column collapse', () => {
  beforeEach(() => {
    visitStubbedBoard();
  });

  it('collapses a list to the narrow strip and expands it again', () => {
    cy.get('.column:not(.add-column)')
      .first()
      .as('firstList');

    cy.get('@firstList').find('.btn-collapse-column').click();

    cy.get('@firstList').should('have.class', 'column--collapsed');
    cy.get('@firstList').find('.column-collapsed-strip').should('be.visible');
    cy.get('@firstList').find('.column-collapsed-title').should('contain', 'To Do');

    cy.get('@firstList').find('.column-collapsed-strip').click();

    cy.get('@firstList').should('not.have.class', 'column--collapsed');
    cy.get('@firstList').find('.tasks-list').should('be.visible');
  });
});

describe('Board — due date (calendar) inputs', () => {
  beforeEach(() => {
    visitStubbedBoard();
  });

  it('opens task details and shows the due date field', () => {
    cy.get('.task-card .task-content').first().click();
    cy.get('.task-modal').should('be.visible');
    cy.contains('label', 'Due Date').should('be.visible');
    cy.get('.task-modal input[type="date"]').should('exist');
  });

  it('persists due date via Save and sends it in the task update request', () => {
    cy.get('.task-card .task-content').first().click();
    cy.get('.task-modal input[type="date"]').invoke('val', '2026-06-15').trigger('input', { force: true });
    cy.get('.task-modal input[type="date"]').invoke('val', '2026-06-15').trigger('change', { force: true });

    cy.contains('.task-modal button', 'Save').click();

    cy.wait('@updateTask').its('request.body.description').should('include', 'due:2026-06-15');
    cy.get('.task-modal').should('not.exist');
  });

  it('exposes a date input when adding a task with details', () => {
    cy.get('.add-task-form')
      .first()
      .within(() => {
        cy.get('input[placeholder="+ Add a task..."]').type('New item');
        cy.contains('button', '+ Add details').click();
        cy.get('input[type="date"][title="Due date"]').should('be.visible');
      });
  });
});
