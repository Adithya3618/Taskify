import { Injectable, signal, computed, inject } from '@angular/core';
import { Board } from '../models/board.model';
import { List } from '../models/list.model';
import { Card } from '../models/card.model';
import { Comment } from '../models/comment.model';
import { ChecklistItem } from '../models/checklist.model';
import { Label } from '../models/label.model';
import { Attachment } from '../models/attachment.model';
import { ApiService } from './api.service';
import { Stage } from '../models/stage.model';

@Injectable({ providedIn: 'root' })
export class BoardService {
  private apiService = inject(ApiService);
  
  private boardsSignal = signal<Board[]>([]);
  private listsSignal = signal<List[]>([]);
  private cardsSignal = signal<Card[]>([]);

  private currentBoardIdSignal = signal<string | null>(null);
  private selectedCardIdSignal = signal<string | null>(null);

  boards = this.boardsSignal.asReadonly();
  selectedCardId = this.selectedCardIdSignal.asReadonly();
  lists = this.listsSignal.asReadonly();
  cards = this.cardsSignal.asReadonly();
  currentBoardId = this.currentBoardIdSignal.asReadonly();

  currentLists = computed(() => {
    const boardId = this.currentBoardIdSignal();
    const lists = this.listsSignal();
    if (!boardId) return [];
    return [...lists.filter((l) => l.boardId === boardId)].sort((a, b) => a.order - b.order);
  });

  currentBoardCards = computed(() => {
    const boardId = this.currentBoardIdSignal();
    const lists = this.currentLists();
    const cards = this.cardsSignal();
    if (!boardId) return [];
    const listIds = new Set(lists.map((l) => l.id));
    return cards.filter((c) => listIds.has(c.listId)).sort((a, b) => a.order - b.order);
  });

  constructor() {
    this.seedData();
  }

  private seedData(): void {
    // Use numeric IDs to match backend
    const boardId = '1';
    const board: Board = { id: boardId, title: "Meg's Trel", color: '#0079bf' };
    this.boardsSignal.set([board]);
    this.currentBoardIdSignal.set(board.id);

    const lists: List[] = [
      { id: '1', title: 'To Do', boardId: boardId, order: 0 },
      { id: '2', title: 'In Progress', boardId: boardId, order: 1 },
      { id: '3', title: 'Done', boardId: boardId, order: 2 },
    ];
    this.listsSignal.set(lists);

    const cards: Card[] = [
      {
        id: '1',
        title: 'Setup project',
        listId: '1',
        order: 0,
        description: 'Initialize repo and dependencies.',
        dueDate: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString().slice(0, 10),
        startDate: new Date().toISOString().slice(0, 10),
        labels: [{ id: 'l1', name: 'Dev', color: '#61bd4f' }],
        assignees: ['Alex'],
        checklist: [
          { id: 'ch1', cardId: '1', title: 'Create repo', completed: true, order: 0 },
          { id: 'ch2', cardId: '1', title: 'Install deps', completed: false, order: 1 },
        ],
        comments: [],
        attachments: [],
      },
      {
        id: '2',
        title: 'Design UI',
        listId: '1',
        order: 1,
        dueDate: new Date(Date.now() + 14 * 24 * 60 * 60 * 1000).toISOString().slice(0, 10),
        startDate: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString().slice(0, 10),
        labels: [{ id: 'l2', name: 'Design', color: '#f2d600' }],
        assignees: [],
        checklist: [],
        comments: [],
        attachments: [],
      },
      {
        id: '3',
        title: 'Implement drag & drop',
        listId: '2',
        order: 0,
        dueDate: new Date(Date.now() + 5 * 24 * 60 * 60 * 1000).toISOString().slice(0, 10),
        startDate: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString().slice(0, 10),
        labels: [{ id: 'l1', name: 'Dev', color: '#61bd4f' }],
        assignees: ['Sam'],
        checklist: [],
        comments: [],
        attachments: [],
      },
      {
        id: '4',
        title: 'Add new list',
        listId: '3',
        order: 0,
        labels: [],
        assignees: [],
        checklist: [],
        comments: [],
        attachments: [],
      },
    ];
    this.cardsSignal.set(cards);
  }

  setCurrentBoard(id: string | null): void {
    this.currentBoardIdSignal.set(id);
  }

  getCardsForList(listId: string): Card[] {
    return [...this.cardsSignal().filter((c) => c.listId === listId)].sort((a, b) => a.order - b.order);
  }

  addBoard(title: string): Board {
    const board: Board = {
      id: `board-${Date.now()}`,
      title: title || 'Untitled Board',
    };
    this.boardsSignal.update((b) => [...b, board]);
    return board;
  }

  addList(boardId: string, title: string): void {
    if (!title.trim()) return;
    
    const lists = this.listsSignal().filter((l) => l.boardId === boardId);
    const order = lists.length;
    const projectId = parseInt(boardId, 10);
    
    if (isNaN(projectId)) {
      console.error('Invalid project ID:', boardId);
      return;
    }
    
    this.apiService.createStage(projectId, { name: title, position: order }).subscribe({
      next: (stage: Stage) => {
        const list: List = {
          id: String(stage.id),
          title: stage.name,
          boardId,
          order: stage.position,
        };
        this.listsSignal.update((l) => [...l, list]);
      },
      error: (err) => console.error('Failed to create list:', err)
    });
  }

  addCard(listId: string, title: string): Card {
    const cards = this.cardsSignal().filter((c) => c.listId === listId);
    const order = cards.length;
    const card: Card = {
      id: `card-${Date.now()}`,
      title: title || 'New Card',
      listId,
      order,
      comments: [],
      checklist: [],
      labels: [],
      attachments: [],
      assignees: [],
    };
    this.cardsSignal.update((c) => [...c, card]);
    return card;
  }

  updateCard(cardId: string, updates: Partial<Card>): void {
    this.cardsSignal.update((cards) =>
      cards.map((c) => (c.id === cardId ? { ...c, ...updates } : c))
    );
  }

  updateList(listId: string, updates: Partial<List>): void {
    this.listsSignal.update((lists) =>
      lists.map((l) => (l.id === listId ? { ...l, ...updates } : l))
    );
  }

  updateBoard(boardId: string, updates: Partial<Board>): void {
    this.boardsSignal.update((boards) =>
      boards.map((b) => (b.id === boardId ? { ...b, ...updates } : b))
    );
  }

  deleteCard(cardId: string): void {
    this.cardsSignal.update((cards) => cards.filter((c) => c.id !== cardId));
  }

  deleteList(listId: string): void {
    this.listsSignal.update((lists) => lists.filter((l) => l.id !== listId));
    this.cardsSignal.update((cards) => cards.filter((c) => c.listId !== listId));
  }

  moveCard(cardId: string, targetListId: string, newOrder: number): void {
    this.cardsSignal.update((cards) => {
      const card = cards.find((c) => c.id === cardId);
      if (!card) return cards;
      const sourceListId = card.listId;
      const targetCards = cards
        .filter((c) => c.listId === targetListId && c.id !== cardId)
        .sort((a, b) => a.order - b.order);
      const sourceCards = cards
        .filter((c) => c.listId === sourceListId && c.id !== cardId)
        .sort((a, b) => a.order - b.order);
      const newTargetCards = [...targetCards];
      newTargetCards.splice(newOrder, 0, { ...card, listId: targetListId, order: newOrder });
      const resultTarget = newTargetCards.map((c, i) => ({ ...c, order: i }));
      const resultSource = sourceCards.map((c, i) => ({ ...c, order: i }));
      const others = cards.filter(
        (c) => c.listId !== sourceListId && c.listId !== targetListId
      );
      return [...others, ...resultSource, ...resultTarget];
    });
  }

  reorderCardsInList(listId: string, cardIds: string[]): void {
    this.cardsSignal.update((cards) => {
      const byList = cards.filter((c) => c.listId === listId);
      const others = cards.filter((c) => c.listId !== listId);
      const reordered = cardIds.map((id, order) => {
        const card = byList.find((c) => c.id === id);
        return card ? { ...card, order } : null;
      }).filter(Boolean) as Card[];
      const rest = byList.filter((c) => !cardIds.includes(c.id));
      let maxOrder = reordered.length;
      rest.forEach((c) => {
        reordered.push({ ...c, order: maxOrder++ });
      });
      return [...others, ...reordered];
    });
  }

  reorderLists(boardId: string, listIds: string[]): void {
    this.listsSignal.update((lists) => {
      const boardLists = lists.filter((l) => l.boardId === boardId);
      const others = lists.filter((l) => l.boardId !== boardId);
      const reordered = listIds.map((id, order) => {
        const list = boardLists.find((l) => l.id === id);
        return list ? { ...list, order } : null;
      }).filter(Boolean) as List[];
      return [...others, ...reordered];
    });
  }

  getCard(cardId: string): Card | undefined {
    return this.cardsSignal().find((c) => c.id === cardId);
  }

  openCardDetail(cardId: string): void {
    this.selectedCardIdSignal.set(cardId);
  }

  closeCardDetail(): void {
    this.selectedCardIdSignal.set(null);
  }

  addComment(cardId: string, author: string, text: string): void {
    const card = this.getCard(cardId);
    if (!card) return;
    const comments = card.comments ?? [];
    const comment: Comment = {
      id: `comment-${Date.now()}`,
      cardId,
      author,
      text,
      createdAt: new Date().toISOString(),
    };
    this.updateCard(cardId, { comments: [...comments, comment] });
  }

  deleteComment(cardId: string, commentId: string): void {
    const card = this.getCard(cardId);
    if (!card?.comments) return;
    this.updateCard(cardId, { comments: card.comments.filter((c) => c.id !== commentId) });
  }

  addChecklistItem(cardId: string, title: string): void {
    const card = this.getCard(cardId);
    const items = card?.checklist ?? [];
    const order = items.length;
    const item: ChecklistItem = {
      id: `ch-${Date.now()}`,
      cardId,
      title,
      completed: false,
      order,
    };
    this.updateCard(cardId, { checklist: [...items, item] });
  }

  toggleChecklistItem(cardId: string, itemId: string): void {
    const card = this.getCard(cardId);
    if (!card?.checklist) return;
    const updated = card.checklist.map((i) =>
      i.id === itemId ? { ...i, completed: !i.completed } : i
    );
    this.updateCard(cardId, { checklist: updated });
  }

  deleteChecklistItem(cardId: string, itemId: string): void {
    const card = this.getCard(cardId);
    if (!card?.checklist) return;
    const updated = card.checklist.filter((i) => i.id !== itemId).map((i, idx) => ({ ...i, order: idx }));
    this.updateCard(cardId, { checklist: updated });
  }

  addLabelToCard(cardId: string, label: Label): void {
    const card = this.getCard(cardId);
    const labels = card?.labels ?? [];
    if (labels.some((l) => l.id === label.id)) return;
    this.updateCard(cardId, { labels: [...labels, label] });
  }

  removeLabelFromCard(cardId: string, labelId: string): void {
    const card = this.getCard(cardId);
    if (!card?.labels) return;
    this.updateCard(cardId, { labels: card.labels.filter((l) => l.id !== labelId) });
  }

  addAttachment(cardId: string, name: string, url: string, type: Attachment['type']): void {
    const card = this.getCard(cardId);
    const attachments = card?.attachments ?? [];
    const att: Attachment = {
      id: `att-${Date.now()}`,
      cardId,
      name,
      url,
      type,
    };
    this.updateCard(cardId, { attachments: [...attachments, att] });
  }

  removeAttachment(cardId: string, attachmentId: string): void {
    const card = this.getCard(cardId);
    if (!card?.attachments) return;
    this.updateCard(cardId, { attachments: card.attachments.filter((a) => a.id !== attachmentId) });
  }

  addAssignee(cardId: string, name: string): void {
    const card = this.getCard(cardId);
    const assignees = card?.assignees ?? [];
    if (assignees.includes(name)) return;
    this.updateCard(cardId, { assignees: [...assignees, name] });
  }

  removeAssignee(cardId: string, name: string): void {
    const card = this.getCard(cardId);
    if (!card?.assignees) return;
    this.updateCard(cardId, { assignees: card.assignees.filter((a) => a !== name) });
  }
}
