import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable, catchError, map, of, tap, throwError } from 'rxjs';

import { AddProjectMemberRequest, CreateProjectRequest, Project, ProjectMember } from '../models/project.model';
import { Comment, CreateCommentRequest } from '../models/comment.model';
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
  private readonly commentStorePrefix = 'taskify.task.comments.v1.';
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

  private commentStoreKey(taskId: number): string {
    return `${this.commentStorePrefix}${taskId}`;
  }

  getCachedTaskComments(taskId: number): Comment[] {
    const raw = localStorage.getItem(this.commentStoreKey(taskId));
    if (!raw) return [];

    try {
      const parsed = JSON.parse(raw);
      return Array.isArray(parsed) ? parsed : [];
    } catch {
      return [];
    }
  }

  private saveCachedTaskComments(taskId: number, comments: Comment[]): void {
    localStorage.setItem(this.commentStoreKey(taskId), JSON.stringify(comments));
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

  private normalizeComment(taskId: number, comment: Partial<Comment>): Comment {
    const currentUser = this.authService.getCurrentUser();

    return {
      id: comment.id ?? `local-${Date.now()}`,
      task_id: comment.task_id ?? taskId,
      user_id: comment.user_id || currentUser?.id || currentUser?.email || 'local-user',
      author_name: comment.author_name || currentUser?.name || 'You',
      content: comment.content?.trim() || '',
      created_at: comment.created_at || new Date().toISOString(),
      updated_at: comment.updated_at || comment.created_at || new Date().toISOString(),
      isLocalOnly: !!comment.isLocalOnly
    };
  }

  private mergeComments(taskId: number, incomingComments: Comment[], existingComments: Comment[] = []): Comment[] {
    const byKey = new Map<string, Comment>();

    [...existingComments, ...incomingComments]
      .map((comment) => this.normalizeComment(taskId, comment))
      .forEach((comment) => {
        const key = String(comment.id);
        const previous = byKey.get(key);
        byKey.set(key, previous ? { ...previous, ...comment } : comment);
      });

    return [...byKey.values()].sort((a, b) => {
      const createdDiff = new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
      if (createdDiff !== 0) return createdDiff;
      return String(a.id).localeCompare(String(b.id));
    });
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
    const cachedMembers = this.getCachedProjectMembers(projectId);

    return this.http.get<ApiPaginatedResponse<ProjectMember[]> | ProjectMember[]>(
      `${this.baseUrl}/projects/${projectId}/members`
    ).pipe(
      this.unwrapPaginated<ProjectMember[]>(),
      map((members) => (members || []).map((member) => this.normalizeMember(projectId, member))),
      tap((members) => {
        const mergedMembers = this.mergeMembers(projectId, cachedMembers, members || []);
        this.saveCachedProjectMembers(projectId, mergedMembers);
      }),
      map((members) => this.mergeMembers(projectId, cachedMembers, members || [])),
      catchError(() => of(cachedMembers))
    );
  }

  private mergeMembers(projectId: number, existingMembers: ProjectMember[], nextMembers: ProjectMember[]): ProjectMember[] {
    const byKey = new Map<string, ProjectMember>();

    [...existingMembers, ...nextMembers]
      .map((member) => this.normalizeMember(projectId, member))
      .forEach((member) => {
        const key = (member.user_email || member.user_id).trim().toLowerCase();
        const previous = byKey.get(key);
        byKey.set(key, previous ? { ...previous, ...member } : member);
      });

    return [...byKey.values()].sort((a, b) => {
      if (a.role !== b.role) return a.role === 'owner' ? -1 : 1;
      return (a.user_name || a.user_email || '').localeCompare(b.user_name || b.user_email || '');
    });
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
        this.saveCachedProjectMembers(projectId, this.mergeMembers(projectId, cachedMembers, [member]));
      }),
      catchError((error) => {
        if (error?.status === 409) {
          return throwError(() => error);
        }

        this.saveCachedProjectMembers(projectId, this.mergeMembers(projectId, cachedMembers, [fallbackMember]));
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

  // ---------------- Comments ----------------
  getComments(taskId: number): Observable<Comment[]> {
    const cachedComments = this.getCachedTaskComments(taskId);

    return this.http.get<Comment[]>(`${this.baseUrl}/tasks/${taskId}/comments`).pipe(
      map((comments) => this.mergeComments(taskId, comments || [], cachedComments)),
      tap((comments) => this.saveCachedTaskComments(taskId, comments)),
      catchError(() => of(cachedComments))
    );
  }

  createComment(taskId: number, request: CreateCommentRequest): Observable<Comment> {
    const currentUser = this.authService.getCurrentUser();
    const fallbackComment = this.normalizeComment(taskId, {
      id: `local-${Date.now()}`,
      task_id: taskId,
      user_id: currentUser?.id || currentUser?.email || 'local-user',
      author_name: currentUser?.name || 'You',
      content: request.content,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      isLocalOnly: true
    });

    return this.http.post<Comment>(
      `${this.baseUrl}/tasks/${taskId}/comments`,
      request,
      { headers: this.jsonHeaders() }
    ).pipe(
      map((comment) => this.normalizeComment(taskId, comment)),
      tap((comment) => {
        const merged = this.mergeComments(taskId, [comment], this.getCachedTaskComments(taskId));
        this.saveCachedTaskComments(taskId, merged);
      }),
      catchError(() => {
        const merged = this.mergeComments(taskId, [fallbackComment], this.getCachedTaskComments(taskId));
        this.saveCachedTaskComments(taskId, merged);
        return of(fallbackComment);
      })
    );
  }

  updateComment(commentId: number | string, taskId: number, request: CreateCommentRequest): Observable<Comment> {
    const cachedComments = this.getCachedTaskComments(taskId);
    const target = cachedComments.find((comment) => String(comment.id) === String(commentId));
    const fallbackComment = this.normalizeComment(taskId, {
      ...target,
      id: commentId,
      content: request.content,
      updated_at: new Date().toISOString(),
      isLocalOnly: target?.isLocalOnly ?? true
    });

    return this.http.patch<Comment>(
      `${this.baseUrl}/comments/${commentId}`,
      request,
      { headers: this.jsonHeaders() }
    ).pipe(
      map((comment) => this.normalizeComment(taskId, comment)),
      tap((comment) => {
        const merged = this.mergeComments(taskId, [comment], cachedComments);
        this.saveCachedTaskComments(taskId, merged);
      }),
      catchError(() => {
        const merged = this.mergeComments(taskId, [fallbackComment], cachedComments);
        this.saveCachedTaskComments(taskId, merged);
        return of(fallbackComment);
      })
    );
  }

  deleteComment(commentId: number | string, taskId: number): Observable<void> {
    const cachedComments = this.getCachedTaskComments(taskId);
    const nextComments = cachedComments.filter((comment) => String(comment.id) !== String(commentId));

    return this.http.delete<void>(`${this.baseUrl}/comments/${commentId}`).pipe(
      tap(() => this.saveCachedTaskComments(taskId, nextComments)),
      map(() => void 0),
      catchError(() => {
        this.saveCachedTaskComments(taskId, nextComments);
        return of(void 0);
      })
    );
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
