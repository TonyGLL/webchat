import 'reflect-metadata';
import { env } from './config/env';
import app from './presentation/app';

const PORT = env.PORT;

const server = app.listen(PORT, () => {
  console.log('Server running on port:', PORT);
});

process.on('SIGINT', () => {
  console.log('\n Shutting down server...');
  server.close(() => {
    console.log('Server off.');
    process.exit(0);
  });
});