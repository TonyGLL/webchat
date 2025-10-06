import { Router } from 'express';
import { container } from '../inversify.config';
import { MessageController } from '../controllers/message.controller';

const router = Router();
const messageController = container.resolve(MessageController);

router.post('/', (req, res) => messageController.create(req, res));

export default router;
