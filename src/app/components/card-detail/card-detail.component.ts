import { Component, inject, computed, signal } from '@angular/core';
import { BoardService } from '../../services/board.service';
import { Card } from '../../models/card.model';
import { NgFor, NgIf } from '@angular/common';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-card-detail',
  standalone: true,
  imports: [NgFor, NgIf, FormsModule],
  templateUrl: './card-detail.component.html',
  styleUrl: './card-detail.component.scss',
})
export class CardDetailComponent {
  private boardService = inject(BoardService);

  selectedCardId = this.boardService.selectedCardId;
  card = computed(() => {
    const id = this.boardService.selectedCardId();
    return id ? this.boardService.getCard(id) : null;
  });

  descriptionEdit = false;
  descriptionText = '';
  newCommentText = '';
  commentAuthor = 'You';
  newChecklistTitle = '';
  newAssigneeName = '';
  newAttachmentName = '';
  newAttachmentUrl = '';
  newAttachmentType: 'link' | 'drive' | 'dropbox' | 'file' = 'link';

  getListName(listId: string): string {
    return this.boardService.lists().find((l) => l.id === listId)?.title ?? listId;
  }

  close(): void {
    this.boardService.closeCardDetail();
  }

  startEditDescription(): void {
    const c = this.card();
    this.descriptionText = c?.description ?? '';
    this.descriptionEdit = true;
  }

  saveDescription(): void {
    const c = this.card();
    if (!c) return;
    this.boardService.updateCard(c.id, { description: this.descriptionText.trim() || undefined });
    this.descriptionEdit = false;
  }

  addComment(): void {
    const c = this.card();
    const text = this.newCommentText.trim();
    if (!c || !text) return;
    this.boardService.addComment(c.id, this.commentAuthor, text);
    this.newCommentText = '';
  }

  deleteComment(commentId: string): void {
    const c = this.card();
    if (!c) return;
    this.boardService.deleteComment(c.id, commentId);
  }

  addChecklistItem(): void {
    const c = this.card();
    const title = this.newChecklistTitle.trim();
    if (!c || !title) return;
    this.boardService.addChecklistItem(c.id, title);
    this.newChecklistTitle = '';
  }

  toggleChecklistItem(itemId: string): void {
    const c = this.card();
    if (!c) return;
    this.boardService.toggleChecklistItem(c.id, itemId);
  }

  deleteChecklistItem(itemId: string): void {
    const c = this.card();
    if (!c) return;
    this.boardService.deleteChecklistItem(c.id, itemId);
  }

  addAssignee(): void {
    const c = this.card();
    const name = this.newAssigneeName.trim();
    if (!c || !name) return;
    this.boardService.addAssignee(c.id, name);
    this.newAssigneeName = '';
  }

  removeAssignee(name: string): void {
    const c = this.card();
    if (!c) return;
    this.boardService.removeAssignee(c.id, name);
  }

  addAttachment(): void {
    const c = this.card();
    const name = this.newAttachmentName.trim();
    const url = this.newAttachmentUrl.trim();
    if (!c || !name || !url) return;
    this.boardService.addAttachment(c.id, name, url, this.newAttachmentType);
    this.newAttachmentName = '';
    this.newAttachmentUrl = '';
  }

  removeAttachment(attId: string): void {
    const c = this.card();
    if (!c) return;
    this.boardService.removeAttachment(c.id, attId);
  }

  formatDate(iso: string): string {
    return new Date(iso).toLocaleDateString(undefined, { dateStyle: 'medium' });
  }

  formatDateTime(iso: string): string {
    return new Date(iso).toLocaleString(undefined, { dateStyle: 'short', timeStyle: 'short' });
  }

  getChecklistDoneCount(card: Card): number {
    return (card.checklist ?? []).filter((i) => i.completed).length;
  }
}
