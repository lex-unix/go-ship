FROM registry:4443/node:20.15.1-slim

WORKDIR /app

COPY package*.json ./

COPY . .

EXPOSE 3000

CMD ["node", "index.js"]
