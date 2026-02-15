import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable } from 'rxjs';

import { Project, CreateProjectRequest } from '../models/project.model';
import { Stage, CreateStageRequest } from '../models/stage.model';
import { Task, CreateTaskRequest, MoveTaskRequest } from '../models/task.model';
import { Message, CreateMessageRequest } from '../models/message.model';

@Injectable({
  providedIn: 'root'
})
export class ApiService {
  private baseUrl = 'http://localhost:8080/api';

  constructor(private http: HttpClient) {}

  // Use JSON headers only for requests that SEND a body (POST/PUT)
  private jsonHeaders(): HttpHeaders {
    return new HttpHeaders({ 'Content-Type': 'application/json' });
  }

  // ---------------- Projects ----------------
  getProjects(): Observable<Project[]> {
    return this.http.get<Project[]>(`${this.baseUrl}/projects`);
  }

  getProject(id: number): Observable<Project> {
    return this.http.get<Project>(`${this.baseUrl}/projects/${id}`);
  }

  createProject(request: CreateProjectRequest): Observable<Project> {
    return this.http.post<Project>(
      `${this.baseUrl}/projects`,
      request,
      { headers: this.jsonHeaders() }
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
  getTasks(stageId: number): Observable<Task[]> {
    return this.http.get<Task[]>(`${this.baseUrl}/stages/${stageId}/tasks`);
  }

  createTask(stageId: number, request: CreateTaskRequest): Observable<Task> {
    return this.http.post<Task>(
      `${this.baseUrl}/stages/${stageId}/tasks`,
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
