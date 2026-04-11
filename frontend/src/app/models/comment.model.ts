export interface Comment {
  id: number | string;
  task_id: number;
  user_id: string;
  author_name: string;
  content: string;
  created_at: string;
  updated_at: string;
  isLocalOnly?: boolean;
}

export interface CreateCommentRequest {
  content: string;
}
