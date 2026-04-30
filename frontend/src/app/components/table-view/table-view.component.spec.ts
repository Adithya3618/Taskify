import { TestBed } from '@angular/core/testing';
import { computed, signal } from '@angular/core';
import { TableViewComponent } from './table-view.component';
import { BoardService } from '../../services/board.service';
import { List } from '../../models/list.model';
import { Card } from '../../models/card.model';

describe('TableViewComponent', () => {
  let listsSig: ReturnType<typeof signal<List[]>>;
  let cardsSig: ReturnType<typeof signal<Card[]>>;
  let openSpy: jasmine.Spy;

  beforeEach(() => {
    listsSig = signal<List[]>([
      { id: '10', title: 'Open', boardId: '1', order: 0 },
      { id: '20', title: 'Done', boardId: '1', order: 1 },
    ]);
    cardsSig = signal<Card[]>([
      { id: 'a', title: 'Alpha', listId: '20', order: 0 },
      { id: 'b', title: 'Beta', listId: '10', order: 0 },
    ]);
    openSpy = jasmine.createSpy('openCardDetail');

    const boardStub = {
      currentLists: computed(() => listsSig()),
      currentBoardCards: computed(() => cardsSig()),
      lists: () => listsSig(),
      openCardDetail: openSpy,
    } as unknown as BoardService;

    TestBed.configureTestingModule({
      imports: [TableViewComponent],
      providers: [{ provide: BoardService, useValue: boardStub }],
    });
  });

  it('should sort table rows by list order then card order', () => {
    const fixture = TestBed.createComponent(TableViewComponent);
    const rows = fixture.componentInstance.tableRows();
    expect(rows.map((c) => c.title)).toEqual(['Beta', 'Alpha']);
  });

  it('trackByCardId should return the card id string', () => {
    const fixture = TestBed.createComponent(TableViewComponent);
    const card: Card = { id: 'xyz', title: 'T', listId: '10', order: 0 };
    expect(fixture.componentInstance.trackByCardId(0, card)).toBe('xyz');
  });

  it('openCard should delegate to BoardService.openCardDetail', () => {
    const fixture = TestBed.createComponent(TableViewComponent);
    fixture.componentInstance.openCard('card-99');
    expect(openSpy).toHaveBeenCalledWith('card-99');
  });

  it('should expose a fixed virtual row height for CDK scrolling', () => {
    const fixture = TestBed.createComponent(TableViewComponent);
    expect(fixture.componentInstance.virtualRowHeight).toBe(64);
  });
});
