FROM golang:alpine
COPY --from=boitsov14/minimal-bussproofs-latex /usr/local/texlive /usr/local/texlive
ENV PATH=/usr/local/texlive/bin/x86_64-linuxmusl:$PATH
WORKDIR /work
COPY go-echo/go.mod go-echo/go.sum ./
RUN go mod download
