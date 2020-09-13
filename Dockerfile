FROM golang:1.15.2-buster
WORKDIR /build

ADD go.mod go.sum ./
RUN go mod download

ADD . ./
RUN go test ./cmd && go build -o github-command ./cmd

FROM gcr.io/distroless/base
COPY --from=0 /build/github-command /github-command

ENTRYPOINT [ "/github-command" ]
