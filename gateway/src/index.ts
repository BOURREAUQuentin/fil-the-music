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
    CATALOGUE: process.env.CATALOGUE_SERVICE_URL || 'http://catalogue_service:4000',
    INGESTION: process.env.INGESTION_SERVICE_URL || 'http://ingestion_service:5000',
    GAME_GRPC: process.env.GAME_SERVICE_URL || 'game_service:50051'
};

// Configure USER service
app.use('/api/users', createProxyMiddleware({
    target: SERVICES.USER,
    changeOrigin: true,
    pathRewrite: { '^/api/users': '' },
}));

// Configure CATALOGUE service
app.use('/graphql', createProxyMiddleware({
    target: SERVICES.CATALOGUE,
    changeOrigin: true,
}));

// Configure GAME service (gRPC)
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

// GET /api/game/quizzes
gameRouter.get('/quizzes', async (req, res) => {
    try {
        const response = await grpcCall(gameClient.GetQuizzes, {});
        res.json(response);
    } catch (err) {
        res.status(500).json(err);
    }
});

// POST /api/game/quizzes (Create Quiz)
gameRouter.post('/quizzes', async (req, res) => {
    try {
        const response = await grpcCall(gameClient.CreateQuiz, req.body);
        res.json(response);
    } catch (err) {
        res.status(500).json(err);
    }
});

// GET /api/game/sessions
gameRouter.get('/sessions', async (req, res) => {
    try {
        const response = await grpcCall(gameClient.GetSessions, {});
        res.json(response);
    } catch (err) {
        res.status(500).json(err);
    }
});

// POST /api/game/join
gameRouter.post('/join', async (req, res) => {
    try {
        // Expecting { user_id, quiz_id } in body
        const response = await grpcCall(gameClient.JoinQuiz, req.body);
        res.json(response);
    } catch (err) {
        res.status(500).json(err);
    }
});

// POST /api/game/quit
gameRouter.post('/quit', async (req, res) => {
    try {
        // Expecting { user_id, quiz_id } in body
        const response = await grpcCall(gameClient.QuitQuiz, req.body);
        res.json(response);
    } catch (err) {
        res.status(500).json(err);
    }
});

// GET /api/game/active
gameRouter.get('/active', async (req, res) => {
    try {
        // Expecting user_id in query params
        const { user_id } = req.query;
        if (!user_id) {
             return res.status(400).json({ error: "Missing user_id query parameter" });
        }
        const response = await grpcCall(gameClient.GetActiveQuiz, { user_id });
        res.json(response);
    } catch (err) {
        res.status(500).json(err);
    }
});

// POST /api/game/answer
gameRouter.post('/answer', async (req, res) => {
    try {
        // Expecting { user_id, answers: [] } in body
        const response = await grpcCall(gameClient.AnswerQuestions, req.body);
        res.json(response);
    } catch (err) {
        res.status(500).json(err);
    }
});

app.use('/api/game', gameRouter);

const PORT = process.env.PORT || 8080;
app.listen(PORT, () => {
    console.log(`Gateway running on http://localhost:${PORT}`);
});