import { Router, Response } from 'express';
import { User } from './model/User';
import { authMiddleware, AuthRequest } from './middleware/auth';

const router = Router();

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

        const { username, favorite_artists } = req.body;

        const user = await User.findByIdAndUpdate(
            req.params.id,
            {
                $set: {
                    ...(username && { username }),
                    ...(favorite_artists && { favorite_artists })
                },
                updated_at: new Date()
            },
            { new: true }
        ).select('-password');

        if (!user) {
            return res.status(404).json({ message: 'User not found' });
        }

        res.json({
            username: user.username,
            favorite_artists: user.favorite_artists
        });
    } catch (error) {
        res.status(500).json({ message: 'Server error' });
    }
});

// PUT /users/role (admin only)
router.put('/role', authMiddleware, async (req: AuthRequest, res: Response) => {
    try {
        const currentUser = await User.findById(req.user?.id);
        if (!currentUser || currentUser.role !== 'ADMIN') {
            return res.status(403).json({ message: 'Admin access required' });
        }

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

// DELETE /users/:id
router.delete('/:id', authMiddleware, async (req: AuthRequest, res) => {
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