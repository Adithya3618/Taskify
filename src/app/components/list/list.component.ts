import { Component, input, inject, computed } from '@angular/core';
import { CdkDropList, CdkDragDrop, moveItemInArray } from '@angular/cdk/drag-drop';
import { List } from '../../models/list.model';
import { Card } from '../../models/card.model';
import { BoardService } from '../../services/board.service';
import { CardComponent } from '../card/card.component';
import { NgFor, NgIf } from '@angular/common';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-list',
  standalone: true,
  imports: [CdkDropList, CardComponent, NgFor, NgIf, FormsModule],
  templateUrl: './list.component.html',
  styleUrl: './list.component.scss',
})
export class ListComponent {
  private boardService = inject(BoardService);
  list = input.required<List>();
  listIds = input<string[]>([]); // for cdkDropListConnectedTo

  cards = computed(() => {
    const list = this.list();
    this.boardService.cards(); // subscribe to cards changes
    return this.boardService.getCardsForList(list.id);
  });

  addingCard = false;
  newCardTitle = '';

  get listId(): string {
    return this.list().id;
  }

  get connectedListIds(): string[] {
    const ids = this.listIds();
    return ids.filter((id) => id !== this.listId);
  }

  onDrop(event: CdkDragDrop<Card[]>): void {
    const card = event.item.data as Card;
    const targetListId = event.container.id;
    const targetIndex = event.currentIndex;

    if (event.previousContainer === event.container) {
      const cards = this.boardService.getCardsForList(this.listId);
      const fromIndex = event.previousIndex;
      const reordered = [...cards];
      moveItemInArray(reordered, fromIndex, targetIndex);
      this.boardService.reorderCardsInList(
        this.listId,
        reordered.map((c) => c.id)
      );
    } else {
      this.boardService.moveCard(card.id, targetListId, targetIndex);
    }
  }

  startAddCard(): void {
    this.addingCard = true;
  }

  cancelAddCard(): void {
    this.addingCard = false;
    this.newCardTitle = '';
  }

  addCard(): void {
    const title = this.newCardTitle.trim();
    if (title) {
      this.boardService.addCard(this.listId, title);
      this.newCardTitle = '';
      this.addingCard = false;
    }
  }

  onDeleteCard(cardId: string): void {
    this.boardService.deleteCard(cardId);
  }

  updateListTitle(title: string): void {
    if (title.trim()) {
      this.boardService.updateList(this.listId, { title: title.trim() });
    }
  }

  /** Stop propagation so the click doesn’t bubble (e.g. to CDK drag) and only delete after confirm. */
  onListMenuClick(event: Event): void {
    event.preventDefault();
    event.stopPropagation();
    if (confirm(`Delete list "${this.list().title}"? This will remove all cards in it.`)) {
      this.boardService.deleteList(this.listId);
    }
  }
}
