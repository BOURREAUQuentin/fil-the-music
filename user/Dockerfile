FROM node:20-alpine

WORKDIR /app

# Copier package.json et package-lock.json pour installer deps
COPY package*.json ./
RUN npm install

# Copier tout le reste
COPY . .

# Compiler TypeScript
RUN npx tsc

EXPOSE 3201

# Lancer le microservice
CMD ["node", "dist/server.js"]
