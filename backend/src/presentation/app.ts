import 'reflect-metadata';
import express, { Application } from 'express';
import apiRouter from './routes';

const app: Application = express();
const port = process.env.PORT || 3000;

app.use(express.json());

app.use('/api', apiRouter);

export { app, port };