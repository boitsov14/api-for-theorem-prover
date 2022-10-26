######################################
FROM boitsov14/graalvm-static-native-image AS jar-builder
WORKDIR /build
COPY prover.jar .
RUN native-image --static --libc=musl --no-fallback -jar prover.jar

######################################
FROM alpine:latest
COPY --from=boitsov14/minimal-bussproofs-latex /usr/local/texlive /usr/local/texlive
ENV PATH=/usr/local/texlive/bin/x86_64-linuxmusl:$PATH
WORKDIR /app
COPY --from=jar-builder /build/prover ./prover
COPY . .
EXPOSE 5000
