FROM node:20-alpine

WORKDIR /app

# Copier uniquement package.json et package-lock.json pour installer les deps
COPY package*.json ./
RUN npm install

# Copier le reste du projet
COPY . .

# Compiler TypeScript
RUN npx tsc

EXPOSE 3201

# Lancer le service
CMD ["node", "dist/index.js"]
