import winston from 'winston';
import DailyRotateFile from 'winston-daily-rotate-file';
import path from 'path';
import fs from 'fs';

// Carpeta de logs (fuera de src)
const logDir = path.join(process.cwd(), 'logs');
if (!fs.existsSync(logDir)) fs.mkdirSync(logDir);

// Format común
const fileFormat = winston.format.combine(
    winston.format.timestamp(),
    winston.format.json()
);

// Error logs
const errorTransport = new DailyRotateFile({
    filename: 'error-%DATE%.log',
    dirname: logDir,
    datePattern: 'YYYY-MM-DD',
    zippedArchive: true,
    maxSize: '10m',
    maxFiles: '14d',
    level: 'error',
    format: fileFormat,
});

// API/info logs
const apiTransport = new DailyRotateFile({
    filename: 'api-%DATE%.log',
    dirname: logDir,
    datePattern: 'YYYY-MM-DD',
    zippedArchive: true,
    maxSize: '10m',
    maxFiles: '7d',
    level: 'http', // o 'info', depende cómo lo uses
    format: fileFormat,
});

// Consola en desarrollo
const consoleTransport = new winston.transports.Console({
    level: 'debug',
    format: winston.format.combine(
        winston.format.colorize(),
        winston.format.timestamp({ format: 'YYYY-MM-DD HH:mm:ss' }),
        winston.format.printf(({ timestamp, level, message, ...meta }) => {
            const extras = Object.keys(meta).length ? JSON.stringify(meta) : '';
            return `[${timestamp}] ${level}: ${message} ${extras}`;
        })
    ),
});

// Logger con niveles personalizados
export const logger = winston.createLogger({
    levels: winston.config.npm.levels, // default: error, warn, info, http, verbose, debug, silly
    level: 'debug',
    transports: [
        errorTransport,
        apiTransport,
        consoleTransport,
    ],
});