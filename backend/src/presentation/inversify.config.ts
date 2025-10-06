import 'reflect-metadata';
import { Container } from 'inversify';
import { Pool } from 'pg';

import { IMessageRepository } from '../application/repositories/message.repository';
import { CreateMessageUseCase } from '../application/use-cases/create-message.use-case';
import dbPool from '../infrastructure/database/db';
import { PgMessageRepository } from '../infrastructure/database/repositories/pg-message.repository';
import { TYPES } from '../types';
import { MessageController } from './controllers/message.controller';

const container = new Container();

// Database
container.bind<Pool>(TYPES.DbPool).toConstantValue(dbPool);

// Repositories
container.bind<IMessageRepository>(TYPES.IMessageRepository).to(PgMessageRepository);

// Use Cases
container.bind<CreateMessageUseCase>(TYPES.CreateMessageUseCase).to(CreateMessageUseCase);

// Controllers
container.bind<MessageController>(MessageController).to(MessageController);

export { container };