import { injectable, inject } from 'inversify';
import { Pool } from 'pg';
import { IMessageRepository } from '../../../application/repositories/message.repository';
import { Message } from '../../../domain/message';
import { TYPES } from '../../../types';

@injectable()
export class PgMessageRepository implements IMessageRepository {
  constructor(@inject(TYPES.DbPool) private pool: Pool) {}

  async create(messageData: Omit<Message, 'id' | 'createdAt'>): Promise<Message> {
    const { text, authorId, channelId } = messageData;
    const result = await this.pool.query(
      'INSERT INTO messages (text, author_id, channel_id) VALUES ($1, $2, $3) RETURNING *',
      [text, authorId, channelId]
    );
    const row = result.rows[0];
    return {
      id: row.id,
      text: row.text,
      authorId: row.author_id,
      channelId: row.channel_id,
      createdAt: row.created_at,
    };
  }

  async findById(id: string): Promise<Message | null> {
    const result = await this.pool.query('SELECT * FROM messages WHERE id = $1', [id]);
    if (result.rows.length === 0) {
      return null;
    }
    const row = result.rows[0];
    return {
      id: row.id,
      text: row.text,
      authorId: row.author_id,
      channelId: row.channel_id,
      createdAt: row.created_at,
    };
  }

  async findByChannelId(channelId: string): Promise<Message[]> {
    const result = await this.pool.query('SELECT * FROM messages WHERE channel_id = $1 ORDER BY created_at ASC', [channelId]);
    return result.rows.map(row => ({
      id: row.id,
      text: row.text,
      authorId: row.author_id,
      channelId: row.channel_id,
      createdAt: row.created_at,
    }));
  }
}