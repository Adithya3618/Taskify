import { Comment } from './comment.model';
import { ChecklistItem } from './checklist.model';
import { Label } from './label.model';
import { Attachment } from './attachment.model';

export interface Card {
  id: string;
  title: string;
  description?: string;
  listId: string;
  order: number;
  color?: string;
  dueDate?: string;      // ISO date
  startDate?: string;    // ISO date (for Gantt)
  reminder?: string;     // ISO date-time
  labels?: Label[];
  assignees?: string[];  // display names for frontend
  comments?: Comment[];
  checklist?: ChecklistItem[];
  attachments?: Attachment[];
}
