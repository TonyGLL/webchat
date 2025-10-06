import 'reflect-metadata';
import express, { Application } from 'express';
import { container } from './inversify.config';
import { MessageController } from './controllers/message.controller';

const app: Application = express();
const port = process.env.PORT || 3000;

app.use(express.json());

const messageController = container.resolve(MessageController);

app.post('/messages', (req, res) => messageController.create(req, res));

export { app, port };