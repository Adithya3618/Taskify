export interface Project {
  id: number;
  owner_id?: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
  member_count?: number;
}

export interface CreateProjectRequest {
  name: string;
  description: string;
}

export interface ProjectMember {
  id?: number;
  project_id: number;
  user_id: string;
  user_name?: string;
  user_email?: string;
  role: 'owner' | 'member';
  invited_by?: string;
  joined_at?: string;
  avatar?: string;
}

export interface AddProjectMemberRequest {
  user_id?: string;
  email?: string;
}
