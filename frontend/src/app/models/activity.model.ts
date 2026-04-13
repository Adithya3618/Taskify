export interface ActivityLog {
  id: number;
  project_id: number;
  user_id: string;
  user_name?: string;
  action: string;
  entity_type: string;
  entity_id?: number;
  description: string;
  created_at: string;
}

export interface ActivityQueryParams {
  page?: number;
  limit?: number;
  user_id?: string;
  from?: string;
  to?: string;
}
