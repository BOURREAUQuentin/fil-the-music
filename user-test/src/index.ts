import express from "express";
import mongoose from "mongoose";

const app = express();
const port = 3201;

const mongoUri = process.env.MONGO_URI || "mongodb://root:password@mongo:27017/user_db?authSource=admin";

mongoose.connect(mongoUri)
    .then(() => console.log("Connected to MongoDB"))
    .catch(err => console.error("MongoDB connection error:", err));

app.get("/users", (req, res) => {
    res.json([{ id: 1, name: "John Doe" }]);
});

app.listen(port, () => {
    console.log(`User service running on port ${port}`);
});
