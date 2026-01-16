import { Router, Response } from 'express';
import { User } from './model/User';
import { authMiddleware, adminMiddleware, AuthRequest } from './middleware/auth';

const router = Router();

// GET /users/:id/favorites
router.get('/:id/favorites', async (req, res) => {
    try {
        const user = await User.findById(req.params.id);
        if (!user) {
            return res.status(404).json({ message: 'User not found' });
        }
        res.json({ favoriteArtists: user.favorite_artists || [] });
    } catch (error) {
        res.status(500).json({ message: 'Server error', error });
    }
});

// GET /users/:id/infos
router.get('/:id/infos', authMiddleware, async (req: AuthRequest, res) => {
    try {
        const user = await User.findById(req.params.id).select('-password');
        if (!user) {
            return res.status(404).json({ message: 'User not found' });
        }

        res.json({
            id: user._id,
            username: user.username,
            email: user.email,
            role: user.role,
            stats: user.stats,
            favorite_artists: user.favorite_artists,
            created_at: user.created_at
        });
    } catch (error) {
        res.status(500).json({ message: 'Server error', error });
    }
});

// PUT /users/:id/infos
router.put('/:id/infos', authMiddleware, async (req: AuthRequest, res) => {
    try {
        if (req.user?.id !== req.params.id) {
            return res.status(403).json({ message: 'Unauthorized' });
        }

        const { username, favorite_artists, spotify_username } = req.body;
        let newFavoriteArtists = favorite_artists;

        // Si un username Spotify est fourni, on déclenche l'ingestion
        if (spotify_username) {
            try {
                const ingestionUrl = process.env.INGESTION_URL || 'http://ingestion:8000';
                console.log(`Triggering ingestion for spotify user: ${spotify_username}`);
                
                const response = await fetch(`${ingestionUrl}/ingest/user`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ identifier: spotify_username })
                });

                if (response.ok) {
                    const data = await response.json() as { artists?: string[] };
                    if (data.artists && Array.isArray(data.artists)) {
                        console.log(`Ingestion success. Found artists: ${data.artists}`);
                        if (data.artists.length > 0) {
                            newFavoriteArtists = data.artists;
                        }
                    }
                } else {
                    console.error('Ingestion service failed', await response.text());
                }
            } catch (ingestError) {
                console.error('Error contacting ingestion service:', ingestError);
                // On ne bloque pas la mise à jour utilisateur si l'ingestion échoue
            }
        }

        const updateData: any = {
            updated_at: new Date()
        };
        if (username) updateData.username = username;
        if (newFavoriteArtists) updateData.favorite_artists = newFavoriteArtists;
        if (spotify_username) updateData.spotify_username = spotify_username;

        const user = await User.findByIdAndUpdate(
            req.params.id,
            { $set: updateData },
            { new: true }
        ).select('-password');

        if (!user) {
            return res.status(404).json({ message: 'User not found' });
        }

        res.json({
            username: user.username,
            favorite_artists: user.favorite_artists,
            spotify_username: user.spotify_username
        });
    } catch (error) {
        console.error(error);
        res.status(500).json({ message: 'Server error' });
    }
});

// PUT /users/role (admin only)
router.put('/role', authMiddleware, adminMiddleware, async (req: AuthRequest, res: Response) => {
    try {
        const { userId, role } = req.body;

        if (!['ADMIN', 'USER'].includes(role)) {
            return res.status(400).json({ message: 'Invalid role' });
        }

        const user = await User.findByIdAndUpdate(
            userId,
            { $set: { role } },
            { new: true }
        ).select('-password');

        if (!user) {
            return res.status(404).json({ message: 'User not found' });
        }

        res.json({
            message: 'Role updated',
            user: {
                id: user._id,
                username: user.username,
                role: user.role
            }
        });
    } catch (error) {
        res.status(500).json({ message: 'Server error' });
    }
});

// PUT /users/:id/stats
router.put('/:id/stats', authMiddleware, async (req: AuthRequest, res) => {
    try {
        const user = await User.findByIdAndUpdate(
            req.params.id,
            { $set: { stats: req.body } },
            { new: true }
        ).select('-password');

        if (!user) {
            return res.status(404).json({ message: 'User not found' });
        }

        res.json(user.stats);
    } catch (error) {
        res.status(500).json({ message: 'Server error', error });
    }
});

// DELETE /users/:id (admin only)
router.delete('/:id', authMiddleware, adminMiddleware, async (req: AuthRequest, res) => {
    try {
        if (req.user?.id !== req.params.id) {
            return res.status(403).json({ message: 'Unauthorized' });
        }

        const user = await User.findByIdAndDelete(req.params.id);
        if (!user) {
            return res.status(404).json({ message: 'User not found' });
        }

        res.json({ message: 'User deleted' });
    } catch (error) {
        res.status(500).json({ message: 'Server error', error });
    }
});

export default router;