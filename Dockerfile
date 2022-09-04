######################################
FROM alpine:latest AS latex-installer
RUN apk add --no-cache perl tar wget
WORKDIR /install-tl-unx
COPY texlive.profile .
RUN wget -nv https://mirror.ctan.org/systems/texlive/tlnet/install-tl-unx.tar.gz
RUN tar -xzf ./install-tl-unx.tar.gz --strip-components=1
RUN ./install-tl --profile=texlive.profile
RUN ln -sf /usr/local/texlive/*/bin/* /usr/local/bin/texlive
ENV PATH=/usr/local/bin/texlive:$PATH
RUN tlmgr install \
  bussproofs \
  dvipng \
  preview \
  standalone \
  varwidth \
  xkeyval

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
COPY --from=latex-installer /usr/local/texlive /usr/local/texlive
RUN ln -sf /usr/local/texlive/*/bin/* /usr/local/bin/texlive
ENV PATH=/usr/local/bin/texlive:$PATH
ENV JAVA_HOME=/opt/java/openjdk
COPY --from=eclipse-temurin:18-jre-alpine $JAVA_HOME $JAVA_HOME
ENV PATH=$JAVA_HOME/bin:$PATH
WORKDIR /app
COPY --from=npm-installer /app/node_modules ./node_modules
ENV NODE_ENV=production
COPY . .
EXPOSE 5000
CMD [ "node", "app.js" ]
