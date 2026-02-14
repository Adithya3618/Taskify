import { Task } from './task.model';

export interface Stage {
  id: number;
  project_id: number;
  name: string;
  position: number;
  created_at: string;
  updated_at: string;
  tasks?: Task[];
}

export interface CreateStageRequest {
  name: string;
  position: number;
}

export interface UpdateStageRequest {
  name: string;
  position: number;
}