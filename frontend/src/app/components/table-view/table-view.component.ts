import { Component, inject, computed } from '@angular/core';
import { BoardService } from '../../services/board.service';
import { Card } from '../../models/card.model';
import { NgFor } from '@angular/common';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-table-view',
  standalone: true,
  imports: [NgFor, FormsModule],
  templateUrl: './table-view.component.html',
  styleUrl: './table-view.component.scss',
})
export class TableViewComponent {
  private boardService = inject(BoardService);

  lists = this.boardService.currentLists;
  cards = this.boardService.currentBoardCards;

  /** All cards for the table, sorted by list order then card order */
  tableRows = computed(() => {
    const listOrder = this.lists();
    const cards = this.cards();
    return [...cards].sort((a, b) => {
      const ai = listOrder.findIndex((l) => l.id === a.listId);
      const bi = listOrder.findIndex((l) => l.id === b.listId);
      if (ai !== bi) return ai - bi;
      return a.order - b.order;
    });
  });

  getListName(listId: string): string {
    return this.boardService.lists().find((l) => l.id === listId)?.title ?? listId;
  }

  formatDate(iso: string | undefined): string {
    if (!iso) return '—';
    return new Date(iso).toLocaleDateString(undefined, { dateStyle: 'short' });
  }

  labelsDisplay(card: Card): string {
    const labels = card.labels ?? [];
    return labels.map((l) => l.name).join(', ') || '—';
  }

  assigneesDisplay(card: Card): string {
    const a = card.assignees ?? [];
    return a.length ? a.join(', ') : '—';
  }

  checklistDisplay(card: Card): string {
    const list = card.checklist ?? [];
    if (list.length === 0) return '—';
    const done = list.filter((i) => i.completed).length;
    return `${done}/${list.length}`;
  }

  openCard(cardId: string): void {
    this.boardService.openCardDetail(cardId);
  }
}
