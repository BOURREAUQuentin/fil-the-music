db = db.getSiblingDB("user_db");

db.users.insertMany([
    {
        username: "admin",
        email: "admin@quizmusic.com",
        password: "$2a$10$CwTycUXWue0Thq9StjUM0uJ8eFZ7rOe0Cz9Z6XJ5x1YxE6Z8vQe4e",
        role: "ADMIN",
        stats: {
            total_games_played: 0,
            total_score_accumulated: 0,
            average_score: 0
        },
        favorite_artists: ["Daft Punk", "Radiohead"],
        created_at: new Date(),
        updated_at: new Date()
    },
    {
        username: "user1",
        email: "user1@quizmusic.com",
        password: "$2a$10$CwTycUXWue0Thq9StjUM0uJ8eFZ7rOe0Cz9Z6XJ5x1YxE6Z8vQe4e",
        role: "USER",
        stats: {
            total_games_played: 3,
            total_score_accumulated: 210,
            average_score: 70
        },
        favorite_artists: ["Muse", "Coldplay"],
        created_at: new Date(),
        updated_at: new Date()
    }
]);
