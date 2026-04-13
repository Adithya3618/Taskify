/// <reference types="cypress" />

import {
  TASK_ID,
  makeSubtask,
  visitBoard,
  visitPlanner,
} from '../support/board-stubs';

function openBoardTaskDetails() {
  cy.get('.task-card .task-content').first().click();
  cy.get('.task-modal').should('be.visible');
  cy.contains('.subtaskSectionLabel', 'Checklist').should('be.visible');
}

function openPlannerTaskDetails() {
  cy.get('#planner-nodue-toggle').click();
  cy.contains('.planner-task-title', 'Test task').click();
  cy.contains('.plannerSubtaskSectionLabel', 'Checklist').should('be.visible');
}

describe('Task checklists', () => {
  it('shows the board empty checklist state when no subtasks exist', () => {
    visitBoard({ subtasksByTaskId: { [TASK_ID]: [] } });
    openBoardTaskDetails();
    cy.contains('.memberEmptyState', 'No checklist items yet. Add one above to get started.').should('be.visible');
  });

  it('keeps the board Add button disabled until a checklist title is entered', () => {
    visitBoard({ subtasksByTaskId: { [TASK_ID]: [] } });
    openBoardTaskDetails();
    cy.contains('.subtaskComposer button', 'Add').should('be.disabled');
    cy.get('.subtaskComposer input').type('Write release summary');
    cy.contains('.subtaskComposer button', 'Add').should('not.be.disabled');
  });

  it('adds a board checklist item when the Add button is clicked', () => {
    visitBoard({
      subtasksByTaskId: {
        [TASK_ID]: [makeSubtask(TASK_ID, 7101, 'Draft outline')],
      },
    });

    openBoardTaskDetails();
    cy.get('.subtaskComposer input').type('Review analytics');
    cy.contains('.subtaskComposer button', 'Add').click();
    cy.wait('@createSubtask').its('request.body.title').should('eq', 'Review analytics');
    cy.contains('.subtaskItemTitle', 'Review analytics').should('be.visible');
  });

  it('updates the board task progress bar fill after a checklist item is checked', () => {
    visitBoard({
      subtasksByTaskId: {
        [TASK_ID]: [
          makeSubtask(TASK_ID, 7101, 'Draft outline'),
          makeSubtask(TASK_ID, 7102, 'Review copy'),
        ],
      },
    });

    openBoardTaskDetails();
    cy.contains('.subtaskItem', 'Draft outline').find('input[type="checkbox"]').check({ force: true });
    cy.wait('@updateSubtask');
    cy.get('.task-subtask-progress-fill').first().should('have.attr', 'style').and('include', '50%');
    cy.get('.task-subtask-progress-text').first().should('contain', '1/2 done');
  });

  it('shows the board delete confirmation with the selected checklist title', () => {
    visitBoard({
      subtasksByTaskId: {
        [TASK_ID]: [makeSubtask(TASK_ID, 7102, 'Review copy')],
      },
    });

    openBoardTaskDetails();
    cy.contains('.subtaskItem', 'Review copy').contains('button', 'Delete').click();
    cy.contains('.delete-confirm-modal h3', 'Delete checklist item?').should('be.visible');
    cy.contains('.delete-confirm-task-name', 'Review copy').should('be.visible');
  });

  it('shows the planner empty checklist state when no subtasks exist', () => {
    visitPlanner({ subtasksByTaskId: { [TASK_ID]: [] } });
    openPlannerTaskDetails();
    cy.contains('.plannerCommentEmptyState', 'No checklist items yet. Add one above to get started.').should('be.visible');
  });

  it('keeps the planner Add button disabled until a checklist title is entered', () => {
    visitPlanner({ subtasksByTaskId: { [TASK_ID]: [] } });
    openPlannerTaskDetails();
    cy.contains('.plannerSubtaskComposer button', 'Add').should('be.disabled');
    cy.get('.plannerSubtaskComposer input').type('Planner checklist item');
    cy.contains('.plannerSubtaskComposer button', 'Add').should('not.be.disabled');
  });

  it('adds a planner checklist item from the Add button', () => {
    visitPlanner({
      subtasksByTaskId: {
        [TASK_ID]: [makeSubtask(TASK_ID, 7101, 'First step')],
      },
    });

    openPlannerTaskDetails();
    cy.get('.plannerSubtaskComposer input').type('Second step');
    cy.contains('.plannerSubtaskComposer button', 'Add').click();
    cy.wait('@createSubtask').its('request.body.title').should('eq', 'Second step');
    cy.contains('.plannerSubtaskItemTitle', 'Second step').should('be.visible');
  });

  it('deletes a planner checklist item from the task modal', () => {
    visitPlanner({
      subtasksByTaskId: {
        [TASK_ID]: [makeSubtask(TASK_ID, 7101, 'Disposable step')],
      },
    });

    openPlannerTaskDetails();
    cy.contains('.plannerSubtaskItem', 'Disposable step').contains('button', 'Delete').click();
    cy.wait('@deleteSubtask');
    cy.contains('.plannerSubtaskItem', 'Disposable step').should('not.exist');
  });
});
