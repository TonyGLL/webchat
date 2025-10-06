import { inject, injectable } from 'inversify';
import { Message } from '../../../domain/message';
import { IMessageRepository } from '../../repositories/message.repository';

interface CreateMessageRequest {
  text: string;
  authorId: string;
  channelId: string;
}

@injectable()
export class CreateMessageUseCase {
  constructor(
    @inject(IMessageRepository) private messageRepository: IMessageRepository
  ) { }

  async execute(request: CreateMessageRequest): Promise<Message> {
    const { text, authorId, channelId } = request;

    const message = await this.messageRepository.create({
      text,
      authorId,
      channelId,
    });

    return message;
  }
}