/// <reference types="cypress" />

import {
  PROJECT_ID,
  STAGE_ID,
  TASK_ID,
  makeTask,
  makeSubtask,
  taskDescriptionWithMeta,
  visitBoard,
  visitPlanner,
} from '../support/board-stubs';

/** Local YYYY-MM-DD for today (planner uses the same date-key format as card meta). */
function dueTodayIso(): string {
  const d = new Date();
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${y}-${m}-${day}`;
}

describe('Planner — load', () => {
  it('shows the month calendar and weekday header', () => {
    visitPlanner();
    cy.get('.planner-weekdays').should('be.visible');
    cy.get('.planner-calendar-header-nav .btn-today').should('be.visible');
  });

  it('does not show the loading state after data loads', () => {
    visitPlanner();
    cy.get('.planner-loading').should('not.exist');
  });
});

describe('Planner — board ↔ planner tabs', () => {
  it('navigates from board to planner without full reload', () => {
    visitBoard();
    cy.url().should('not.include', '/planner');
    cy.get('nav.viewTabs a.viewTab').contains('Planner').click();
    cy.url().should('include', `/board/${PROJECT_ID}/planner`);
    cy.get('.planner-main').should('be.visible');
  });

  it('navigates from planner back to the kanban board', () => {
    visitPlanner();
    cy.get('nav.viewTabs a.viewTab').contains('Board').click();
    cy.url().should('match', new RegExp(`/board/${PROJECT_ID}($|\\?)`));
    cy.get('.board-content').should('be.visible');
  });
});

describe('Planner — sidebar lists', () => {
  it('opens the No due date panel and lists tasks without a due date', () => {
    visitPlanner();
    cy.get('#planner-nodue-toggle').click();
    cy.get('#planner-nodue-toggle').should('have.attr', 'aria-expanded', 'true');
    cy.contains('.planner-no-due-list .planner-task-title', 'Test task').should('be.visible');
  });

  it('opens the Scheduled panel', () => {
    visitPlanner();
    cy.get('#planner-scheduled-toggle').click();
    cy.get('#planner-scheduled-toggle').should('have.attr', 'aria-expanded', 'true');
  });
});

describe('Planner — task on calendar', () => {
  it('shows a task with today’s due date on the calendar grid', () => {
    const due = dueTodayIso();
    visitPlanner({
      tasksByStageId: {
        [STAGE_ID]: [
          makeTask(
            5001,
            STAGE_ID,
            'Due today on planner',
            taskDescriptionWithMeta('', { due, priority: 'High' })
          ),
        ],
      },
    });
    cy.contains('.planner-grid .planner-task-title', 'Due today on planner').should('be.visible');
  });

  it('toggles completion on a calendar task chip', () => {
    const due = dueTodayIso();
    visitPlanner({
      tasksByStageId: {
        [STAGE_ID]: [
          makeTask(
            5002,
            STAGE_ID,
            'Completable',
            taskDescriptionWithMeta('', { due })
          ),
        ],
      },
    });
    cy.contains('.planner-task-title', 'Completable')
      .closest('.planner-task-chip')
      .find('input.board-checkbox')
      .check({ force: true });
    cy.contains('.planner-task-title', 'Completable')
      .closest('.planner-task-chip')
      .should('have.class', 'planner-task-chip--done');
  });
});

describe('Planner — month navigation', () => {
  it('changes the visible month when using ‹ and resets with Today', () => {
    visitPlanner();
    cy.get('#planner-cal-title')
      .invoke('text')
      .then((before) => {
        cy.get('.planner-calendar-header-nav .btn-nav-month').first().click();
        cy.get('#planner-cal-title').invoke('text').should('not.eq', before.trim());
      });
    cy.get('.planner-calendar-header-nav .btn-today').click();
    cy.get('#planner-cal-title').should('contain', String(new Date().getFullYear()));
  });

  it('opens the month/year picker from the title button', () => {
    visitPlanner();
    cy.get('#planner-cal-title').click();
    cy.contains('h3', 'Go to month').should('be.visible');
    cy.get('#monthYearInput').should('exist');
    cy.contains('.month-picker-modal button', 'Cancel').click();
    cy.contains('h3', 'Go to month').should('not.exist');
  });
});

describe('Planner — add task from calendar', () => {
  it('opens Add task when clicking an empty day in the current month', () => {
    visitPlanner();
    // Default stub has no dated tasks — any in-month cell is empty of chips.
    cy.get('.planner-grid .planner-day:not(.planner-day--muted)').first().click();
    cy.get('#addTaskTitle').contains('Add task');
    cy.contains('.btn-close-modal', '×').click();
    cy.get('#addTaskTitle').should('not.exist');
  });
});

describe('Planner â€” checklist subtasks', () => {
  it('shows and updates checklist items inside the planner task modal', () => {
    visitPlanner({
      subtasksByTaskId: {
        [TASK_ID]: [
          makeSubtask(TASK_ID, 9101, 'First step'),
          makeSubtask(TASK_ID, 9102, 'Second step', true, 1),
        ],
      },
    });

    cy.get('#planner-nodue-toggle').click();
    cy.contains('.planner-task-title', 'Test task').click();
    cy.get('.plannerSubtaskItemTitle').eq(0).should('contain', 'First step');
    cy.contains('.plannerSubtaskItem', 'First step').find('input[type="checkbox"]').check({ force: true });
    cy.wait('@updateSubtask').its('request.body.is_completed').should('eq', true);
    cy.get('.plannerSubtaskSectionMeta').should('contain', '2/2 done');
  });
});
