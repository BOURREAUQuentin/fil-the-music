import mongoose from 'mongoose';

interface IUser extends mongoose.Document {
    username: string;
    email: string;
    password: string;
    role: 'ADMIN' | 'USER';
    stats: {
        total_games_played: number;
        total_score_accumulated: number;
        average_score: number;
    };
    spotify_username?: string;
    favorite_artists: string[];
    created_at: Date;
    updated_at: Date;
}

const userSchema = new mongoose.Schema<IUser>({
    username: { type: String, required: true, unique: true },
    email: { type: String, required: true, unique: true },
    password: { type: String, required: true },
    role: { type: String, enum: ['ADMIN', 'USER'], default: 'USER' },
    stats: {
        total_games_played: { type: Number, default: 0 },
        total_score_accumulated: { type: Number, default: 0 },
        average_score: { type: Number, default: 0 }
    },
    spotify_username: { type: String },
    favorite_artists: [{ type: String }],
    created_at: { type: Date, default: Date.now },
    updated_at: { type: Date, default: Date.now }
});

export const User = mongoose.model<IUser>('User', userSchema);