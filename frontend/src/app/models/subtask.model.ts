export interface Subtask {
  id: number;
  task_id: number;
  title: string;
  is_completed: boolean;
  position: number;
  created_at: string;
  updated_at: string;
}

export interface CreateSubtaskRequest {
  title: string;
  position?: number;
}

export interface UpdateSubtaskRequest {
  title?: string;
  is_completed?: boolean;
  position?: number;
}
