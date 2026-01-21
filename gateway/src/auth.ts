import { Request, Response, NextFunction } from 'express';
import jwt from 'jsonwebtoken';

// Define the structure of our token payload
interface UserPayload {
    id: string;
    role: 'USER' | 'ADMIN';
}

// Extend the Express Request type to include our user payload
declare global {
    namespace Express {
        interface Request {
            user?: UserPayload;
        }
    }
}

const JWT_SECRET = process.env.JWT_SECRET;
if (!JWT_SECRET) {
    console.error('[AUTH FATAL] JWT_SECRET is not defined.');
    throw new Error('FATAL ERROR: JWT_SECRET is not defined.');
}

/**
 * Middleware to authenticate a JWT token.
 * If valid, attaches the user payload to the request object.
 */
export const authenticateToken = (req: Request, res: Response, next: NextFunction) => {
    console.log('[AUTH] authenticateToken middleware triggered.');
    const authHeader = req.headers['authorization'];
    const token = authHeader && authHeader.split(' ')[1]; // Bearer TOKEN

    if (!token) {
        console.log('[AUTH-ERROR] No token provided.');
        return res.status(401).json({ error: 'Access denied, no token provided.' });
    }

    try {
        console.log('[AUTH] Verifying token...');
        const payload = jwt.verify(token, JWT_SECRET) as UserPayload;
        req.user = payload;
        console.log(`[AUTH-SUCCESS] Token verified. User ID: ${payload.id}, Role: ${payload.role}`);
        next();
    } catch (err: unknown) { // Explicitly type as unknown
        let errorMessage = 'Token verification failed: Unknown error';
        if (err instanceof Error) {
            errorMessage = `Token verification failed: ${err.message}`;
        }
        console.error('[AUTH-ERROR]', errorMessage);
        return res.status(403).json({ error: 'Invalid or expired token.' });
    }
};

/**
 * Middleware factory to authorize based on role.
 * Ensures that a user has the required role to access a route.
 * Should be used AFTER authenticateToken.
 * @param requiredRole The role required to access the route.
 */
export const authorizeRole = (requiredRole: 'USER' | 'ADMIN') => {
    return (req: Request, res: Response, next: NextFunction) => {
        console.log(`[AUTH] authorizeRole middleware triggered for role: '${requiredRole}'.`);

        if (!req.user) {
            console.error('[AUTH-ERROR] authorizeRole called but no user on request. Ensure authenticateToken runs first.');
            return res.status(401).json({ error: 'Authentication required.' });
        }

        const userRole = req.user.role;
        console.log(`[AUTH] User role: '${userRole}', Required role: '${requiredRole}'.`);

        // Admins can access everything
        if (userRole === 'ADMIN') {
            console.log('[AUTH-SUCCESS] Admin authorized.');
            return next();
        }

        // Users can only access 'user' routes
        if (requiredRole === 'USER' && userRole === 'USER') {
            console.log('[AUTH-SUCCESS] User authorized.');
            return next();
        }

        console.log('[AUTH-ERROR] Insufficient permissions.');
        return res.status(403).json({ error: 'Forbidden: Insufficient permissions.' });
    };
};
