FROM node:20-alpine
WORKDIR /app
COPY package.json .
COPY node_modules/ node_modules/
COPY build/ build/
CMD ["node", "build/index.js"]