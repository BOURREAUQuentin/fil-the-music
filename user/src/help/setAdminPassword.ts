import bcrypt from "bcryptjs";
import mongoose from "mongoose";
import { User } from "~/model/User";

async function run() {
    await mongoose.connect(process.env.MONGO_URI || "");

    const hash = await bcrypt.hash("admin123", 10);

    await User.updateOne(
        { email: "admin@quizmusic.com" },
        { $set: { password: hash } }
    );

    console.log("✅ Password updated : admin123");
    process.exit(0);
}

run();

/////////// for changing the admin password to "admin123" ///////////
// run 'docker-compose up -d'
// then 'docker compose exec user sh'
// then 'npx ts-node src/help/setAdminPassword.ts'
// if you don't have ts-node installed, run 'npm install -g ts-node typescript'
