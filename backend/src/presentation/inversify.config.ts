import 'reflect-metadata';
import { Container } from 'inversify';
import { Pool } from 'pg';

import { TYPES } from '../types';
import dbPool from '../infrastructure/database/db';
import { IMessageRepository } from '../application/repositories/message.repository';
import { PgMessageRepository } from '../infrastructure/database/repositories/pg-message.repository';

import { MessageController } from './controllers/message.controller';

import { CreateMessageUseCase } from '../application/use-cases/messages/create-message.use-case';

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