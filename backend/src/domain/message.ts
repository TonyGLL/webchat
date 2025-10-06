export interface Message {
  id: string;
  text: string;
  authorId: string;
  channelId: string;
  createdAt: Date;
}