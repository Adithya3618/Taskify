import { Component, inject, computed, signal } from '@angular/core';
import { BoardService } from '../../services/board.service';
import { Card } from '../../models/card.model';
import { NgFor, NgIf } from '@angular/common';

@Component({
  selector: 'app-timeline-view',
  standalone: true,
  imports: [NgFor, NgIf],
  templateUrl: './timeline-view.component.html',
  styleUrl: './timeline-view.component.scss',
})
export class TimelineViewComponent {
  private boardService = inject(BoardService);

  cards = this.boardService.currentBoardCards;
  /** Timeline range: start and end of visible range (dates) */
  rangeStart = signal(new Date());
  rangeEnd = signal(new Date(Date.now() + 30 * 24 * 60 * 60 * 1000));
  /** Pixels per day for horizontal scale */
  pxPerDay = signal(24);

  /** Cards that have startDate or dueDate for Gantt bars */
  timelineCards = computed(() => {
    const list = this.boardService.currentLists();
    const cards = this.cards().filter((c) => c.dueDate || c.startDate);
    return cards.map((c) => {
      const start = c.startDate ? new Date(c.startDate) : (c.dueDate ? new Date(c.dueDate) : new Date());
      const end = c.dueDate ? new Date(c.dueDate) : new Date(start.getTime() + 24 * 60 * 60 * 1000);
      const listTitle = list.find((l) => l.id === c.listId)?.title ?? c.listId;
      return { card: c, start, end, listTitle };
    });
  });

  /** Timeline header: array of dates from rangeStart to rangeEnd */
  timelineDays = computed(() => {
    const start = this.rangeStart();
    const end = this.rangeEnd();
    const days: Date[] = [];
    const d = new Date(start);
    d.setHours(0, 0, 0, 0);
    const e = new Date(end);
    e.setHours(0, 0, 0, 0);
    while (d <= e) {
      days.push(new Date(d));
      d.setDate(d.getDate() + 1);
    }
    return days;
  });

  /** Base date for pixel offset (start of range) */
  baseTime = computed(() => this.rangeStart().getTime());

  getBarLeft(item: { start: Date }): number {
    const base = this.baseTime();
    const t = item.start.getTime();
    const days = (t - base) / (24 * 60 * 60 * 1000);
    return Math.max(0, days * this.pxPerDay());
  }

  getBarWidth(item: { start: Date; end: Date }): number {
    const start = item.start.getTime();
    const end = item.end.getTime();
    const days = Math.max(1, (end - start) / (24 * 60 * 60 * 1000));
    return days * this.pxPerDay();
  }

  getListName(listId: string): string {
    return this.boardService.lists().find((l) => l.id === listId)?.title ?? listId;
  }

  prevRange(): void {
    const days = 14;
    this.rangeStart.update((d) => new Date(d.getTime() - days * 24 * 60 * 60 * 1000));
    this.rangeEnd.update((d) => new Date(d.getTime() - days * 24 * 60 * 60 * 1000));
  }

  nextRange(): void {
    const days = 14;
    this.rangeStart.update((d) => new Date(d.getTime() + days * 24 * 60 * 60 * 1000));
    this.rangeEnd.update((d) => new Date(d.getTime() + days * 24 * 60 * 60 * 1000));
  }

  formatDate(d: Date): string {
    return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
  }

  openCard(cardId: string): void {
    this.boardService.openCardDetail(cardId);
  }
}
