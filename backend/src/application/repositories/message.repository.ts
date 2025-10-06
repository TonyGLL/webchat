import { Message } from '../../domain/message';

export interface IMessageRepository {
  create(message: Omit<Message, 'id' | 'createdAt'>): Promise<Message>;
  findById(id: string): Promise<Message | null>;
  findByChannelId(channelId: string): Promise<Message[]>;
}

export const IMessageRepository = Symbol.for('IMessageRepository');