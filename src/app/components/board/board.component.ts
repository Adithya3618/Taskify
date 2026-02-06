import { Component, inject, computed, signal } from '@angular/core';
import { BoardService } from '../../services/board.service';
import { ListComponent } from '../list/list.component';
import { CalendarViewComponent } from '../calendar-view/calendar-view.component';
import { TimelineViewComponent } from '../timeline-view/timeline-view.component';
import { DashboardViewComponent } from '../dashboard-view/dashboard-view.component';
import { TableViewComponent } from '../table-view/table-view.component';
import { CardDetailComponent } from '../card-detail/card-detail.component';
import { NgFor, NgIf } from '@angular/common';
import { FormsModule } from '@angular/forms';

export type BoardViewType = 'board' | 'calendar' | 'timeline' | 'dashboard' | 'table';

@Component({
  selector: 'app-board',
  standalone: true,
  imports: [
    ListComponent,
    CalendarViewComponent,
    TimelineViewComponent,
    DashboardViewComponent,
    TableViewComponent,
    CardDetailComponent,
    NgFor,
    NgIf,
    FormsModule,
  ],
  templateUrl: './board.component.html',
  styleUrl: './board.component.scss',
})
export class BoardComponent {
  private boardService = inject(BoardService);

  currentView = signal<BoardViewType>('board');
  views: { id: BoardViewType; label: string }[] = [
    { id: 'board', label: 'Board' },
    { id: 'calendar', label: 'Calendar' },
    { id: 'timeline', label: 'Timeline' },
    { id: 'dashboard', label: 'Dashboard' },
    { id: 'table', label: 'Table' },
  ];

  currentBoard = computed(() => {
    const id = this.boardService.currentBoardId();
    if (!id) return null;
    return this.boardService.boards().find((b) => b.id === id) ?? null;
  });

  lists = this.boardService.currentLists;

  listIds = computed(() => this.lists().map((l) => l.id));

  addingList = false;
  newListTitle = '';

  setView(view: BoardViewType): void {
    this.currentView.set(view);
  }

  startAddList(): void {
    this.addingList = true;
  }

  cancelAddList(): void {
    this.addingList = false;
    this.newListTitle = '';
  }

  addList(): void {
    const board = this.currentBoard();
    const title = this.newListTitle.trim();
    if (board && title) {
      this.boardService.addList(board.id, title);
      this.newListTitle = '';
      this.addingList = false;
    }
  }
}
