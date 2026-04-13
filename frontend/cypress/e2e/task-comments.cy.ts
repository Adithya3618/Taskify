/// <reference types="cypress" />

import {
  TASK_ID,
  makeComment,
  makeTask,
  taskDescriptionWithMeta,
  visitBoard,
  visitPlanner,
} from '../support/board-stubs';

function dueTodayIso(): string {
  const d = new Date();
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${y}-${m}-${day}`;
}

function openBoardTaskDetails() {
  cy.get('.task-card .task-content').first().click();
  cy.get('.task-modal').should('be.visible');
  cy.contains('.commentSectionLabel', 'Comments').should('be.visible');
}

function openPlannerTaskDetails() {
  cy.get('#planner-nodue-toggle').click();
  cy.contains('.planner-task-title', 'Test task').click();
  cy.contains('.plannerCommentSectionLabel', 'Comments').should('be.visible');
}

describe('Task comments', () => {
  it('shows the board comment count on the task card', () => {
    visitBoard({
      commentsByTaskId: {
        [TASK_ID]: [
          makeComment(TASK_ID, 9101, 'Initial note'),
          makeComment(TASK_ID, 9102, 'Follow-up comment'),
        ],
      },
    });

    cy.get('.task-comment-meta').first().should('contain', '2');
  });

  it('shows the board empty state when a task has no comments', () => {
    visitBoard({ commentsByTaskId: { [TASK_ID]: [] } });
    openBoardTaskDetails();
    cy.contains('.memberEmptyState', 'No comments yet. Start the conversation below.').should('be.visible');
  });

  it('keeps the board Post button disabled until comment text is entered', () => {
    visitBoard({ commentsByTaskId: { [TASK_ID]: [] } });
    openBoardTaskDetails();
    cy.contains('.commentsComposer button', 'Post').should('be.disabled');
    cy.get('.commentsComposer textarea').type('Board note from Cypress');
    cy.contains('.commentsComposer button', 'Post').should('not.be.disabled');
  });

  it('shows edit and delete actions only for the board comment author', () => {
    visitBoard({
      commentsByTaskId: {
        [TASK_ID]: [
          makeComment(TASK_ID, 9101, 'My note', 'E2E User', { userId: 'e2e@test.com' }),
          makeComment(TASK_ID, 9102, 'Someone else noted this', 'Casey Doe', { userId: 'member-2' }),
        ],
      },
    });

    openBoardTaskDetails();
    cy.contains('.commentCard', 'My note').within(() => {
      cy.contains('button', 'Edit').should('be.visible');
      cy.contains('button', 'Delete').should('be.visible');
    });
    cy.contains('.commentCard', 'Someone else noted this').within(() => {
      cy.contains('button', 'Edit').should('not.exist');
      cy.contains('button', 'Delete').should('not.exist');
    });
  });

  it('cancels a board inline comment edit without saving changes', () => {
    visitBoard({
      commentsByTaskId: {
        [TASK_ID]: [makeComment(TASK_ID, 9101, 'Editable note')],
      },
    });

    openBoardTaskDetails();
    cy.contains('.commentCard', 'Editable note').contains('button', 'Edit').click();
    cy.get('.commentEditForm textarea').clear().type('Changed but cancelled');
    cy.contains('.commentEditActions button', 'Cancel').click();
    cy.contains('.commentCard', 'Editable note').should('be.visible');
    cy.contains('.commentCard', 'Changed but cancelled').should('not.exist');
  });

  it('auto-scrolls to the latest board comment when task details open', () => {
    visitBoard({
      commentsByTaskId: {
        [TASK_ID]: [
          makeComment(TASK_ID, 9101, 'First note'),
          makeComment(TASK_ID, 9102, 'Second note'),
          makeComment(TASK_ID, 9103, 'Latest note'),
        ],
      },
    });

    cy.window().then((win) => {
      cy.stub(win.HTMLElement.prototype, 'scrollIntoView').callsFake(() => {}).as('boardScrollIntoView');
    });
    openBoardTaskDetails();
    cy.get('@boardScrollIntoView').should('have.been.called');
  });

  it('shows the planner comment count on the task chip', () => {
    const due = dueTodayIso();
    visitPlanner({
      tasksByStageId: {
        100: [makeTask(TASK_ID, 100, 'Test task', taskDescriptionWithMeta('', { due }))],
      },
      commentsByTaskId: {
        [TASK_ID]: [
          makeComment(TASK_ID, 9201, 'Planner note one'),
          makeComment(TASK_ID, 9202, 'Planner note two'),
        ],
      },
    });

    cy.contains('.planner-grid .planner-task-title', 'Test task')
      .closest('.planner-task-chip')
      .find('.planner-task-comment-meta')
      .first()
      .should('contain', '2');
  });

  it('keeps the planner Post button disabled until text is entered', () => {
    visitPlanner({ commentsByTaskId: { [TASK_ID]: [] } });
    openPlannerTaskDetails();
    cy.contains('.plannerCommentsComposer button', 'Post').should('be.disabled');
    cy.get('.plannerCommentsComposer textarea').type('Planner note from Cypress');
    cy.contains('.plannerCommentsComposer button', 'Post').should('not.be.disabled');
  });

  it('shows author-only actions in the planner comment list', () => {
    visitPlanner({
      commentsByTaskId: {
        [TASK_ID]: [
          makeComment(TASK_ID, 9201, 'My planner note', 'E2E User', { userId: 'e2e@test.com' }),
          makeComment(TASK_ID, 9202, 'Other user note', 'Casey Doe', { userId: 'member-2' }),
        ],
      },
    });

    openPlannerTaskDetails();
    cy.contains('.plannerCommentCard', 'My planner note').within(() => {
      cy.contains('button', 'Edit').should('be.visible');
      cy.contains('button', 'Delete').should('be.visible');
    });
    cy.contains('.plannerCommentCard', 'Other user note').within(() => {
      cy.contains('button', 'Edit').should('not.exist');
      cy.contains('button', 'Delete').should('not.exist');
    });
  });

  it('keeps the planner comment when deletion is cancelled', () => {
    visitPlanner({
      commentsByTaskId: {
        [TASK_ID]: [makeComment(TASK_ID, 9201, 'Keep this planner comment')],
      },
    });

    openPlannerTaskDetails();
    cy.on('window:confirm', () => false);
    cy.contains('.plannerCommentCard', 'Keep this planner comment').contains('button', 'Delete').click();
    cy.contains('.plannerCommentCard', 'Keep this planner comment').should('be.visible');
  });

  it('posts a planner comment from the Post button and updates the count', () => {
    const due = dueTodayIso();
    visitPlanner({
      commentsByTaskId: {
        [TASK_ID]: [makeComment(TASK_ID, 9201, 'Existing planner comment')],
      },
      tasksByStageId: {
        100: [makeTask(TASK_ID, 100, 'Test task', taskDescriptionWithMeta('', { due }))],
      },
    });

    openPlannerTaskDetails();
    cy.get('.plannerCommentsComposer textarea').type('New planner comment');
    cy.contains('.plannerCommentsComposer button', 'Post').click();
    cy.wait('@createComment').its('request.body.content').should('eq', 'New planner comment');
    cy.contains('.plannerCommentCard', 'New planner comment').should('be.visible');
    cy.get('.btn-close-modal').click();
    cy.contains('.planner-grid .planner-task-title', 'Test task')
      .closest('.planner-task-chip')
      .find('.planner-task-comment-meta')
      .first()
      .should('contain', '2');
  });
});
