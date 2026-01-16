import { Router, Response, Request } from 'express';
import bcrypt from 'bcryptjs';
import jwt from 'jsonwebtoken';
import { User } from './model/User';
import {adminMiddleware, authMiddleware, AuthRequest} from "~/middleware/auth";

const authRouter = Router();

// Verify the secret at loading of the file
const JWT_SECRET = process.env.JWT_SECRET;
if (!JWT_SECRET) {
    throw new Error('JWT_SECRET is not defined in the .env file');
}

// POST /auth/register - Registration
authRouter.post('/auth/register', async (req: Request, res: Response) => {
    try {
        const { username, email, password, role, spotify_username } = req.body;

        // Validation
        if (!username || !email || !password) {
            return res.status(400).json({ error: 'All fields are required' });
        }

        // Check if user already exists
        const existingUser = await User.findOne({ $or: [{ email }, { username }] });
        if (existingUser) {
            return res.status(409).json({ error: 'Email or username already in use' });
        }

        // Hash password
        const hashedPassword = await bcrypt.hash(password, 10);

        let favorite_artists: string[] = [];
        
        // Handle Spotify Ingestion if provided
        if (spotify_username) {
            try {
                const ingestionUrl = process.env.INGESTION_URL || 'http://ingestion:8000';
                console.log(`[Register] Triggering ingestion for spotify user: ${spotify_username}`);
                
                const response = await fetch(`${ingestionUrl}/ingest/user`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ identifier: spotify_username })
                });

                if (response.ok) {
                    const data = await response.json() as { artists?: string[] };
                    if (data.artists && Array.isArray(data.artists)) {
                        console.log(`[Register] Ingestion success. Found artists: ${data.artists}`);
                        favorite_artists = data.artists;
                    }
                } else {
                    console.error('[Register] Ingestion service failed', await response.text());
                }
            } catch (ingestError) {
                console.error('[Register] Error contacting ingestion service:', ingestError);
            }
        }

        // Create user
        const newUser = new User({
            username,
            email,
            password: hashedPassword,
            role: role || 'USER',
            stats: {
                total_games_played: 0,
                total_score_accumulated: 0,
                average_score: 0
            },
            spotify_username,
            favorite_artists: favorite_artists
        });

        await newUser.save();

        // Generate JWT token
        const token = jwt.sign(
            { id: newUser._id, role: newUser.role },
            JWT_SECRET,
            { expiresIn: '7d' }
        );

        res.status(201).json({
            message: 'User successfully created',
            user: {
                id: newUser._id,
                username: newUser.username,
                email: newUser.email,
                role: newUser.role,
                favorite_artists: newUser.favorite_artists,
                spotify_username: newUser.spotify_username
            },
            token
        });
    } catch (error) {
        console.error('Error during registration:', error);
        res.status(500).json({ error: 'Server error' });
    }
});

// POST /auth/login - Login
authRouter.post('/auth/login', async (req: Request, res: Response) => {
    try {
        const { email, password } = req.body;

        if (!email || !password) {
            return res.status(400).json({ error: 'Email and password are required' });
        }

        // Find user
        const user = await User.findOne({ email });
        if (!user) {
            return res.status(401).json({ error: 'Invalid email or password' });
        }

        // Check password
        const isPasswordValid = await bcrypt.compare(password, user.password);
        if (!isPasswordValid) {
            return res.status(401).json({ error: 'Invalid email or password' });
        }

        // Generate JWT token
        const token = jwt.sign(
            { id: user._id, role: user.role },
            JWT_SECRET,
            { expiresIn: '7d' }
        );

        res.json({
            message: 'Login successful',
            user: {
                id: user._id,
                username: user.username,
                email: user.email,
                role: user.role,
                stats: user.stats,
                favorite_artists: user.favorite_artists
            },
            token
        });
    } catch (error) {
        console.error('Error during login:', error);
        res.status(500).json({ error: 'Server error' });
    }
});

// POST /auth/admin/register - Create user with a specific role (ADMIN only)
authRouter.post(
    '/auth/admin/register',
    authMiddleware,
    adminMiddleware,
    async (req: AuthRequest, res: Response) => {
        try {
            const {username, email, password, role} = req.body;

            if (!username || !email || !password) {
                return res.status(400).json({error: 'All fields are required'});
            }

            if (!['ADMIN', 'USER'].includes(role)) {
                return res.status(400).json({error: 'Invalid role'});
            }

            const existingUser = await User.findOne({
                $or: [{email}, {username}]
            });
            if (existingUser) {
                return res
                    .status(409)
                    .json({error: 'Email or username already in use'});
            }

            const hashedPassword = await bcrypt.hash(password, 10);

            const newUser = new User({
                username,
                email,
                password: hashedPassword,
                role,
                stats: {
                    total_games_played: 0,
                    total_score_accumulated: 0,
                    average_score: 0
                },
                favorite_artists: []
            });

            await newUser.save();

            res.status(201).json({
                message: 'User created by admin',
                user: {
                    id: newUser._id,
                    username: newUser.username,
                    email: newUser.email,
                    role: newUser.role
                }
            });
        } catch (error) {
            console.error('Admin user creation error:', error);
            res.status(500).json({error: 'Server error'});
        }
    }
);

export default authRouter;