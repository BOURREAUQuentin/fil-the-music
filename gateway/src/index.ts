import cors from 'cors';
import * as grpc from '@grpc/grpc-js';
import * as protoLoader from '@grpc/proto-loader';
import path from 'path';
import dotenv from 'dotenv';
import helmet from 'helmet';
import rateLimit from 'express-rate-limit';
import express, { Request, Response } from 'express';
import { createProxyMiddleware } from 'http-proxy-middleware';
import { authenticateToken, authorizeRole } from './auth'; // Import auth middlewares

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
  GAME_GRPC: process.env.GAME_SERVICE_URL || 'game:50051',
};

console.log('🔧 Services configuration:', SERVICES);

// Configure USER service proxy (for login/register)
app.use('/api/users', createProxyMiddleware({
  target: SERVICES.USER,
  changeOrigin: true,
  pathRewrite: {
    '^/api/users': '',
  },
}));

// Health check
app.get('/health', (req, res) => {
  res.json({
    status: 'ok',
    services: SERVICES,
  });
});

app.use(express.json());

// Configure GAME service gRPC Client
const PROTO_PATH = path.join(__dirname, '../proto/game.proto');
const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
  keepCase: true,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true,
});

const grpcObject = grpc.loadPackageDefinition(packageDefinition) as any;
const gameProto = grpcObject.user.v1; // Adjusted to match the package name in proto
const gameClient = new gameProto.GameService(
  SERVICES.GAME_GRPC,
  grpc.credentials.createInsecure()
);

// Helper for gRPC async/await with metadata
const grpcCall = (method: Function, payload: any, metadata: grpc.Metadata): Promise<any> => {
  return new Promise((resolve, reject) => {
    method.call(gameClient, payload, metadata, (err: grpc.ServiceError, response: any) => {
      if (err) reject(err);
      else resolve(response);
    });
  });
};

// Game Routes
const gameRouter = express.Router();

type HttpMethod = 'get' | 'post';
type AuthLevel = 'NONE' | 'USER' | 'ADMIN';

interface GrpcRouteDefinition {
  path: string;
  method: HttpMethod;
  grpcAction: string;
  authLevel: AuthLevel;
}

// Define which routes are protected and by which role
const GRPC_ROUTES: GrpcRouteDefinition[] = [
  // Admin routes
  { path: '/create/simple', method: 'post', grpcAction: 'CreateQuiz', authLevel: 'ADMIN' },
  { path: '/sessions', method: 'get', grpcAction: 'GetSessions', authLevel: 'ADMIN' },

  // User routes
  { path: '/start/genre', method: 'post', grpcAction: 'StartGenreQuiz', authLevel: 'USER' },
  { path: '/start/foryou', method: 'post', grpcAction: 'StartForYouQuiz', authLevel: 'USER' },
  { path: '/start/random', method: 'post', grpcAction: 'StartRandomQuiz', authLevel: 'USER' },

  { path: '/start/:quizId', method: 'post', grpcAction: 'JoinQuiz', authLevel: 'USER' },
  { path: '/answer', method: 'post', grpcAction: 'AnswerQuestions', authLevel: 'USER' }, // Assuming SubmitAnswers is the gRPC action
  { path: '/quizzes', method: 'get', grpcAction: 'GetQuizzes', authLevel: 'USER' },
  { path: '/active', method: 'get', grpcAction: 'GetActiveQuiz', authLevel: 'USER' },
  { path: '/quit', method: 'post', grpcAction: 'QuitQuiz', authLevel: 'USER' },
];

// Link routes to gRPC actions with authentication
GRPC_ROUTES.forEach(route => {
  const middlewares = [];

  // Add authentication and authorization middlewares if the route is not public
  if (route.authLevel !== 'NONE') {
    middlewares.push(authenticateToken);
    middlewares.push(authorizeRole(route.authLevel));
  }

  gameRouter[route.method](route.path, ...middlewares, async (req: Request, res: Response) => {
    console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
    console.log(`[HTTP->GRPC] START: Handling ${req.method} ${req.originalUrl} -> ${route.grpcAction}`);

    try {
      const grpcMethod = gameClient[route.grpcAction];
      if (typeof grpcMethod !== 'function') {
        console.error(`[HTTP->GRPC ERROR] Method ${route.grpcAction} not found on gRPC client`);
        return res.status(500).json({ error: 'Internal Server Error: Method not found' });
      }

      // Prepare payload with correct field casing
      const payload: { [key: string]: any } = { ...req.body };
      if (req.params.quizId) {
        payload.quiz_id = req.params.quizId;
      }
      
      const metadata = new grpc.Metadata();
      
      // If user is authenticated, pass their info in metadata
      if (req.user) {
        metadata.add('x-user-id', req.user.id);
        metadata.add('x-user-role', req.user.role);
      }

      console.log(`[HTTP->GRPC] Preparing to call gRPC method: ${route.grpcAction}`);
      console.log('[HTTP->GRPC] Payload:', JSON.stringify(payload, null, 2));
      console.log('[HTTP->GRPC] Metadata:', JSON.stringify(metadata.getMap(), null, 2));

      const response = await grpcCall(grpcMethod, payload, metadata);
      
      console.log(`[HTTP->GRPC] SUCCESS: gRPC call for ${route.grpcAction} completed.`);
      res.json(response);
    } catch (err: any) {
      console.error(`[HTTP->GRPC] FATAL ERROR during gRPC call for action: ${route.grpcAction}`);
      console.error('[HTTP->GRPC] Full Error Object:', JSON.stringify(err, null, 2));
      
      const statusCode = err.code === grpc.status.UNAUTHENTICATED ? 401 : err.code === grpc.status.PERMISSION_DENIED ? 403 : 502;
      res.status(statusCode).json({
        error: 'gRPC call failed',
        grpc_code: err.code,
        grpc_details: err.details,
        message: err.message,
      });
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