"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const express_1 = __importDefault(require("express"));
const mongoose_1 = __importDefault(require("mongoose"));
const app = (0, express_1.default)();
const port = 3201;
const mongoUri = process.env.MONGO_URI || "mongodb://root:password@mongo:27017/user_db?authSource=admin";
mongoose_1.default.connect(mongoUri)
    .then(() => console.log("Connected to MongoDB"))
    .catch(err => console.error("MongoDB connection error:", err));
app.get("/users", (req, res) => {
    res.json([{ id: 1, name: "John Doe" }]);
});
app.listen(port, () => {
    console.log(`User service running on port ${port}`);
});
