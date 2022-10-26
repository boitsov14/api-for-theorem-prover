######################################
FROM alpine:latest AS npm-installer
RUN apk add --no-cache nodejs npm
ENV NODE_ENV=production
WORKDIR /app
COPY package*.json ./
RUN npm install --omit=dev --no-progress

######################################
FROM alpine:latest
RUN apk add --no-cache bash nodejs
COPY --from=boitsov14/minimal-bussproofs-latex /usr/local/texlive /usr/local/texlive
ENV PATH=/usr/local/texlive/bin/x86_64-linuxmusl:$PATH
ENV JAVA_HOME=/opt/java/openjdk
COPY --from=eclipse-temurin:18-jre-alpine $JAVA_HOME $JAVA_HOME
ENV PATH=$JAVA_HOME/bin:$PATH
WORKDIR /app
COPY --from=npm-installer /app/node_modules ./node_modules
ENV NODE_ENV=production
COPY . .
EXPOSE 5000
CMD [ "node", "app.js" ]
