export interface Task {
  id: number;
  stage_id: number;
  title: string;
  description: string;
  position: number;
  /** Client-only; persisted via TaskCompletionStorageService (localStorage). */
  completed?: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateTaskRequest {
  title: string;
  description: string;
  position: number;
}

export interface UpdateTaskRequest {
  title: string;
  description: string;
  position: number;
}

export interface MoveTaskRequest {
  newStageId: number;
  newPos: number;
}