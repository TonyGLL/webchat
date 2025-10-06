import { Router } from 'express';
import { container } from '../inversify.config';
import { MessageController } from '../controllers/message.controller';

const router = Router();
const controller = container.get<MessageController>(MessageController);

router
    .post('/', controller.create);

export default router;
