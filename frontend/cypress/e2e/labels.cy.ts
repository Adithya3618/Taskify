/// <reference types="cypress" />

import {
  visitBoard,
  makeTask,
  taskDescriptionWithMeta,
  STAGE_ID,
  isoNow,
} from '../support/board-stubs';

// ── Label Manager (#72) ────────────────────────────────────────────────────

describe('Labels — Label Manager modal', () => {
  beforeEach(() => visitBoard());

  it('shows a Labels button in the board topbar', () => {
    cy.contains('button', 'Labels').should('be.visible');
  });

  it('opens the Label Manager modal when the Labels button is clicked', () => {
    cy.contains('button', 'Labels').click();
    cy.contains('h3', 'Label Manager').should('be.visible');
  });

  it('closes the Label Manager modal with the X button', () => {
    cy.contains('button', 'Labels').click();
    cy.contains('h3', 'Label Manager').should('be.visible');
    cy.get('.labelManagerModal .btn-close-modal').click();
    cy.contains('h3', 'Label Manager').should('not.exist');
  });

  it('shows an input field for the new label name', () => {
    cy.contains('button', 'Labels').click();
    cy.get('.labelManagerModal input[placeholder]').should('be.visible');
  });

  it('creates a new label and shows it in the label list', () => {
    cy.contains('button', 'Labels').click();
    cy.get('.labelManagerModal input[placeholder]').type('Bug');
    cy.contains('.labelManagerModal button', 'Add').click();
    cy.get('.labelManagerModal').contains('Bug').should('be.visible');
  });

  it('clears the input after adding a label', () => {
    cy.contains('button', 'Labels').click();
    cy.get('.labelManagerModal input[placeholder]').type('Feature');
    cy.contains('.labelManagerModal button', 'Add').click();
    cy.get('.labelManagerModal input[placeholder]').should('have.value', '');
  });

  it('can delete a label from the manager', () => {
    cy.contains('button', 'Labels').click();
    cy.get('.labelManagerModal input[placeholder]').type('ToDelete');
    cy.contains('.labelManagerModal button', 'Add').click();
    cy.contains('.labelManagerModal', 'ToDelete').should('be.visible');
    cy.contains('.labelManagerModal .labelRow', 'ToDelete').find('button').click();
    cy.contains('.labelManagerModal', 'ToDelete').should('not.exist');
  });
});

// ── Labels on task cards (#72) ─────────────────────────────────────────────

describe('Labels — assign in task detail modal', () => {
  beforeEach(() => visitBoard());

  it('shows a label section in the task detail modal', () => {
    cy.get('.task-card .task-content').first().click();
    cy.get('.task-modal').contains(/label/i).should('be.visible');
  });

  it('shows label chips for selection in the task detail modal', () => {
    // First create a label
    cy.contains('button', 'Labels').click();
    cy.get('.labelManagerModal input[placeholder]').type('Urgent');
    cy.contains('.labelManagerModal button', 'Add').click();
    cy.get('.labelManagerModal .btn-close-modal').click();

    // Open task modal and check for the label chip
    cy.get('.task-card .task-content').first().click();
    cy.get('.task-modal').contains('Urgent').should('be.visible');
  });

  it('shows label picker in the add task form when expanded', () => {
    cy.get('.add-task-form')
      .first()
      .within(() => {
        cy.get('input[placeholder="+ Add a task..."]').type('Test');
        cy.contains('button', '+ Add details').click();
      });
    cy.contains(/label/i).should('be.visible');
  });
});

// ── Label filter in filter panel (#72) ─────────────────────────────────────

describe('Labels — filter panel', () => {
  beforeEach(() => visitBoard());

  it('shows a Label section in the filter panel', () => {
    cy.get('[data-testid="board-filter-toggle"]').click();
    cy.contains('.filterGroupLabel', /label/i).should('be.visible');
  });

  it('All labels chip is active by default in the filter panel', () => {
    cy.get('[data-testid="board-filter-toggle"]').click();
    cy.contains('.filterGroupLabel', /label/i)
      .parent()
      .contains('button', 'All')
      .should('have.class', 'active');
  });
});

// ── Priority Escalation (#71) ──────────────────────────────────────────────

describe('Priority escalation — dynamic display based on deadline', () => {
  it('shows "Urgent" priority badge for a task with Low priority that is overdue', () => {
    cy.clock(new Date(2026, 5, 15, 12, 0, 0).getTime(), ['Date']);
    visitBoard({
      tasksByStageId: {
        100: [
          makeTask(
            8001,
            STAGE_ID,
            'Overdue Low Task',
            taskDescriptionWithMeta('Desc', { priority: 'Low', due: '2026-06-10' })
          ),
        ],
        101: [],
      },
    });
    cy.contains('.task-card', 'Overdue Low Task')
      .find('.priority-badge, .chip-urgent, [class*="urgent"]')
      .should('exist');
  });

  it('shows "High" effective priority for a Medium-priority task due today', () => {
    cy.clock(new Date(2026, 5, 15, 12, 0, 0).getTime(), ['Date']);
    visitBoard({
      tasksByStageId: {
        100: [
          makeTask(
            8002,
            STAGE_ID,
            'Due Today Medium',
            taskDescriptionWithMeta('Desc', { priority: 'Medium', due: '6/15/2026' })
          ),
        ],
        101: [],
      },
    });
    cy.contains('.task-card', 'Due Today Medium')
      .find('.priority-badge, [class*="high"], [class*="urgent"]')
      .should('exist');
  });

  it('shows the due-overdue class on the due date label for an overdue task', () => {
    cy.clock(new Date(2026, 5, 15, 12, 0, 0).getTime(), ['Date']);
    visitBoard({
      tasksByStageId: {
        100: [
          makeTask(
            8003,
            STAGE_ID,
            'Overdue task',
            taskDescriptionWithMeta('Desc', { due: '2026-06-01' })
          ),
        ],
        101: [],
      },
    });
    cy.contains('.task-card', 'Overdue task')
      .find('.due-overdue, [class*="overdue"]')
      .should('exist');
  });

  it('shows the due-today class on the due date label for a task due today', () => {
    cy.clock(new Date(2026, 5, 15, 12, 0, 0).getTime(), ['Date']);
    visitBoard({
      tasksByStageId: {
        100: [
          makeTask(
            8004,
            STAGE_ID,
            'Due today task',
            taskDescriptionWithMeta('Desc', { due: '6/15/2026' })
          ),
        ],
        101: [],
      },
    });
    cy.contains('.task-card', 'Due today task')
      .find('.due-today, [class*="today"]')
      .should('exist');
  });

  it('task detail modal priority selector reflects dynamic priority for overdue task', () => {
    cy.clock(new Date(2026, 5, 15, 12, 0, 0).getTime(), ['Date']);
    visitBoard({
      tasksByStageId: {
        100: [
          makeTask(
            8005,
            STAGE_ID,
            'Priority check task',
            taskDescriptionWithMeta('Desc', { priority: 'Low', due: '2026-06-01' })
          ),
        ],
        101: [],
      },
    });
    cy.contains('.task-card', 'Priority check task').find('.task-content').click();
    cy.get('.task-modal').should('be.visible');
    // Detail modal should show escalated priority label
    cy.get('.task-modal').contains(/urgent|high/i).should('be.visible');
  });
});
