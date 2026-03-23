/// <reference types="cypress" />

import {
  STAGE_ID,
  makeTask,
  taskDescriptionWithMeta,
  visitBoard,
  isoNow,
} from '../support/board-stubs';

function openFilterPanel() {
  cy.get('button.btn-ghost').contains('Filter').click();
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
    cy.get('button.btn-ghost').contains('Filter').click();
    cy.get('.filterPanel').should('not.exist');
  });

  it('shows active state on Filter when a chip is selected', () => {
    openFilterPanel();
    clickFilterChip('Completion', 'Active');
    cy.get('button.btn-ghost').contains('Filter').should('have.class', 'btn-ghost-active');
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
    clickFilterChip('Priority', 'Critical');
    cy.contains('.filterGroupLabel', 'Priority')
      .parent()
      .find('.filterChip.chip-critical')
      .should('have.class', 'active');
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
    cy.get('.task-modal textarea').last().type('My notes');
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
  it('removes a column after browser confirm', () => {
    visitBoard();
    cy.window().then((win) => {
      win.confirm = () => true;
    });
    cy.get('.column:not(.add-column)').eq(1).find('.btn-delete').click();
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
    cy.get('.boardSwitcherBtn').click();
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

describe('Board — empty list', () => {
  it('shows no tasks when a stage has no tasks', () => {
    visitBoard({
      tasksByStageId: { 100: [], 101: [] },
      skipTaskCardAssert: true,
    });
    cy.get('.column:not(.add-column)').first().find('.no-tasks').should('contain', 'No tasks yet');
  });
});
