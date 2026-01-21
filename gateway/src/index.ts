import cors from 'cors';
import * as grpc from '@grpc/grpc-js';
import * as protoLoader from '@grpc/proto-loader';
import path from 'path';
import dotenv from 'dotenv';
import helmet from 'helmet';
import rateLimit from 'express-rate-limit';
import express from 'express';
import { createProxyMiddleware } from 'http-proxy-middleware';

dotenv.config();

const app = express();

// Security Middleware
app.use(helmet());
app.use(cors());

// Rate Limiting
const limiter = rateLimit({
  windowMs: 15 * 60 * 1000,
  max: 100,
  standardHeaders: true,
  legacyHeaders: false,
});
app.use(limiter);

const SERVICES = {
  USER: process.env.USER_SERVICE_URL || 'http://user:3201',
  GAME_GRPC: process.env.GAME_SERVICE_URL || 'game_service:50051'
};

console.log('🔧 Services configuration:', SERVICES);

// Configure USER service
app.use('/api/users', createProxyMiddleware({
  target: SERVICES.USER,
  changeOrigin: true,
  pathRewrite: {
    '^/api/users': ''
  },
  on: {
    proxyReq: (proxyReq, req, res) => {
      console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
      console.log(`[PROXY DEBUG - USER]`);
      console.log(`  Method: ${req.method}`);
      console.log(`  req.originalUrl: ${(req as any).originalUrl}`);
      console.log(`  Target: ${SERVICES.USER}${proxyReq.path}`);
      console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
    },
    proxyRes: (proxyRes, req, res) => {
      console.log(`[PROXY RESPONSE - USER] Status: ${proxyRes.statusCode}`);
    },
    error: (err, req, res) => {
      console.error('[PROXY ERROR]', err.message);
      if (res && 'writeHead' in res) {
        res.writeHead(502, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({
          error: 'Gateway error',
          message: err.message
        }));
      }
    }
  }
}));

app.use(express.json());

// Health check
app.get('/health', (req, res) => {
  res.json({ 
    status: 'ok', 
    services: SERVICES 
  });
});

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
  { path: '/start/random', method: 'post', grpcAction: 'StartRandomQuiz' },
  { path: '/start/genre', method: 'post', grpcAction: 'StartGenreQuiz' },
  { path: '/start/foryou', method: 'post', grpcAction: 'StartForYouQuiz' },
];

// Link routes to grpc actions
GRPC_ROUTES.forEach(route => {
  gameRouter[route.method](route.path, async (req, res) => {
    console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
    console.log(`[GAME GRPC DEBUG]`);
    console.log(`  HTTP Request: ${req.method.toUpperCase()} ${req.originalUrl}`);
    console.log(`  Mapped gRPC Action: ${route.grpcAction}`);
    console.log(`  Payload sending to gRPC:`, JSON.stringify(req.body, null, 2));
    console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');

    try {
      const grpcMethod = gameClient[route.grpcAction];
      if (typeof grpcMethod !== 'function') {
        console.error(`[GAME ERROR] Method ${route.grpcAction} not found on gRPC client`);
        return res.status(500).json({ error: "Internal Server Error: Method not found" });
      }

      const response = await grpcCall(grpcMethod, req.body);
      
      console.log(`[GAME GRPC RESPONSE] Success`);
      
      res.json(response);
    } catch (err: any) {
      console.error('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
      console.error(`[GAME GRPC ERROR]`);
      console.error(`  Action: ${route.grpcAction}`);
      console.error(`  Message:`, err.details || err.message || err);
      console.error('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
      
      res.status(500).json(err);
    }
  });
});

app.use('/api/game', gameRouter);

const PORT = process.env.PORT || 8080;
app.listen(PORT, () => {
  console.log(`🚀 Gateway running on port ${PORT}`);
  console.log(`📡 Proxying /api/users -> ${SERVICES.USER}`);
  console.log(`📡 Proxying /api/game  -> ${SERVICES.GAME_GRPC} (gRPC)`);
});