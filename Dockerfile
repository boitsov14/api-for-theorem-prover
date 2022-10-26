######################################
FROM alpine:latest AS npm-installer
RUN apk add --no-cache nodejs npm
ENV NODE_ENV=production
WORKDIR /app
COPY package*.json ./
RUN npm install --omit=dev --no-progress

######################################
FROM boitsov14/graalvm-static-native-image AS jar-builder
WORKDIR /build
COPY prover.jar .
RUN native-image --static --libc=musl --no-fallback -jar prover.jar

######################################
FROM alpine:latest
RUN apk add --no-cache bash nodejs
COPY --from=boitsov14/minimal-bussproofs-latex /usr/local/texlive /usr/local/texlive
ENV PATH=/usr/local/texlive/bin/x86_64-linuxmusl:$PATH
WORKDIR /app
COPY --from=jar-builder /build/prover ./prover
COPY --from=npm-installer /app/node_modules ./node_modules
ENV NODE_ENV=production
COPY . .
EXPOSE 5000
CMD [ "node", "app.js" ]
