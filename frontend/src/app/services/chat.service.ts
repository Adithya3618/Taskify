import { Injectable } from '@angular/core';
import { Subject, Observable } from 'rxjs';
import { ChatMessage } from '../models/message.model';

@Injectable({
  providedIn: 'root'
})
export class ChatService {
  private socket: WebSocket | null = null;
  private messagesSubject = new Subject<ChatMessage>();
  private connectedSubject = new Subject<boolean>();
  private projectId: number | null = null;

  constructor() {}

  connect(projectId: number): void {
    this.projectId = projectId;
    
    if (this.socket) {
      this.socket.close();
    }
const proto = window.location.protocol === 'https:' ? 'wss' : 'ws';
const host = window.location.host;
this.socket = new WebSocket(`${proto}://${host}/ws/${projectId}`);
this.socket = new WebSocket(`ws://localhost:8080/ws/${projectId}`);

    this.socket.onopen = () => {
      console.log('WebSocket connected');
      this.connectedSubject.next(true);
    };

    this.socket.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data) as ChatMessage;
        this.messagesSubject.next(message);
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    };

    this.socket.onclose = () => {
      console.log('WebSocket disconnected');
      this.connectedSubject.next(false);
    };

    this.socket.onerror = (error) => {
      console.error('WebSocket error:', error);
      this.connectedSubject.next(false);
    };
  }

  disconnect(): void {
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
    this.projectId = null;
  }

  sendMessage(senderName: string, content: string): void {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      const message: ChatMessage = {
        type: 'chat',
        sender_name: senderName,
        content: content,
        created_at: new Date().toISOString()
      };
      this.socket.send(JSON.stringify(message));
    } else {
      console.error('WebSocket is not connected');
    }
  }

  getMessages(): Observable<ChatMessage> {
    return this.messagesSubject.asObservable();
  }

  getConnectionStatus(): Observable<boolean> {
    return this.connectedSubject.asObservable();
  }

  isConnected(): boolean {
    return this.socket !== null && this.socket.readyState === WebSocket.OPEN;
  }
}