export interface Message {
  id: number;
  project_id: number;
  sender_name: string;
  content: string;
  created_at: string;
}

export interface CreateMessageRequest {
  sender_name: string;
  content: string;
}

export interface ChatMessage {
  type: string;
  sender_name: string;
  content: string;
  created_at: string;
}