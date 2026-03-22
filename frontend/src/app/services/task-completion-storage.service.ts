import { Injectable } from '@angular/core';
import { Task } from '../models/task.model';

/**
 * Persists task "done" state in the browser only (no backend).
 * Keyed by project + task id so completion survives refresh.
 */
@Injectable({ providedIn: 'root' })
export class TaskCompletionStorageService {
  private readonly key = 'taskify.taskCompletion.v1';

  private read(): Record<string, Record<string, boolean>> {
    try {
      const raw = localStorage.getItem(this.key);
      if (!raw) return {};
      const parsed = JSON.parse(raw) as Record<string, Record<string, boolean>>;
      return parsed && typeof parsed === 'object' ? parsed : {};
    } catch {
      return {};
    }
  }

  private write(data: Record<string, Record<string, boolean>>): void {
    try {
      localStorage.setItem(this.key, JSON.stringify(data));
    } catch {
      // quota / private mode — UI still works for the session
    }
  }

  getCompleted(projectId: number, taskId: number): boolean {
    const pid = String(projectId);
    const tid = String(taskId);
    return !!this.read()[pid]?.[tid];
  }

  setCompleted(projectId: number, taskId: number, completed: boolean): void {
    const map = this.read();
    const pid = String(projectId);
    const tid = String(taskId);
    if (!map[pid]) map[pid] = {};
    if (completed) {
      map[pid][tid] = true;
    } else {
      delete map[pid][tid];
      if (Object.keys(map[pid]).length === 0) {
        delete map[pid];
      }
    }
    this.write(map);
  }

  /** Attach `completed` from storage to tasks returned by the API. */
  mergeTasks(projectId: number, tasks: Task[]): Task[] {
    return tasks.map((t) => ({
      ...t,
      completed: this.getCompleted(projectId, t.id),
    }));
  }
}
