import { NextFunction, Request, Response } from 'express';
import { inject, injectable } from 'inversify';
import { TYPES } from '../../types';
import { CreateMessageUseCase } from '../../application/use-cases/messages/create-message.use-case';

@injectable()
export class MessageController {
  constructor(
    @inject(TYPES.CreateMessageUseCase) private createMessageUseCase: CreateMessageUseCase
  ) { }

  public create = async (req: Request, res: Response, next: NextFunction): Promise<Response> => {
    try {
      const { text, authorId, channelId } = req.body;
      const message = await this.createMessageUseCase.execute({ text, authorId, channelId });
      return res.status(201).json(message);
    } catch (error) {
      return res.status(500).json({ error: 'Internal server error' });
    }
  }
}