import { Request, Response } from 'express';
import { inject, injectable } from 'inversify';
import { CreateMessageUseCase } from '../../application/use-cases/create-message.use-case';
import { TYPES } from '../../types';

@injectable()
export class MessageController {
  constructor(
    @inject(TYPES.CreateMessageUseCase) private createMessageUseCase: CreateMessageUseCase
  ) {}

  async create(req: Request, res: Response): Promise<Response> {
    try {
      const { text, authorId, channelId } = req.body;
      const message = await this.createMessageUseCase.execute({ text, authorId, channelId });
      return res.status(201).json(message);
    } catch (error) {
      return res.status(500).json({ error: 'Internal server error' });
    }
  }
}