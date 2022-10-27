# Goをビルドするステージ
FROM golang:alpine as go-builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY *.go .
RUN go build -o server -ldflags="-s -w"

# jarファイルを元に静的バイナリを生成するステージ
FROM boitsov14/graalvm-static-native-image AS jar-builder
WORKDIR /build
COPY prover.jar reflection.json ./
RUN native-image --static --libc=musl --no-fallback -J-Dfile.encoding=UTF-8 -H:ReflectionConfigurationFiles=reflection.json -jar prover.jar

FROM alpine:latest
COPY --from=boitsov14/minimal-bussproofs-latex /usr/local/texlive /usr/local/texlive
ENV PATH=/usr/local/texlive/bin/x86_64-linuxmusl:$PATH
WORKDIR /app
COPY --from=go-builder /build/server server
COPY --from=jar-builder /build/prover prover
COPY .env .
EXPOSE 3000
ENTRYPOINT ["./server"]
