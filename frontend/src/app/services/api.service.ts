import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable, catchError, map, of, tap, throwError } from 'rxjs';

import { AddProjectMemberRequest, CreateProjectRequest, Project, ProjectMember } from '../models/project.model';
import { Stage, CreateStageRequest } from '../models/stage.model';
import { Task, CreateTaskRequest, MoveTaskRequest } from '../models/task.model';
import { Message, CreateMessageRequest } from '../models/message.model';
import { AuthService, AuthUser } from './auth.service';

interface ApiSuccessResponse<T> {
  success: boolean;
  data?: T;
  message?: string;
}

interface ApiPaginatedResponse<T> {
  success: boolean;
  data: T;
  page: number;
  limit: number;
  total: number;
}

@Injectable({
  providedIn: 'root'
})
export class ApiService {
  private baseUrl = '/api';
  private readonly memberStorePrefix = 'taskify.project.members.v1.';
  private readonly boardOwnersKey = 'taskify.board.owners';

  constructor(
    private http: HttpClient,
    private authService: AuthService
  ) {}

  // Use JSON headers only for requests that SEND a body (POST/PUT)
  private jsonHeaders(): HttpHeaders {
    return new HttpHeaders({ 'Content-Type': 'application/json' });
  }

  private memberStoreKey(projectId: number): string {
    return `${this.memberStorePrefix}${projectId}`;
  }

  private unwrapSuccess<T>() {
    return map((response: T | ApiSuccessResponse<T>) => {
      if (response && typeof response === 'object' && 'success' in response) {
        return (response as ApiSuccessResponse<T>).data as T;
      }
      return response as T;
    });
  }

  private unwrapPaginated<T>() {
    return map((response: T | ApiPaginatedResponse<T>) => {
      if (response && typeof response === 'object' && 'success' in response && 'data' in response) {
        return (response as ApiPaginatedResponse<T>).data;
      }
      return response as T;
    });
  }

  getCachedProjectMembers(projectId: number): ProjectMember[] {
    const raw = localStorage.getItem(this.memberStoreKey(projectId));
    if (!raw) return [];

    try {
      const parsed = JSON.parse(raw);
      return Array.isArray(parsed) ? parsed : [];
    } catch {
      return [];
    }
  }

  private saveCachedProjectMembers(projectId: number, members: ProjectMember[]): void {
    localStorage.setItem(this.memberStoreKey(projectId), JSON.stringify(members));
  }

  private getBoardOwnerEmail(projectId: number): string {
    try {
      const raw = localStorage.getItem(this.boardOwnersKey);
      const owners = raw ? JSON.parse(raw) as Record<string, string> : {};
      return owners[String(projectId)] || '';
    } catch {
      return '';
    }
  }

  private normalizeMember(projectId: number, member: Partial<ProjectMember>): ProjectMember {
    const email = member.user_email?.trim().toLowerCase() || '';
    const name = member.user_name?.trim() || email || 'Team member';

    return {
      id: member.id,
      project_id: projectId,
      user_id: member.user_id || email || `member-${Date.now()}`,
      user_name: name,
      user_email: email,
      role: member.role || 'member',
      invited_by: member.invited_by,
      joined_at: member.joined_at || new Date().toISOString(),
      avatar: name.charAt(0).toUpperCase()
    };
  }

  seedProjectOwner(projectId: number, project?: Project | null): void {
    const currentUser = this.authService.getCurrentUser();
    const boardOwnerEmail = this.getBoardOwnerEmail(projectId);
    const cachedMembers = this.getCachedProjectMembers(projectId);

    const ownerEmail = (
      boardOwnerEmail ||
      (project?.owner_id && currentUser?.id === project.owner_id ? currentUser.email : '') ||
      (cachedMembers.find((member) => member.role === 'owner')?.user_email) ||
      (currentUser?.email ?? '')
    ).trim().toLowerCase();

    const ownerUser =
      this.authService.getKnownUsers().find((user) => user.email.trim().toLowerCase() === ownerEmail) ||
      (currentUser && currentUser.email.trim().toLowerCase() === ownerEmail ? currentUser : null);

    if (!ownerEmail) return;

    const ownerMember = this.normalizeMember(projectId, {
      user_id: ownerUser?.id || project?.owner_id || ownerEmail,
      user_name: ownerUser?.name || ownerEmail,
      user_email: ownerEmail,
      role: 'owner'
    });

    const otherMembers = cachedMembers.filter(
      (member) => (member.user_email || '').trim().toLowerCase() !== ownerEmail
    );

    this.saveCachedProjectMembers(projectId, [ownerMember, ...otherMembers].map((member) => this.normalizeMember(projectId, member)));
  }

  userHasProjectAccess(projectId: number, email: string): boolean {
    const normalizedEmail = email.trim().toLowerCase();
    if (!normalizedEmail) return false;

    const cachedMembers = this.getCachedProjectMembers(projectId);
    if (cachedMembers.some((member) => (member.user_email || '').trim().toLowerCase() === normalizedEmail)) {
      return true;
    }

    return this.getBoardOwnerEmail(projectId) === normalizedEmail;
  }

  getProjectMemberCount(projectId: number): number {
    const members = this.getCachedProjectMembers(projectId);
    return members.length || 1;
  }

  syncProjectMemberCount(projects: Project[]): Project[] {
    return (projects || []).map((project) => ({
      ...project,
      member_count: this.getProjectMemberCount(project.id)
    }));
  }

  // ---------------- Projects ----------------
  getProjects(): Observable<Project[]> {
    return this.http.get<Project[]>(`${this.baseUrl}/projects`).pipe(
      map((projects) => {
        (projects || []).forEach((project) => this.seedProjectOwner(project.id, project));
        return this.syncProjectMemberCount(projects || []);
      })
    );
  }

  getProject(id: number): Observable<Project> {
    return this.http.get<Project>(`${this.baseUrl}/projects/${id}`).pipe(
      tap((project) => this.seedProjectOwner(id, project))
    );
  }

  createProject(request: CreateProjectRequest): Observable<Project> {
    return this.http.post<Project>(
      `${this.baseUrl}/projects`,
      request,
      { headers: this.jsonHeaders() }
    ).pipe(
      tap((project) => this.seedProjectOwner(project.id, project))
    );
  }

  updateProject(id: number, request: CreateProjectRequest): Observable<Project> {
    return this.http.put<Project>(
      `${this.baseUrl}/projects/${id}`,
      request,
      { headers: this.jsonHeaders() }
    );
  }

  deleteProject(id: number): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/projects/${id}`);
  }

  getProjectMembers(projectId: number): Observable<ProjectMember[]> {
    this.seedProjectOwner(projectId);

    return this.http.get<ApiPaginatedResponse<ProjectMember[]> | ProjectMember[]>(
      `${this.baseUrl}/projects/${projectId}/members`
    ).pipe(
      this.unwrapPaginated<ProjectMember[]>(),
      map((members) => (members || []).map((member) => this.normalizeMember(projectId, member))),
      tap((members) => {
        const cachedOwner = this.getCachedProjectMembers(projectId).find((member) => member.role === 'owner');
        const combinedMembers = cachedOwner
          ? [cachedOwner, ...members.filter((member) => member.user_id !== cachedOwner.user_id)]
          : members;
        this.saveCachedProjectMembers(projectId, combinedMembers);
      }),
      catchError(() => of(this.getCachedProjectMembers(projectId)))
    );
  }

  addProjectMember(projectId: number, request: AddProjectMemberRequest): Observable<ProjectMember> {
    this.seedProjectOwner(projectId);
    const cachedMembers = this.getCachedProjectMembers(projectId);
    const normalizedEmail = request.email?.trim().toLowerCase() || '';
    const duplicate = cachedMembers.find((member) => {
      const memberEmail = (member.user_email || '').trim().toLowerCase();
      return (normalizedEmail && memberEmail === normalizedEmail) || (!!request.user_id && member.user_id === request.user_id);
    });

    if (duplicate) {
      return throwError(() => ({ status: 409, error: { error: 'This member is already on the project.' } }));
    }

    const knownUser = this.authService.getKnownUsers().find((user) => {
      const emailMatches = normalizedEmail && user.email.trim().toLowerCase() === normalizedEmail;
      const idMatches = !!request.user_id && user.id === request.user_id;
      return emailMatches || idMatches;
    });

    const fallbackMember = this.normalizeMember(projectId, {
      user_id: request.user_id || knownUser?.id || normalizedEmail || `member-${Date.now()}`,
      user_name: knownUser?.name || normalizedEmail || 'Team member',
      user_email: normalizedEmail || knownUser?.email,
      role: 'member',
      joined_at: new Date().toISOString()
    });

    return this.http.post<ApiSuccessResponse<ProjectMember> | ProjectMember>(
      `${this.baseUrl}/projects/${projectId}/members`,
      request,
      { headers: this.jsonHeaders() }
    ).pipe(
      this.unwrapSuccess<ProjectMember>(),
      map((member) => this.normalizeMember(projectId, member || fallbackMember)),
      tap((member) => {
        this.saveCachedProjectMembers(projectId, [...cachedMembers, member]);
      }),
      catchError((error) => {
        if (error?.status === 409) {
          return throwError(() => error);
        }

        this.saveCachedProjectMembers(projectId, [...cachedMembers, fallbackMember]);
        return of(fallbackMember);
      })
    );
  }

  removeProjectMember(projectId: number, userId: string): Observable<void> {
    const cachedMembers = this.getCachedProjectMembers(projectId);

    return this.http.delete<ApiSuccessResponse<null> | void>(
      `${this.baseUrl}/projects/${projectId}/members/${userId}`
    ).pipe(
      map(() => void 0),
      tap(() => {
        this.saveCachedProjectMembers(projectId, cachedMembers.filter((member) => member.user_id !== userId));
      }),
      catchError(() => {
        this.saveCachedProjectMembers(projectId, cachedMembers.filter((member) => member.user_id !== userId));
        return of(void 0);
      })
    );
  }

  // ---------------- Stages ----------------
  getStages(projectId: number): Observable<Stage[]> {
    return this.http.get<Stage[]>(`${this.baseUrl}/projects/${projectId}/stages`);
  }

  createStage(projectId: number, request: CreateStageRequest): Observable<Stage> {
    return this.http.post<Stage>(
      `${this.baseUrl}/projects/${projectId}/stages`,
      request,
      { headers: this.jsonHeaders() }
    );
  }

  updateStage(id: number, request: { name: string; position: number }): Observable<Stage> {
    return this.http.put<Stage>(
      `${this.baseUrl}/stages/${id}`,
      request,
      { headers: this.jsonHeaders() }
    );
  }

  deleteStage(id: number): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/stages/${id}`);
  }

  // ---------------- Tasks ----------------
  getTasks(projectId: number, stageId: number): Observable<Task[]> {
  return this.http.get<Task[]>(
    `${this.baseUrl}/projects/${projectId}/stages/${stageId}/tasks`
  );
}

createTask(projectId: number, stageId: number, request: CreateTaskRequest): Observable<Task> {
  return this.http.post<Task>(
    `${this.baseUrl}/projects/${projectId}/stages/${stageId}/tasks`,
    request,
    { headers: this.jsonHeaders() }
  );
}

  updateTask(id: number, request: { title: string; description: string; position: number }): Observable<Task> {
    return this.http.put<Task>(
      `${this.baseUrl}/tasks/${id}`,
      request,
      { headers: this.jsonHeaders() }
    );
  }

  moveTask(id: number, request: MoveTaskRequest): Observable<Task> {
    return this.http.put<Task>(
      `${this.baseUrl}/tasks/${id}/move`,
      request,
      { headers: this.jsonHeaders() }
    );
  }

  deleteTask(id: number): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/tasks/${id}`);
  }

  // ---------------- Messages ----------------
  getMessages(projectId: number): Observable<Message[]> {
    return this.http.get<Message[]>(`${this.baseUrl}/projects/${projectId}/messages`);
  }

  getRecentMessages(projectId: number): Observable<Message[]> {
    return this.http.get<Message[]>(`${this.baseUrl}/projects/${projectId}/messages/recent`);
  }

  createMessage(projectId: number, request: CreateMessageRequest): Observable<Message> {
    return this.http.post<Message>(
      `${this.baseUrl}/projects/${projectId}/messages`,
      request,
      { headers: this.jsonHeaders() }
    );
  }

  deleteMessage(id: number): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/messages/${id}`);
  }
}
