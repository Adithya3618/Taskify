import { Component, inject, computed, signal } from '@angular/core';
import { BoardService } from '../../services/board.service';
import { Card } from '../../models/card.model';
import { NgFor, NgIf } from '@angular/common';
import { DatePipe } from '@angular/common';

export type CalendarViewMode = 'month' | 'week';

@Component({
  selector: 'app-calendar-view',
  standalone: true,
  imports: [NgFor, NgIf, DatePipe],
  templateUrl: './calendar-view.component.html',
  styleUrl: './calendar-view.component.scss',
})
export class CalendarViewComponent {
  private boardService = inject(BoardService);

  cards = this.boardService.currentBoardCards;
  currentDate = signal(new Date());
  viewMode = signal<CalendarViewMode>('month');

  private readonly MONTH_NAMES = ['January', 'February', 'March', 'April', 'May', 'June',
    'July', 'August', 'September', 'October', 'November', 'December'];

  /** Cards that have a due date, grouped by YYYY-MM-DD */
  cardsByDate = computed(() => {
    const map = new Map<string, Card[]>();
    for (const card of this.cards()) {
      const d = card.dueDate;
      if (!d) continue;
      const key = d.slice(0, 10);
      if (!map.has(key)) map.set(key, []);
      map.get(key)!.push(card);
    }
    return map;
  });

  /** Current month calendar grid: 6 rows x 7 days */
  calendarGrid = computed(() => {
    const d = this.currentDate();
    const year = d.getFullYear();
    const month = d.getMonth();
    const first = new Date(year, month, 1);
    const last = new Date(year, month + 1, 0);
    const startPad = first.getDay();
    const daysInMonth = last.getDate();
    const total = startPad + daysInMonth;
    const rows = Math.ceil(total / 7);
    const grid: { date: Date | null; day: number; key: string }[][] = [];
    let dayCount = 1;
    for (let r = 0; r < 6; r++) {
      const row: { date: Date | null; day: number; key: string }[] = [];
      for (let col = 0; col < 7; col++) {
        const cellIndex = r * 7 + col;
        if (cellIndex < startPad || dayCount > daysInMonth) {
          row.push({ date: null, day: 0, key: '' });
        } else {
          const date = new Date(year, month, dayCount);
          const key = date.toISOString().slice(0, 10);
          row.push({ date, day: dayCount, key });
          dayCount++;
        }
      }
      grid.push(row);
    }
    return grid;
  });

  monthLabel = computed(() => {
    const d = this.currentDate();
    return `${this.MONTH_NAMES[d.getMonth()]} ${d.getFullYear()}`;
  });

  prevMonth(): void {
    const d = new Date(this.currentDate());
    d.setMonth(d.getMonth() - 1);
    this.currentDate.set(d);
  }

  nextMonth(): void {
    const d = new Date(this.currentDate());
    d.setMonth(d.getMonth() + 1);
    this.currentDate.set(d);
  }

  today(): void {
    this.currentDate.set(new Date());
  }

  getListName(listId: string): string {
    return this.boardService.lists().find((l) => l.id === listId)?.title ?? listId;
  }

  openCard(cardId: string): void {
    this.boardService.openCardDetail(cardId);
  }
}
