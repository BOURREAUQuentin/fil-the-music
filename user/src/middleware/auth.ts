import { Request, Response, NextFunction } from 'express';
import jwt from 'jsonwebtoken';

// JWT authentication middleware
export interface AuthRequest extends Request {
    user?: { id: string; role: string };
}

// Verify the secret at loading of the file
const JWT_SECRET = process.env.JWT_SECRET;
if (!JWT_SECRET) {
    throw new Error('JWT_SECRET is not defined in the .env file');
}

export const authMiddleware = async (
    req: AuthRequest,
    res: Response,
    next: NextFunction
) => {
    try {
        const token = req.headers.authorization?.split(' ')[1];

        if (!token) {
            return res.status(401).json({ error: 'Missing token' });
        }

        const decoded = jwt.verify(
            token,
            JWT_SECRET
        ) as { id: string; role: string };

        req.user = decoded;
        next();
    } catch (error) {
        return res.status(401).json({ error: 'Invalid token' });
    }
};

// Middleware to check ADMIN role
export const adminMiddleware = (
    req: AuthRequest,
    res: Response,
    next: NextFunction
) => {
    if (req.user?.role !== 'ADMIN') {
        return res.status(403).json({
            error: 'Access denied: ADMIN role required'
        });
    }
    next();
};