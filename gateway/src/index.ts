import express from 'express';
import { createProxyMiddleware } from 'http-proxy-middleware';
import cors from 'cors';
import * as grpc from '@grpc/grpc-js';
import * as protoLoader from '@grpc/proto-loader';
import path from 'path';
import dotenv from 'dotenv';
import helmet from 'helmet';
import rateLimit from 'express-rate-limit';

dotenv.config();

const app = express();

// Security Middleware
app.use(helmet());
app.use(cors());

// Rate Limiting
const limiter = rateLimit({
    windowMs: 15 * 60 * 1000, // 15 minutes
    max: 100, // Limit each IP to 100 requests per windowMs
    standardHeaders: true,
    legacyHeaders: false,
});
app.use(limiter);

app.use(express.json());

const SERVICES = {
    USER: process.env.USER_SERVICE_URL || 'http://user_service:3001',
    CATALOGUE: process.env.CATALOGUE_SERVICE_URL || 'http://catalog:3200',
    INGESTION: process.env.INGESTION_SERVICE_URL || 'http://ingestion_service:5000',
    GAME_GRPC: process.env.GAME_SERVICE_URL || 'game_service:50051'
};

// Configure USER service
app.use('/api/users', createProxyMiddleware({
    target: SERVICES.USER,
    changeOrigin: true,
    pathRewrite: { '^/api/users': '' },
}));

// Configure GAME service
const PROTO_PATH = path.join(__dirname, '../proto/game.proto');
const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
    keepCase: true,
    longs: String,
    enums: String,
    defaults: true,
    oneofs: true
});
const grpcObject = grpc.loadPackageDefinition(packageDefinition) as any;
const gameProto = grpcObject.user.v1;

const gameClient = new gameProto.GameService(
    SERVICES.GAME_GRPC,
    grpc.credentials.createInsecure()
);

// Helper for gRPC async/await
const grpcCall = (method: Function, payload: any): Promise<any> => {
    return new Promise((resolve, reject) => {
        method.call(gameClient, payload, (err: any, response: any) => {
            if (err) reject(err);
            else resolve(response);
        });
    });
};

// Game Routes
const gameRouter = express.Router();

type HttpMethod = 'get' | 'post';

interface GrpcRouteDefinition {
    path: string;
    method: HttpMethod;
    grpcAction: string; 
}

const GRPC_ROUTES: GrpcRouteDefinition[] = [
    { path: '/quizzes', method: 'get', grpcAction: 'GetQuizzes' },
    { path: '/quizzes', method: 'post', grpcAction: 'CreateQuiz' },
    { path: '/sessions', method: 'get', grpcAction: 'GetSessions' },
    { path: '/join', method: 'post', grpcAction: 'JoinQuiz' },
    { path: '/quit', method: 'post', grpcAction: 'QuitQuiz' },
    { path: '/active', method: 'get', grpcAction: 'GetActiveQuiz' },
    { path: '/answer', method: 'post', grpcAction: 'AnswerQuestions' },
];

// Link routes to grpc actions
GRPC_ROUTES.forEach(route => {
    gameRouter[route.method](route.path, async (req, res) => {
        try {
            const grpcMethod = gameClient[route.grpcAction];
            if (typeof grpcMethod !== 'function') {
                return res.status(500).json({ error: "Internal Server Error: Method not found" });
            }

            const response = await grpcCall(grpcMethod, req.body);
            res.json(response);
        } catch (err) {
            console.error(`Error in ${route.path}:`, err);
            res.status(500).json(err);
        }
    });
});

app.use('/api/game', gameRouter);

const PORT = process.env.PORT || 8080;
app.listen(PORT, () => {
    console.log(`Gateway running on http://localhost:${PORT}`);
});