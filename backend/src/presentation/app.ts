import express, { Request, Response, type Application } from "express";
import router from "./routes";
import helmet from "helmet";
import cors from "cors";
import morgan from "morgan";
import { errorHandler } from "../infrastructure/middlewares/errorHandler";

class App {
    public express: Application;

    constructor() {
        this.express = express();
        this.middlewares();
        this.routes();
        this.express.use(errorHandler);
    }

    private middlewares(): void {
        const corsOptions = {
            origin: 'http://localhost:3001'
        };
        this.express.use(cors(corsOptions));
        this.express.use(express.urlencoded({ extended: true }));
        this.express.use(morgan('dev'));
        this.express.use(helmet({
            contentSecurityPolicy: {
                directives: {
                    defaultSrc: ["'self'"],
                    scriptSrc: ["'self'", "https://js.stripe.com/v3/"],
                    styleSrc: ["'self'", "'unsafe-inline'"], // EJS/inline styles might require this
                    frameSrc: ["'self'", "https://js.stripe.com/v3/"],
                    connectSrc: ["'self'", "https://api.stripe.com"],
                    imgSrc: ["'self'", "data:"], // Allow data URIs for images
                    fontSrc: ["'self'"],
                },
            },
        }));
    }

    private routes(): void {
        this.express.use('/health', (_: Request, res: Response) => {
            res.status(200).json({
                ok: true,
                message: 'Server is healthy',
            })
        });

        this.express.use(express.json());

        this.express.use('/api/v1', router);
    }
}

export default new App().express;