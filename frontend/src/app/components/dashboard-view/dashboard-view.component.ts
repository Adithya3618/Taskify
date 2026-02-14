import { Component, inject, computed } from '@angular/core';
import { BoardService } from '../../services/board.service';
import { Card } from '../../models/card.model';
import { NgFor, NgIf } from '@angular/common';

@Component({
  selector: 'app-dashboard-view',
  standalone: true,
  imports: [NgFor, NgIf],
  templateUrl: './dashboard-view.component.html',
  styleUrl: './dashboard-view.component.scss',
})
export class DashboardViewComponent {
  private boardService = inject(BoardService);

  lists = this.boardService.currentLists;
  cards = this.boardService.currentBoardCards;

  totalCards = computed(() => this.cards().length);
  totalLists = computed(() => this.lists().length);

  cardsWithDueDate = computed(() =>
    this.cards().filter((c) => c.dueDate)
  );
  overdueCount = computed(() =>
    this.cardsWithDueDate().filter((c) => {
      if (!c.dueDate) return false;
      const d = new Date(c.dueDate);
      d.setHours(23, 59, 59, 999);
      return d < new Date();
    }).length
  );
  dueSoonCount = computed(() => {
    const in7 = new Date();
    in7.setDate(in7.getDate() + 7);
    return this.cardsWithDueDate().filter((c) => {
      if (!c.dueDate) return false;
      const d = new Date(c.dueDate);
      return d >= new Date() && d <= in7;
    }).length;
  });

  checklistProgress = computed(() => {
    let total = 0;
    let done = 0;
    for (const card of this.cards()) {
      const list = card.checklist ?? [];
      total += list.length;
      done += list.filter((i) => i.completed).length;
    }
    return total === 0 ? 0 : Math.round((done / total) * 100);
  });

  listProgress = computed(() =>
    this.lists().map((list) => {
      const listCards = this.cards().filter((c) => c.listId === list.id);
      const total = listCards.length;
      const doneList = this.lists().find((l) => l.title.toLowerCase().includes('done'));
      const done = doneList ? listCards.filter((c) => c.listId === doneList.id).length : 0;
      const pct = total === 0 ? 0 : Math.round((listCards.length / this.totalCards()) * 100);
      return { list, count: listCards.length, pct };
    })
  );

  getListName(listId: string): string {
    return this.boardService.lists().find((l) => l.id === listId)?.title ?? listId;
  }

  isOverdue(iso: string): boolean {
    const d = new Date(iso);
    d.setHours(23, 59, 59, 999);
    return d < new Date();
  }

  formatDate(iso: string): string {
    return new Date(iso).toLocaleDateString(undefined, { dateStyle: 'medium' });
  }
}
