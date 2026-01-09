import express from "express";
import mongoose from "mongoose";
import dotenv from "dotenv";
import userRouter from "./index";
import authRouter from "./auth";

dotenv.config();

const app = express();
const port = process.env.PORT || 3201;

app.use(express.json());

// Connexion à MongoDB
mongoose.connect(process.env.MONGO_URI || "")
    .then(() => console.log("Connected to MongoDB"))
    .catch(err => console.error("MongoDB connection error:", err));

// Routes
app.use("/users", userRouter);
app.use("/", authRouter);

// Démarrage serveur
app.listen(port, () => {
    console.log(`User service running on port ${port}`);
});