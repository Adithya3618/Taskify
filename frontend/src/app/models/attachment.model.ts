export interface Attachment {
  id: string;
  cardId: string;
  name: string;
  url: string;
  type: 'drive' | 'dropbox' | 'link' | 'file';
}
