/// <reference types="cypress" />

/** Card meta block (matches board `buildCardDescription` / `parseCardMeta`). */
export function taskDescriptionWithMeta(
  userDesc: string,
  opts: { due?: string; priority?: string; notes?: string }
): string {
  const parts: string[] = [];
  if (opts.due?.trim()) parts.push(`due:${opts.due.trim()}`);
  if (opts.priority?.trim()) parts.push(`priority:${opts.priority.trim()}`);
  if (opts.notes?.trim()) parts.push(`notes:${opts.notes.trim()}`);
  if (parts.length === 0) return userDesc.trim();
  return `${userDesc.trim()}\n---\n${parts.join('\n')}`;
}

export const PROJECT_ID = 1;
export const STAGE_ID = 100;
export const STAGE_2_ID = 101;
export const TASK_ID = 5001;

export const isoNow = () => new Date().toISOString();

export function seedBoardAuth(
  win: Cypress.AUTWindow,
  owners: Record<string, string> = { '1': 'e2e@test.com', '2': 'e2e@test.com' }
) {
  win.localStorage.setItem('taskify.auth.token', 'e2e-token');
  win.localStorage.setItem(
    'taskify.auth.session',
    JSON.stringify({ name: 'E2E User', email: 'e2e@test.com' })
  );
  win.localStorage.setItem('taskify.board.owners', JSON.stringify(owners));
}

export type BoardStubOptions = {
  projectId?: number;
  /** Tasks per stage id (default: one empty task on stage 100, none on 101). */
  tasksByStageId?: Record<number, unknown[]>;
  /** Subtasks per task id. */
  subtasksByTaskId?: Record<number, unknown[]>;
  /** Override stage list for project `projectId` (default: To Do + Doing). */
  stages?: unknown[];
  /** GET /api/projects — for board switcher (default: []). */
  projectsList?: unknown[];
};

export type VisitBoardOptions = BoardStubOptions & {
  /** Set true when the stub has no tasks (skip waiting for `.task-card`). */
  skipTaskCardAssert?: boolean;
};

const defaultProject = (id: number) => ({
  id,
  name: `E2E Board ${id}`,
  description: '',
  created_at: isoNow(),
  updated_at: isoNow(),
});

const defaultStage = (id: number, name: string, position: number, projectId: number) => ({
  id,
  project_id: projectId,
  name,
  position,
  created_at: isoNow(),
  updated_at: isoNow(),
});

const defaultTask = (id: number, stageId: number, title: string, description: string) => ({
  id,
  stage_id: stageId,
  title,
  description,
  position: 0,
  subtask_count: 0,
  completed_count: 0,
  completed: false,
  created_at: isoNow(),
  updated_at: isoNow(),
});

/** Build a task row for `tasksByStageId` stubs. */
export function makeTask(
  id: number,
  stageId: number,
  title: string,
  description: string,
  position = 0
) {
  return {
    id,
    stage_id: stageId,
    title,
    description,
    position,
    subtask_count: 0,
    completed_count: 0,
    completed: false,
    created_at: isoNow(),
    updated_at: isoNow(),
  };
}

export function makeSubtask(taskId: number, id: number, title: string, isCompleted = false, position = 0) {
  return {
    id,
    task_id: taskId,
    title,
    is_completed: isCompleted,
    position,
    created_at: isoNow(),
    updated_at: isoNow(),
  };
}

/**
 * Registers HTTP stubs for board E2E. Call once per visit, before `cy.visit`.
 */
export function registerBoardApiStubs(opts: BoardStubOptions = {}) {
  const projectId = opts.projectId ?? PROJECT_ID;
  const base = `/api/projects/${projectId}`;

  const stages =
    opts.stages ??
    [
      defaultStage(STAGE_ID, 'To Do', 0, projectId),
      defaultStage(STAGE_2_ID, 'Doing', 1, projectId),
    ];

  const tasksByStageId: Record<number, unknown[]> = {
    [STAGE_ID]: [defaultTask(TASK_ID, STAGE_ID, 'Test task', '')],
    [STAGE_2_ID]: [],
  };
  if (opts.tasksByStageId) {
    Object.keys(opts.tasksByStageId).forEach((k) => {
      const id = Number(k);
      tasksByStageId[id] = opts.tasksByStageId![id] as unknown[];
    });
  }

  const subtasksByTaskId: Record<number, any[]> = {};
  Object.values(tasksByStageId).forEach((tasks) => {
    (tasks as any[]).forEach((task) => {
      subtasksByTaskId[Number(task.id)] = [];
    });
  });
  if (opts.subtasksByTaskId) {
    Object.keys(opts.subtasksByTaskId).forEach((k) => {
      const id = Number(k);
      subtasksByTaskId[id] = [...((opts.subtasksByTaskId?.[id] as any[]) || [])];
    });
  }

  const projectsList = opts.projectsList ?? [];

  let nextCreateTaskId = 9000;
  let nextCreateStageId = 8000;
  let nextCreateSubtaskId = 12000;

  const clone = <T>(value: T): T => JSON.parse(JSON.stringify(value));

  const findTask = (taskId: number): any | undefined => {
    for (const tasks of Object.values(tasksByStageId)) {
      const task = (tasks as any[]).find((item) => Number(item.id) === taskId);
      if (task) return task;
    }
    return undefined;
  };

  const syncTaskCounts = (taskId: number) => {
    const task = findTask(taskId);
    if (!task) return;
    const subtasks = subtasksByTaskId[taskId] || [];
    task.subtask_count = subtasks.length;
    task.completed_count = subtasks.filter((subtask) => !!subtask.is_completed).length;
  };

  Object.keys(subtasksByTaskId).forEach((taskId) => syncTaskCounts(Number(taskId)));

  cy.intercept('GET', '**/api/**', (req) => {
    const path = new URL(req.url).pathname.replace(/\/$/, '');

    const subtasksMatch = path.match(/^\/api\/tasks\/(\d+)\/subtasks$/);
    if (subtasksMatch) {
      const taskId = Number(subtasksMatch[1]);
      req.reply(clone((subtasksByTaskId[taskId] || []).sort((a, b) => a.position - b.position)));
      return;
    }

    const tasksMatch = path.match(/^\/api\/projects\/(\d+)\/stages\/(\d+)\/tasks$/);
    if (tasksMatch) {
      const sid = Number(tasksMatch[2]);
      req.reply(clone(tasksByStageId[sid] ?? []));
      return;
    }

    const stagesMatch = path.match(/^\/api\/projects\/(\d+)\/stages$/);
    if (stagesMatch) {
      const pid = Number(stagesMatch[1]);
      if (pid === projectId) {
        req.reply(stages);
      } else {
        req.reply([
          defaultStage(300, 'Other list', 0, pid),
        ]);
      }
      return;
    }

    const projectMatch = path.match(/^\/api\/projects\/(\d+)$/);
    if (projectMatch) {
      const pid = Number(projectMatch[1]);
      req.reply(defaultProject(pid));
      return;
    }

    if (path === '/api/projects') {
      req.reply(projectsList);
      return;
    }

    req.continue();
  });

  cy.intercept('PUT', '**/api/tasks/*', (req) => {
    const url = req.url;
    const m = url.match(/\/api\/tasks\/(\d+)/);
    const id = m ? Number(m[1]) : TASK_ID;
    const existingTask = findTask(id);
    if (existingTask) {
      existingTask.title = req.body.title;
      existingTask.description = req.body.description;
      existingTask.position = req.body.position ?? existingTask.position ?? 0;
      existingTask.updated_at = isoNow();
    }
    req.reply({
      statusCode: 200,
      body: {
        id,
        stage_id: existingTask?.stage_id ?? STAGE_ID,
        title: req.body.title,
        description: req.body.description,
        position: req.body.position ?? existingTask?.position ?? 0,
        subtask_count: existingTask?.subtask_count ?? 0,
        completed_count: existingTask?.completed_count ?? 0,
        completed: false,
        created_at: existingTask?.created_at ?? isoNow(),
        updated_at: isoNow(),
      },
    });
  }).as('updateTask');

  cy.intercept('POST', '**/api/projects/*/stages/*/tasks', (req) => {
    const m = req.url.match(/\/projects\/(\d+)\/stages\/(\d+)\/tasks/);
    const sid = m ? Number(m[2]) : STAGE_ID;
    nextCreateTaskId += 1;
    const createdTask = {
      id: nextCreateTaskId,
      stage_id: sid,
      title: req.body.title,
      description: req.body.description ?? '',
      position: req.body.position ?? 0,
      subtask_count: 0,
      completed_count: 0,
      completed: false,
      created_at: isoNow(),
      updated_at: isoNow(),
    };
    if (!tasksByStageId[sid]) tasksByStageId[sid] = [];
    (tasksByStageId[sid] as any[]).push(createdTask);
    subtasksByTaskId[nextCreateTaskId] = [];
    req.reply({
      statusCode: 201,
      body: clone(createdTask),
    });
  }).as('createTask');

  cy.intercept('POST', '**/api/tasks/*/subtasks', (req) => {
    const m = req.url.match(/\/api\/tasks\/(\d+)\/subtasks/);
    const taskId = m ? Number(m[1]) : TASK_ID;
    const existing = subtasksByTaskId[taskId] || [];
    nextCreateSubtaskId += 1;
    const createdSubtask = {
      id: nextCreateSubtaskId,
      task_id: taskId,
      title: req.body.title,
      is_completed: false,
      position: req.body.position ?? existing.length,
      created_at: isoNow(),
      updated_at: isoNow(),
    };
    subtasksByTaskId[taskId] = [...existing, createdSubtask]
      .sort((a, b) => a.position - b.position)
      .map((subtask, index) => ({ ...subtask, position: index }));
    syncTaskCounts(taskId);
    req.reply({ statusCode: 201, body: clone(createdSubtask) });
  }).as('createSubtask');

  cy.intercept('PATCH', '**/api/subtasks/*', (req) => {
    const m = req.url.match(/\/api\/subtasks\/(\d+)/);
    const subtaskId = m ? Number(m[1]) : 0;
    let updatedSubtask: any = null;

    Object.keys(subtasksByTaskId).forEach((taskIdKey) => {
      const taskId = Number(taskIdKey);
      subtasksByTaskId[taskId] = (subtasksByTaskId[taskId] || []).map((subtask) => {
        if (Number(subtask.id) !== subtaskId) return subtask;
        updatedSubtask = {
          ...subtask,
          ...req.body,
          updated_at: isoNow(),
        };
        return updatedSubtask;
      });
      if (updatedSubtask) {
        subtasksByTaskId[taskId] = subtasksByTaskId[taskId]
          .sort((a, b) => a.position - b.position)
          .map((subtask, index) => ({ ...subtask, position: index }));
        syncTaskCounts(taskId);
      }
    });

    req.reply({ statusCode: 200, body: clone(updatedSubtask) });
  }).as('updateSubtask');

  cy.intercept('DELETE', '**/api/subtasks/*', (req) => {
    const m = req.url.match(/\/api\/subtasks\/(\d+)/);
    const subtaskId = m ? Number(m[1]) : 0;

    Object.keys(subtasksByTaskId).forEach((taskIdKey) => {
      const taskId = Number(taskIdKey);
      subtasksByTaskId[taskId] = (subtasksByTaskId[taskId] || [])
        .filter((subtask) => Number(subtask.id) !== subtaskId)
        .map((subtask, index) => ({ ...subtask, position: index }));
      syncTaskCounts(taskId);
    });

    req.reply({ statusCode: 204, body: null });
  }).as('deleteSubtask');

  cy.intercept('POST', '**/api/projects/*/stages', (req) => {
    const m = req.url.match(/\/projects\/(\d+)\/stages$/);
    const pid = m ? Number(m[1]) : projectId;
    const body = req.body as { name: string; position: number };
    nextCreateStageId += 1;
    req.reply({
      statusCode: 201,
      body: {
        id: nextCreateStageId,
        project_id: pid,
        name: body.name,
        position: body.position ?? 0,
        created_at: isoNow(),
        updated_at: isoNow(),
      },
    });
  }).as('createStage');

  cy.intercept('DELETE', '**/api/stages/*', { statusCode: 204, body: null }).as('deleteStage');

  cy.intercept('DELETE', '**/api/tasks/*', { statusCode: 204, body: null }).as('deleteTask');
}

export function visitBoard(opts: VisitBoardOptions = {}) {
  const { skipTaskCardAssert, ...stubOpts } = opts;
  registerBoardApiStubs(stubOpts);
  const pid = stubOpts.projectId ?? PROJECT_ID;
  cy.visit(`/board/${pid}`, {
    onBeforeLoad(win) {
      seedBoardAuth(win);
    },
  });
  cy.get('.board-content', { timeout: 15000 }).should('be.visible');
  if (!skipTaskCardAssert) {
    cy.get('.task-card', { timeout: 10000 }).should('have.length.at.least', 1);
  }
}

/**
 * Opens the planner (calendar) view for a board. Reuses the same API stubs as {@link visitBoard}.
 */
export function visitPlanner(opts: VisitBoardOptions = {}) {
  const stubOpts: BoardStubOptions = {
    projectId: opts.projectId,
    tasksByStageId: opts.tasksByStageId,
    subtasksByTaskId: opts.subtasksByTaskId,
    stages: opts.stages,
    projectsList: opts.projectsList,
  };
  registerBoardApiStubs(stubOpts);
  const pid = stubOpts.projectId ?? PROJECT_ID;
  cy.visit(`/board/${pid}/planner`, {
    onBeforeLoad(win) {
      seedBoardAuth(win);
    },
  });
  cy.get('.planner-main', { timeout: 15000 }).should('be.visible');
  cy.get('.planner-grid', { timeout: 10000 }).should('exist');
}
