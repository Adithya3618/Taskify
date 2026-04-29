/// <reference types="cypress" />

import {
  makeActivity,
  makeComment,
  makeProjectMember,
  STAGE_ID,
  STAGE_2_ID,
  TASK_ID,
  makeTask,
  makeSubtask,
  taskDescriptionWithMeta,
  visitBoard,
  isoNow,
} from '../support/board-stubs';

function openFilterPanel() {
  cy.get('[data-testid="board-filter-toggle"]').click();
  cy.get('.filterPanel').should('be.visible');
}

function clickFilterChip(groupLabel: string, chipText: string) {
  cy.contains('.filterGroupLabel', groupLabel)
    .parent()
    .find('.filterChips')
    .contains('button', chipText)
    .click();
}

// ── Checkboxes (list + modal) ─────────────────────────────────────────────

describe('Board — task checkbox (list)', () => {
  beforeEach(() => visitBoard());

  it('applies completed styling when the checkbox is checked', () => {
    cy.get('.task-card')
      .first()
      .find('.task-check input[type="checkbox"]')
      .should('not.be.checked')
      .check({ force: true });
    cy.get('.task-card').first().should('have.class', 'task-completed');
  });

  it('removes completed styling when the checkbox is unchecked', () => {
    cy.get('.task-card .task-check input[type="checkbox"]').first().check({ force: true });
    cy.get('.task-card').first().should('have.class', 'task-completed');
    cy.get('.task-card .task-check input[type="checkbox"]').first().uncheck({ force: true });
    cy.get('.task-card').first().should('not.have.class', 'task-completed');
  });

  it('keeps task completed after page reload', () => {
    cy.get('.task-card .task-check input[type="checkbox"]').first().check({ force: true });
    cy.reload();
    cy.get('.board-content', { timeout: 15000 }).should('be.visible');
    cy.get('.task-card').first().should('have.class', 'task-completed');
    cy.get('.task-card .task-check input[type="checkbox"]').first().should('be.checked');
  });
});

describe('Board — task checkbox (detail modal)', () => {
  beforeEach(() => visitBoard());

  it('opens the task modal with the Edit button', () => {
    cy.get('.task-card .btn-edit-task').first().click();
    cy.get('.task-modal').should('be.visible');
    cy.contains('h3', 'Task Details').should('be.visible');
  });

  it('marks complete from the modal and saves', () => {
    cy.get('.task-card .task-content').first().click();
    cy.get('.task-modal-done-row input[type="checkbox"]').check({ force: true });
    cy.contains('.task-modal button', 'Save').click();
    cy.wait('@updateTask');
    cy.get('.task-modal').should('not.exist');
    cy.get('.task-card .task-check input[type="checkbox"]').first().should('be.checked');
    cy.get('.task-card').first().should('have.class', 'task-completed');
  });
});

// ── Column collapse ─────────────────────────────────────────────────────────

describe('Board — column collapse', () => {
  beforeEach(() => visitBoard());

  it('collapses a list and expands it again (strip click)', () => {
    cy.get('.column:not(.add-column)').first().as('col');
    cy.get('@col').find('.btn-collapse-column').click();
    cy.get('@col').should('have.class', 'column--collapsed');
    cy.get('@col').find('.column-collapsed-strip').should('be.visible');
    cy.get('@col').find('.column-collapsed-strip').click();
    cy.get('@col').should('not.have.class', 'column--collapsed');
    cy.get('@col').find('.tasks-list').should('be.visible');
  });

  it('second column can be collapsed with the same button', () => {
    cy.get('.column:not(.add-column)').eq(1).as('col2');
    cy.get('@col2').find('.btn-collapse-column').click();
    cy.get('@col2').should('have.class', 'column--collapsed');
    cy.get('@col2').find('.column-collapsed-title').should('contain', 'Doing');
  });

  it('stays collapsed after reload', () => {
    cy.get('.column:not(.add-column)').first().find('.btn-collapse-column').click();
    cy.get('.column:not(.add-column)').first().should('have.class', 'column--collapsed');
    cy.reload();
    cy.get('.board-content', { timeout: 15000 }).should('be.visible');
    cy.get('.column:not(.add-column)').first().should('have.class', 'column--collapsed');
    cy.get('.column:not(.add-column)').first().find('.column-collapsed-strip').should('be.visible');
  });
});

// ── Due date / calendar ─────────────────────────────────────────────────────

describe('Board — due date (calendar)', () => {
  beforeEach(() => visitBoard());

  it('shows due date field in task details', () => {
    cy.get('.task-card .task-content').first().click();
    cy.contains('label', 'Due Date').should('be.visible');
    cy.get('.task-modal input[type="date"]').should('exist');
  });

  it('saves due date and sends it on update', () => {
    cy.get('.task-card .task-content').first().click();
    cy.get('.task-modal input[type="date"]')
      .invoke('val', '2026-06-15')
      .trigger('input', { force: true })
      .trigger('change', { force: true });
    cy.contains('.task-modal button', 'Save').click();
    cy.wait('@updateTask').its('request.body.description').should('include', 'due:2026-06-15');
    cy.get('.task-modal').should('not.exist');
  });

  it('shows due date input when adding details to a new task', () => {
    cy.get('.add-task-form')
      .first()
      .within(() => {
        cy.get('input[placeholder="+ Add a task..."]').type('New item');
        cy.contains('button', '+ Add details').click();
        cy.get('input[type="date"][title="Due date"]').should('be.visible');
      });
  });

  it('creates a task with due date via Add (API receives description)', () => {
    cy.get('.add-task-form')
      .first()
      .within(() => {
        cy.get('input[placeholder="+ Add a task..."]').type('Dated task');
        cy.contains('button', '+ Add details').click();
        cy.get('input[type="date"][title="Due date"]')
          .invoke('val', '2027-01-20')
          .trigger('input', { force: true })
          .trigger('change', { force: true });
        cy.contains('button', 'Add').click();
      });
    cy.wait('@createTask').its('request.body.description').should('include', 'due:2027-01-20');
    cy.contains('.task-card .task-title', 'Dated task').should('be.visible');
  });
});

// ── Filter panel ───────────────────────────────────────────────────────────

describe('Board — filter panel', () => {
  beforeEach(() => visitBoard());

  it('opens and closes when clicking Filter again', () => {
    openFilterPanel();
    cy.get('[data-testid="board-filter-toggle"]').click();
    cy.get('.filterPanel').should('not.exist');
  });

  it('shows active state on Filter when a chip is selected', () => {
    openFilterPanel();
    clickFilterChip('Completion', 'Active');
    cy.get('[data-testid="board-filter-toggle"]').should('have.class', 'btn-ghost-active');
  });

  it('Active hides a completed task', () => {
    cy.get('.task-card .task-check input[type="checkbox"]').first().check({ force: true });
    openFilterPanel();
    clickFilterChip('Completion', 'Active');
    cy.get('.column:not(.add-column)')
      .first()
      .find('.no-tasks')
      .should('contain', 'No tasks match the filter');
  });

  it('Done shows only completed tasks', () => {
    cy.get('.task-card .task-check input[type="checkbox"]').first().check({ force: true });
    openFilterPanel();
    clickFilterChip('Completion', 'Done');
    cy.get('.task-card').should('have.length', 1);
    cy.get('.task-card').first().should('have.class', 'task-completed');
  });

  it('Completion All shows tasks again after Active', () => {
    cy.get('.task-card .task-check input[type="checkbox"]').first().check({ force: true });
    openFilterPanel();
    clickFilterChip('Completion', 'Active');
    cy.get('.column:not(.add-column)').first().find('.no-tasks').should('exist');
    clickFilterChip('Completion', 'All');
    cy.get('.task-card').should('have.length', 1);
  });

  it('No date keeps a task with no due date', () => {
    openFilterPanel();
    clickFilterChip('Due date', 'No date');
    cy.get('.task-card').should('have.length', 1);
  });

  it('Due date Any clears due filter', () => {
    openFilterPanel();
    clickFilterChip('Due date', 'No date');
    clickFilterChip('Due date', 'Any');
    cy.get('.task-card').should('have.length', 1);
  });

  it('Clear filters shows tasks again after Active hid them', () => {
    cy.get('.task-card .task-check input[type="checkbox"]').first().check({ force: true });
    openFilterPanel();
    clickFilterChip('Completion', 'Active');
    cy.get('.column:not(.add-column)').first().find('.no-tasks').should('contain', 'No tasks match');
    cy.contains('button.filterClear', 'Clear filters').click();
    cy.get('.task-card').should('have.length', 1);
  });

  it('Priority High hides a task with no priority set', () => {
    openFilterPanel();
    clickFilterChip('Priority', 'High');
    cy.get('.column:not(.add-column)').first().find('.no-tasks').should('contain', 'No tasks match');
  });

  it('Priority Medium matches a task with Medium priority', () => {
    visitBoard({
      tasksByStageId: {
        100: [
          makeTask(
            6001,
            STAGE_ID,
            'Prio task',
            taskDescriptionWithMeta('Hello', { priority: 'Medium' })
          ),
        ],
        101: [],
      },
    });
    openFilterPanel();
    clickFilterChip('Priority', 'Medium');
    cy.get('.task-card').should('have.length', 1);
    cy.get('.task-card .task-title').should('contain', 'Prio task');
  });

  it('Priority Low hides a Medium-priority task', () => {
    visitBoard({
      tasksByStageId: {
        100: [
          makeTask(
            6002,
            STAGE_ID,
            'Prio task',
            taskDescriptionWithMeta('x', { priority: 'Medium' })
          ),
        ],
        101: [],
      },
    });
    openFilterPanel();
    clickFilterChip('Priority', 'Low');
    cy.get('.column:not(.add-column)').first().find('.no-tasks').should('contain', 'No tasks match');
  });

  it('Priority Critical chip can be selected', () => {
    openFilterPanel();
    clickFilterChip('Priority', 'Urgent');
    cy.contains('.filterGroupLabel', 'Priority')
      .parent()
      .find('.filterChip.chip-critical')
      .should('have.class', 'active');
  });
});

// ── Task search ────────────────────────────────────────────────────────────

describe('Board — task search', () => {
  beforeEach(() => {
    visitBoard({
      tasksByStageId: {
        [STAGE_ID]: [
          makeTask(8101, STAGE_ID, 'Write sprint demo script', 'Prepare narration notes'),
          makeTask(8102, STAGE_ID, 'Fix login bug', 'Investigate auth timeout'),
        ],
        [STAGE_2_ID]: [
          makeTask(8103, STAGE_2_ID, 'Polish board search', 'Search task cards by title'),
        ],
      },
    });
  });

  it('filters task cards by title and shows the result count', () => {
    cy.get('.searchInput').type('sprint');
    cy.get('.searchResultCount').should('contain', '1 result');
    cy.get('.task-card').should('have.length', 1);
    cy.contains('.task-card .task-title', 'Write sprint demo script').should('be.visible');
    cy.contains('.task-card .task-title', 'Fix login bug').should('not.exist');
  });

  it('filters task cards by description across columns', () => {
    cy.get('.searchInput').type('title');
    cy.get('.searchResultCount').should('contain', '1 result');
    cy.get('.task-card').should('have.length', 1);
    cy.contains('.column:not(.add-column)', 'Doing')
      .find('.task-card .task-title')
      .should('contain', 'Polish board search');
  });

  it('clear search restores all visible tasks', () => {
    cy.get('.searchInput').type('auth');
    cy.get('.task-card').should('have.length', 1);
    cy.get('.searchClear').click();
    cy.get('.searchResultCount').should('not.exist');
    cy.get('.task-card').should('have.length', 3);
  });

  it('shows an empty state when no task matches the search', () => {
    cy.get('.searchInput').type('no matching task');
    cy.get('.searchResultCount').should('contain', '0 results');
    cy.get('.column:not(.add-column)').first().find('.no-tasks').should('contain', 'No tasks match');
  });
});

// ── Filter: due date with fixed clock ───────────────────────────────────────

describe('Board — filter due date (clock)', () => {
  it('Overdue shows a past-due task', () => {
    cy.clock(new Date(2026, 5, 15, 12, 0, 0).getTime(), ['Date']);
    visitBoard({
      tasksByStageId: {
        100: [
          makeTask(
            7001,
            STAGE_ID,
            'Late',
            taskDescriptionWithMeta('x', { due: '2020-01-01' })
          ),
        ],
        101: [],
      },
    });
    openFilterPanel();
    clickFilterChip('Due date', 'Overdue');
    cy.get('.task-card').should('have.length', 1);
    cy.get('.task-card .task-title').should('contain', 'Late');
  });

  it('Today shows a task due today', () => {
    cy.clock(new Date(2026, 5, 15, 12, 0, 0).getTime(), ['Date']);
    // Slash date parses as local midnight in JS; ISO YYYY-MM-DD parses as UTC and breaks "today" vs local.
    visitBoard({
      tasksByStageId: {
        100: [
          makeTask(
            7002,
            STAGE_ID,
            'Due today',
            taskDescriptionWithMeta('x', { due: '6/15/2026' })
          ),
        ],
        101: [],
      },
    });
    openFilterPanel();
    clickFilterChip('Due date', 'Today');
    cy.get('.task-card').should('have.length', 1);
  });

  it('This week shows a task due within 7 days', () => {
    cy.clock(new Date(2026, 5, 15, 12, 0, 0).getTime(), ['Date']);
    visitBoard({
      tasksByStageId: {
        100: [
          makeTask(
            7003,
            STAGE_ID,
            'This week',
            taskDescriptionWithMeta('x', { due: '2026-06-18' })
          ),
        ],
        101: [],
      },
    });
    openFilterPanel();
    clickFilterChip('Due date', 'This week');
    cy.get('.task-card').should('have.length', 1);
  });
});

// ── Add task (create) ───────────────────────────────────────────────────────

describe('Board — add task', () => {
  it('adds a task with title only', () => {
    visitBoard();
    cy.get('.add-task-form')
      .first()
      .find('input[placeholder="+ Add a task..."]')
      .type('Quick task{enter}');
    cy.wait('@createTask').its('request.body.title').should('eq', 'Quick task');
    cy.contains('.task-card .task-title', 'Quick task').should('be.visible');
  });

  it('toggles Hide details after opening add details', () => {
    visitBoard();
    cy.get('.add-task-form')
      .first()
      .within(() => {
        cy.get('input[placeholder="+ Add a task..."]').type('T');
        cy.contains('button', '+ Add details').click();
        cy.get('textarea[placeholder="Description (optional)"]').should('be.visible');
        cy.contains('button', 'Hide details').click();
        cy.get('.task-detail-fields').should('not.exist');
      });
  });

  it('adds a task with description, priority, and notes in POST body', () => {
    visitBoard();
    cy.get('.add-task-form')
      .first()
      .within(() => {
        cy.get('input[placeholder="+ Add a task..."]').type('Full task');
        cy.contains('button', '+ Add details').click();
        cy.get('textarea[placeholder="Description (optional)"]').type('Desc line');
        cy.get('.task-meta-grid select').select('High');
        cy.get('textarea[placeholder="Notes (optional)"]').type('Note A');
        cy.contains('button', 'Add').click();
      });
    cy.wait('@createTask').then((interception) => {
      const d = interception.request.body.description as string;
      expect(d).to.include('Desc line');
      expect(d).to.include('priority:High');
      expect(d).to.include('notes:Note A');
    });
  });
});

// ── Add list (stage) ────────────────────────────────────────────────────────

describe('Board — add list (stage)', () => {
  it('creates a new column via Add another list', () => {
    visitBoard();
    cy.get('.add-column .add-stage-input').type('Backlog');
    cy.contains('button', '+ Add another list').click();
    cy.wait('@createStage');
    cy.get('.column:not(.add-column)').should('have.length', 3);
    cy.get('.column:not(.add-column)').contains('h3', 'Backlog').should('be.visible');
  });
});

// ── Task detail modal (edit + close) ───────────────────────────────────────

describe('Board — task detail modal', () => {
  beforeEach(() => visitBoard());

  it('closes with the X button', () => {
    cy.get('.task-card .task-content').first().click();
    cy.get('.task-modal .btn-close-modal').click();
    cy.get('.task-modal').should('not.exist');
  });

  it('closes with Cancel', () => {
    cy.get('.task-card .task-content').first().click();
    cy.contains('.task-modal button', 'Cancel').click();
    cy.get('.task-modal').should('not.exist');
  });

  it('saves edited title and priority', () => {
    cy.get('.task-card .task-content').first().click();
    cy.get('.task-modal input[type="text"]').first().clear().type('Renamed');
    cy.get('.task-modal select').first().select('Low');
    cy.contains('.task-modal button', 'Save').click();
    cy.wait('@updateTask').then((i) => {
      expect(i.request.body.title).to.eq('Renamed');
      expect(i.request.body.description).to.include('priority:Low');
    });
  });

  it('saves notes in description payload', () => {
    cy.get('.task-card .task-content').first().click();
    cy.get('.task-detail-main textarea').last().type('My notes');
    cy.contains('.task-modal button', 'Save').click();
    cy.wait('@updateTask').its('request.body.description').should('include', 'notes:My notes');
  });
});

// ── Delete task ──────────────────────────────────────────────────────────────

describe('Board — delete task', () => {
  it('opens confirm and cancels', () => {
    visitBoard();
    cy.get('.task-card .btn-delete-task').first().click();
    cy.contains('h3', 'Delete task?').should('be.visible');
    cy.contains('.delete-confirm-modal button', 'Cancel').click();
    cy.get('.delete-confirm-modal').should('not.exist');
    cy.get('.task-card').should('have.length', 1);
  });

  it('deletes after confirm', () => {
    visitBoard();
    cy.get('.task-card .btn-delete-task').first().click();
    cy.contains('.delete-confirm-modal button', 'Delete').click();
    cy.wait('@deleteTask');
    cy.get('.task-card').should('have.length', 0);
    cy.get('.column:not(.add-column)').first().find('.no-tasks').should('contain', 'No tasks yet');
  });
});

// ── Delete stage ───────────────────────────────────────────────────────────

describe('Board — delete stage', () => {
  it('removes a column after modal confirm', () => {
    visitBoard();
    cy.get('.column:not(.add-column)').eq(1).find('.btn-delete').click();
    cy.contains('h3', 'Delete this list?').should('be.visible');
    cy.contains('.delete-confirm-modal button', 'Delete list').click();
    cy.wait('@deleteStage');
    cy.get('.column:not(.add-column)').should('have.length', 1);
    cy.get('.column:not(.add-column)').first().find('h3').should('contain', 'To Do');
  });
});

// ── Top bar: back, theme, share, switcher, profile ─────────────────────────

describe('Board — top bar', () => {
  it('navigates to My boards when clicking Back', () => {
    visitBoard();
    cy.get('button.btn-back').contains('Back').click();
    cy.url().should('include', '/boards');
  });

  it('toggles theme (data-theme on html)', () => {
    visitBoard();
    cy.get('html')
      .invoke('attr', 'data-theme')
      .then((before) => {
        cy.get('button.themeToggle').click();
        const next = before === 'dark' ? 'light' : 'dark';
        cy.get('html').should('have.attr', 'data-theme', next);
      });
  });

  it('opens Share modal and closes with Done', () => {
    visitBoard();
    cy.get('button.btn-primary').contains('Share').click();
    cy.contains('h3', 'Share board').should('be.visible');
    cy.get('.shareModal .shareLinkInput').should('be.visible');
    cy.contains('.shareModal button', 'Done').click();
    cy.get('.shareModal').should('not.exist');
  });

  it('switches board from the board switcher', () => {
    visitBoard({
      projectsList: [
        { id: 1, name: 'Alpha', description: '', created_at: isoNow(), updated_at: isoNow() },
        { id: 2, name: 'Beta', description: '', created_at: isoNow(), updated_at: isoNow() },
      ],
    });
    cy.get('[data-testid="board-switcher-toggle"]').click();
    cy.contains('.boardSwitcherItem', 'Beta').click();
    cy.url().should('include', '/board/2');
    cy.get('.board-title').should('contain', 'E2E Board 2');
  });

  it('opens Account settings from the profile menu', () => {
    visitBoard();
    cy.get('button.profileIcon').click();
    cy.contains('button', 'Account settings').click();
    cy.contains('h3', 'Account settings').should('be.visible');
    cy.get('.accountModal .btn-close-modal').click();
    cy.get('.accountModal').should('not.exist');
  });
});

// ── Empty board ──────────────────────────────────────────────────────────────

describe('Board — checklist subtasks', () => {
  beforeEach(() =>
    visitBoard({
      subtasksByTaskId: {
        [TASK_ID]: [
          makeSubtask(TASK_ID, 7101, 'Draft outline'),
          makeSubtask(TASK_ID, 7102, 'Review copy', true, 1),
        ],
      },
    })
  );

  it('shows checklist items in order and renders progress on the task card', () => {
    cy.get('.task-card').first().within(() => {
      cy.get('.task-subtask-progress-text').should('contain', '1/2 done');
    });

    cy.get('.task-card .task-content').first().click();
    cy.get('.subtaskItemTitle').eq(0).should('contain', 'Draft outline');
    cy.get('.subtaskItemTitle').eq(1).should('contain', 'Review copy');
  });

  it('adds a checklist item from the task detail modal', () => {
    cy.get('.task-card .task-content').first().click();
    cy.get('.subtaskComposer input').type('Ship release notes{enter}');
    cy.wait('@createSubtask').its('request.body.title').should('eq', 'Ship release notes');
    cy.get('.subtaskItemTitle').should('contain', 'Ship release notes');
    cy.get('.task-subtask-progress-text').first().should('contain', '1/3 done');
  });

  it('toggles a checklist item without refreshing and updates progress immediately', () => {
    cy.get('.task-card .task-content').first().click();
    cy.contains('.subtaskItem', 'Draft outline').find('input[type="checkbox"]').check({ force: true });
    cy.wait('@updateSubtask').its('request.body.is_completed').should('eq', true);
    cy.contains('.subtaskItem', 'Draft outline').find('.subtaskItemTitle').should('have.class', 'subtaskItemTitleDone');
    cy.get('.task-subtask-progress-text').first().should('contain', '2/2 done');
  });

  it('deletes a checklist item and shrinks the progress summary', () => {
    cy.get('.task-card .task-content').first().click();
    cy.contains('.subtaskItem', 'Review copy').contains('button', 'Delete').click();
    cy.contains('.delete-confirm-modal h3', 'Delete checklist item?').should('be.visible');
    cy.contains('.delete-confirm-modal button', 'Delete').click();
    cy.wait('@deleteSubtask');
    cy.contains('.subtaskItem', 'Review copy').should('not.exist');
    cy.get('.task-subtask-progress-text').first().should('contain', '0/1 done');
  });

  it('keeps a checklist item when delete is cancelled', () => {
    cy.get('.task-card .task-content').first().click();
    cy.contains('.subtaskItem', 'Review copy').contains('button', 'Delete').click();
    cy.contains('.delete-confirm-modal button', 'Cancel').click();
    cy.get('.delete-confirm-modal').should('not.exist');
    cy.contains('.subtaskItem', 'Review copy').should('exist');
  });

  it('keeps checklist changes visible after switching to planner', () => {
    cy.get('.task-card .task-content').first().click();
    cy.get('.subtaskComposer input').type('Planner sync item');
    cy.contains('.subtaskComposer button', 'Add').click();
    cy.wait('@createSubtask');
    cy.get('.task-modal .btn-close-modal').click();
    cy.get('nav.viewTabs a.viewTab').contains('Planner').click({ force: true });
    cy.get('[data-testid="planner-nodue-toggle"]').click();
    cy.contains('.planner-task-title', 'Test task').click();
    cy.get('.plannerSubtaskItemTitle').should('contain', 'Planner sync item');
  });
});

describe('Board — task comments', () => {
  beforeEach(() =>
    visitBoard({
      commentsByTaskId: {
        [TASK_ID]: [
          makeComment(TASK_ID, 9101, 'Initial note'),
          makeComment(TASK_ID, 9102, 'Follow-up comment'),
        ],
      },
    })
  );

  it('posts a comment from task details', () => {
    cy.get('.task-card .task-content').first().click();
    cy.get('.commentsComposer textarea').type('Fresh update from Cypress');
    cy.contains('.commentsComposer button', 'Post').click();
    cy.wait('@createComment').its('request.body.content').should('eq', 'Fresh update from Cypress');
    cy.contains('.commentCard', 'Fresh update from Cypress').should('be.visible');
    cy.get('.task-comment-meta').first().should('contain', '3');
  });

  it('edits an existing comment', () => {
    cy.get('.task-card .task-content').first().click();
    cy.contains('.commentCard', 'Initial note').contains('button', 'Edit').click();
    cy.get('.commentEditForm textarea').clear().type('Updated note copy');
    cy.contains('.commentEditActions button', 'Save').click();
    cy.wait('@updateComment').its('request.body.content').should('eq', 'Updated note copy');
    cy.contains('.commentCard', 'Updated note copy').should('be.visible');
  });

  it('cancels comment deletion from the confirm modal', () => {
    cy.get('.task-card .task-content').first().click();
    cy.contains('.commentCard', 'Follow-up comment').contains('button', 'Delete').click();
    cy.contains('.delete-confirm-modal h3', 'Delete comment?').should('be.visible');
    cy.contains('.delete-confirm-modal button', 'Cancel').click();
    cy.get('.delete-confirm-modal').should('not.exist');
    cy.contains('.commentCard', 'Follow-up comment').should('exist');
  });

  it('deletes a comment after confirming', () => {
    cy.get('.task-card .task-content').first().click();
    cy.contains('.commentCard', 'Follow-up comment').contains('button', 'Delete').click();
    cy.contains('.delete-confirm-modal button', 'Delete').click();
    cy.wait('@deleteComment');
    cy.contains('.commentCard', 'Follow-up comment').should('not.exist');
    cy.get('.task-comment-meta').first().should('contain', '1');
  });
});

describe('Board — project activity', () => {
  const ownerMember = makeProjectMember(1, 'owner-1', 'Adithya', 'e2e@test.com', 'owner');
  const teammateMember = makeProjectMember(1, 'member-2', 'Casey Doe', 'casey@test.com', 'member');

  it('shows the Activity tab only for the owner', () => {
    visitBoard({
      owners: { '1': 'owner@test.com' },
      sessionUser: { name: 'Jordan', email: 'member@test.com', id: 'member-3' },
      cachedProjectMembers: {
        1: [
          makeProjectMember(1, 'owner-1', 'Owner User', 'owner@test.com', 'owner'),
          makeProjectMember(1, 'member-3', 'Jordan', 'member@test.com', 'member'),
        ],
      },
      projectMembers: [
        makeProjectMember(1, 'owner-1', 'Owner User', 'owner@test.com', 'owner'),
        makeProjectMember(1, 'member-3', 'Jordan', 'member@test.com', 'member'),
      ],
    });

    cy.contains('button.btn-ghost', 'Settings').click();
    cy.contains('h3', 'Project settings').should('be.visible');
    cy.contains('.settingsTab', 'Activity').should('not.exist');
  });

  it('renders newest activity first, filters by member/date, and paginates with load more', () => {
    visitBoard({
      projectMembers: [ownerMember, teammateMember],
      activityByProjectId: {
        1: [
          makeActivity(1, 1, "Adithya moved 'Design login' to Done", { userId: 'owner-1', userName: 'Adithya', createdAt: '2026-04-12T14:00:00.000Z' }),
          makeActivity(1, 2, "Casey added label 'Urgent' to 'Fix bugs'", { userId: 'member-2', userName: 'Casey Doe', action: 'label_assigned', entityType: 'label', createdAt: '2026-04-11T11:00:00.000Z' }),
          makeActivity(1, 3, "Adithya commented on 'Launch prep'", { userId: 'owner-1', userName: 'Adithya', action: 'comment_added', entityType: 'comment', createdAt: '2026-04-10T09:00:00.000Z' }),
          makeActivity(1, 4, "Casey created 'Write release notes'", { userId: 'member-2', userName: 'Casey Doe', action: 'task_created', createdAt: '2026-04-09T09:00:00.000Z' }),
          makeActivity(1, 5, "Adithya updated 'Board polish'", { userId: 'owner-1', userName: 'Adithya', action: 'task_updated', createdAt: '2026-04-08T09:00:00.000Z' }),
          makeActivity(1, 6, "Casey moved 'Homepage QA' to Doing", { userId: 'member-2', userName: 'Casey Doe', createdAt: '2026-04-07T09:00:00.000Z' }),
          makeActivity(1, 7, "Adithya removed label 'Blocked'", { userId: 'owner-1', userName: 'Adithya', action: 'label_removed', entityType: 'label', createdAt: '2026-04-06T09:00:00.000Z' }),
          makeActivity(1, 8, "Casey joined the project", { userId: 'member-2', userName: 'Casey Doe', action: 'member_joined', entityType: 'member', createdAt: '2026-04-05T09:00:00.000Z' }),
          makeActivity(1, 9, "Adithya added Casey to the board", { userId: 'owner-1', userName: 'Adithya', action: 'member_added', entityType: 'member', createdAt: '2026-04-04T09:00:00.000Z' }),
          makeActivity(1, 10, "Casey deleted 'Old draft'", { userId: 'member-2', userName: 'Casey Doe', action: 'task_deleted', createdAt: '2026-04-03T09:00:00.000Z' }),
          makeActivity(1, 11, "Adithya moved 'Sprint wrap-up' to Done", { userId: 'owner-1', userName: 'Adithya', createdAt: '2026-04-02T09:00:00.000Z' }),
          makeActivity(1, 12, "Casey added a comment to 'Roadmap'", { userId: 'member-2', userName: 'Casey Doe', action: 'comment_added', entityType: 'comment', createdAt: '2026-04-01T09:00:00.000Z' }),
          makeActivity(1, 13, "Adithya updated 'Retro notes'", { userId: 'owner-1', userName: 'Adithya', action: 'task_updated', createdAt: '2026-03-31T09:00:00.000Z' }),
        ],
      },
    });

    cy.contains('button.btn-ghost', 'Settings').click();
    cy.contains('.settingsTab', 'Activity').click();
    cy.get('.activityList .activityItem').should('have.length', 12);
    cy.get('.activityList .activityItem').first().should('contain', "Adithya moved 'Design login' to Done");
    cy.get('.activityList .activityItem').last().should('contain', "Casey added a comment to 'Roadmap'");

    cy.get('#activity-member-filter').select('member-2');
    cy.get('.activityList .activityItem').should('have.length', 6);
    cy.get('.activityList').should('contain', 'Casey added label');
    cy.get('.activityList').should('not.contain', "Adithya moved 'Design login' to Done");

    cy.get('#activity-date-from')
      .invoke('val', '2026-04-05')
      .trigger('input', { force: true })
      .trigger('change', { force: true });
    cy.get('.activityList .activityItem').should('have.length', 4);
    cy.get('.activityList').should('not.contain', "Casey deleted 'Old draft'");

    cy.contains('.activityToolbar button', 'Reset').click();
    cy.get('.activityList .activityItem').should('have.length', 12);
    cy.contains('.activityMore button', 'Load more').click();
    cy.get('.activityList .activityItem').should('have.length', 13);
    cy.get('.activityList').should('contain', "Adithya updated 'Retro notes'");
  });

  it('shows an empty state when no activity has been logged', () => {
    visitBoard({
      projectMembers: [ownerMember, teammateMember],
      activityByProjectId: { 1: [] },
    });

    cy.contains('button.btn-ghost', 'Settings').click();
    cy.contains('.settingsTab', 'Activity').click();
    cy.contains('.activityEmptyState strong', 'No activity logged yet.').should('be.visible');
  });
});

describe('Board — empty list', () => {
  it('shows no tasks when a stage has no tasks', () => {
    visitBoard({
      tasksByStageId: { 100: [], 101: [] },
      skipTaskCardAssert: true,
    });
    cy.get('.column:not(.add-column)').first().find('.no-tasks').should('contain', 'No tasks yet');
  });
});

describe('Board — column sorting', () => {
  it('sorts a column alphabetically when A-Z is selected', () => {
    visitBoard({
      tasksByStageId: {
        [STAGE_ID]: [
          makeTask(9001, STAGE_ID, 'Zulu task', ''),
          makeTask(9002, STAGE_ID, 'Alpha task', ''),
        ],
        [STAGE_2_ID]: [],
      },
    });

    cy.get('[data-testid="column-sort-100"]').select('alphabetical');
    cy.get('.column:not(.add-column)').first().find('.task-title').eq(0).should('contain', 'Alpha task');
    cy.get('.column:not(.add-column)').first().find('.task-title').eq(1).should('contain', 'Zulu task');
  });
});

describe('Board — table sorting', () => {
  it('toggles title sort direction when clicking the same header', () => {
    visitBoard({
      tasksByStageId: {
        [STAGE_ID]: [
          makeTask(9101, STAGE_ID, 'Bravo', ''),
          makeTask(9102, STAGE_ID, 'Alpha', ''),
        ],
        [STAGE_2_ID]: [],
      },
    });

    cy.contains('button.viewTab', 'Table').click();
    cy.get('.taskTable tbody .taskRow').eq(0).should('contain', 'Alpha');
    cy.get('[data-testid="table-sort-title"]').click();
    cy.get('.taskTable tbody .taskRow').eq(0).should('contain', 'Bravo');
  });

  it('sorts by due date when Due Date header is clicked', () => {
    visitBoard({
      tasksByStageId: {
        [STAGE_ID]: [
          makeTask(9201, STAGE_ID, 'Later due', taskDescriptionWithMeta('', { due: '2030-02-01' })),
          makeTask(9202, STAGE_ID, 'Sooner due', taskDescriptionWithMeta('', { due: '2030-01-01' })),
        ],
        [STAGE_2_ID]: [],
      },
    });

    cy.contains('button.viewTab', 'Table').click();
    cy.get('[data-testid="table-sort-due"]').click();
    cy.get('.taskTable tbody .taskRow').eq(0).should('contain', 'Later due');
    cy.get('[data-testid="table-sort-due"]').click();
    cy.get('.taskTable tbody .taskRow').eq(0).should('contain', 'Sooner due');
  });

  it('sorts by recently updated when Updated header is clicked', () => {
    visitBoard({
      tasksByStageId: {
        [STAGE_ID]: [
          {
            ...makeTask(9301, STAGE_ID, 'Old update', ''),
            updated_at: '2022-01-01T00:00:00.000Z',
          },
          {
            ...makeTask(9302, STAGE_ID, 'New update', ''),
            updated_at: '2024-01-01T00:00:00.000Z',
          },
        ],
        [STAGE_2_ID]: [],
      },
    });

    cy.contains('button.viewTab', 'Table').click();
    cy.get('[data-testid="table-sort-updated"]').click();
    cy.get('.taskTable tbody .taskRow').eq(0).should('contain', 'Old update');
    cy.get('[data-testid="table-sort-updated"]').click();
    cy.get('.taskTable tbody .taskRow').eq(0).should('contain', 'New update');
  });
});
