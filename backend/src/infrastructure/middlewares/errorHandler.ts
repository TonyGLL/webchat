import { ErrorRequestHandler, NextFunction, Request, Response } from "express";
import { HttpError } from "../../domain/errors/http";
import { logger } from "../../config/winston";

export const errorHandler: ErrorRequestHandler = (err: Error, _: Request, res: Response, __: NextFunction): any => {
    if (err instanceof HttpError) {
        return res.status(err.statusCode).json({ error: err.message });
    }

    // Log unexpected errors for debugging in production
    logger.error(err);

    return res.status(500).json({ error: 'Server error' });
}