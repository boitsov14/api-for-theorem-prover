FROM boitsov14/java-node-latex:latest
ENV NODE_ENV=production
WORKDIR /app
COPY package*.json ./
RUN npm install --production --no-progress
COPY . .
CMD [ "node", "app.js" ]
