# jarファイルを元に静的バイナリを生成するステージ
FROM boitsov14/graalvm-static-native-image AS jar-builder
WORKDIR /build
COPY prover.jar reflection.json ./
RUN native-image --static --libc=musl --no-fallback -J-Dfile.encoding=UTF-8 -H:ReflectionConfigurationFiles=reflection.json -jar prover.jar

FROM python:alpine
COPY --from=boitsov14/minimal-bussproofs-latex /usr/local/texlive /usr/local/texlive
ENV PATH=/usr/local/texlive/bin/x86_64-linuxmusl:$PATH
WORKDIR /app
COPY --from=jar-builder /build/prover prover
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY .env .
COPY *.py .
ENV PYTHONUNBUFFERED 1
EXPOSE 3000
CMD ["python", "app.py"]
