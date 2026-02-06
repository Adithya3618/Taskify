import { Component, input, output, inject, computed } from '@angular/core';
import { Card } from '../../models/card.model';
import { BoardService } from '../../services/board.service';
import { CdkDrag } from '@angular/cdk/drag-drop';
import { NgIf } from '@angular/common';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-card',
  standalone: true,
  imports: [CdkDrag, NgIf, FormsModule],
  templateUrl: './card.component.html',
  styleUrl: './card.component.scss',
})
export class CardComponent {
  private boardService = inject(BoardService);
  card = input.required<Card>();
  deleteCard = output<string>();
  editMode = false;
  title = '';
  showDatePicker = false;
  dueDateInput = '';

  mailtoLink = computed(() => {
    const c = this.card();
    const subject = encodeURIComponent(c.title);
    const body = encodeURIComponent(
      (c.description ? `${c.title}\n\n${c.description}` : c.title) +
        (c.dueDate ? `\n\nDue: ${this.formatDueDate(c.dueDate)}` : '')
    );
    return `mailto:?subject=${subject}&body=${body}`;
  });

  onEdit(): void {
    this.editMode = true;
    this.title = this.card().title;
  }

  onSave(): void {
    this.editMode = false;
    if (this.title.trim()) {
      this.boardService.updateCard(this.card().id, { title: this.title.trim() });
    }
  }

  onDelete(): void {
    this.deleteCard.emit(this.card().id);
  }

  openDetail(): void {
    this.boardService.openCardDetail(this.card().id);
  }

  toggleDatePicker(): void {
    this.showDatePicker = !this.showDatePicker;
    this.dueDateInput = this.card().dueDate || '';
  }

  setDueDate(): void {
    this.boardService.updateCard(this.card().id, { dueDate: this.dueDateInput || undefined });
    this.showDatePicker = false;
    this.dueDateInput = '';
  }

  clearDueDate(): void {
    this.boardService.updateCard(this.card().id, { dueDate: undefined });
    this.showDatePicker = false;
    this.dueDateInput = '';
  }

  formatDueDate(iso: string): string {
    const d = new Date(iso);
    return d.toLocaleDateString(undefined, { dateStyle: 'medium' });
  }

  isOverdue(iso: string): boolean {
    const due = new Date(iso);
    due.setHours(23, 59, 59, 999);
    return due < new Date();
  }

  checklistCount(): number {
    return (this.card().checklist ?? []).length;
  }

  checklistDoneCount(): number {
    return (this.card().checklist ?? []).filter((i) => i.completed).length;
  }

  assigneesCount(): number {
    return (this.card().assignees ?? []).length;
  }

  hasBadges(): boolean {
    return this.checklistCount() > 0 || this.assigneesCount() > 0;
  }
}
