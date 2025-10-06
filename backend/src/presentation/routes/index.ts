import { Router } from 'express';
import messageRouter from './message.routes';

const router = Router();

router.use('/messages', messageRouter);

export default router;
