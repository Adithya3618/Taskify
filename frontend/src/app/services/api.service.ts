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

  // Headers
  private getHeaders(): HttpHeaders {
    return new HttpHeaders({
      'Content-Type': 'application/json'
    });
  }

  // Project endpoints
  getProjects(): Observable<Project[]> {
    return this.http.get<Project[]>(`${this.baseUrl}/projects`, { headers: this.getHeaders() });
  }

  getProject(id: number): Observable<Project> {
    return this.http.get<Project>(`${this.baseUrl}/projects/${id}`, { headers: this.getHeaders() });
  }

  createProject(request: CreateProjectRequest): Observable<Project> {
    return this.http.post<Project>(`${this.baseUrl}/projects`, request, { headers: this.getHeaders() });
  }

  updateProject(id: number, request: CreateProjectRequest): Observable<Project> {
    return this.http.put<Project>(`${this.baseUrl}/projects/${id}`, request, { headers: this.getHeaders() });
  }

  deleteProject(id: number): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/projects/${id}`, { headers: this.getHeaders() });
  }

  // Stage endpoints
  getStages(projectId: number): Observable<Stage[]> {
    return this.http.get<Stage[]>(`${this.baseUrl}/projects/${projectId}/stages`, { headers: this.getHeaders() });
  }

  createStage(projectId: number, request: CreateStageRequest): Observable<Stage> {
    return this.http.post<Stage>(`${this.baseUrl}/projects/${projectId}/stages`, request, { headers: this.getHeaders() });
  }

  updateStage(id: number, request: { name: string; position: number }): Observable<Stage> {
    return this.http.put<Stage>(`${this.baseUrl}/stages/${id}`, request, { headers: this.getHeaders() });
  }

  deleteStage(id: number): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/stages/${id}`, { headers: this.getHeaders() });
  }

  // Task endpoints
  getTasks(stageId: number): Observable<Task[]> {
    return this.http.get<Task[]>(`${this.baseUrl}/stages/${stageId}/tasks`, { headers: this.getHeaders() });
  }

  createTask(stageId: number, request: CreateTaskRequest): Observable<Task> {
    return this.http.post<Task>(`${this.baseUrl}/stages/${stageId}/tasks`, request, { headers: this.getHeaders() });
  }

  updateTask(id: number, request: { title: string; description: string; position: number }): Observable<Task> {
    return this.http.put<Task>(`${this.baseUrl}/tasks/${id}`, request, { headers: this.getHeaders() });
  }

  moveTask(id: number, request: MoveTaskRequest): Observable<Task> {
    return this.http.put<Task>(`${this.baseUrl}/tasks/${id}/move`, request, { headers: this.getHeaders() });
  }

  deleteTask(id: number): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/tasks/${id}`, { headers: this.getHeaders() });
  }

  // Message endpoints
  getMessages(projectId: number): Observable<Message[]> {
    return this.http.get<Message[]>(`${this.baseUrl}/projects/${projectId}/messages`, { headers: this.getHeaders() });
  }

  getRecentMessages(projectId: number): Observable<Message[]> {
    return this.http.get<Message[]>(`${this.baseUrl}/projects/${projectId}/messages/recent`, { headers: this.getHeaders() });
  }

  createMessage(projectId: number, request: CreateMessageRequest): Observable<Message> {
    return this.http.post<Message>(`${this.baseUrl}/projects/${projectId}/messages`, request, { headers: this.getHeaders() });
  }

  deleteMessage(id: number): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/messages/${id}`, { headers: this.getHeaders() });
  }
}