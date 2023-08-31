FROM golang:1.19

RUN go version

ENV GOPATH=/

ENV POSTGRES_DB=avitodb
ENV POSTGRES_USER=avito
ENV POSTGRES_PORT=5432
ENV POSTGRES_PASSWORD=avitosecret
ENV DATABASE_HOST=database

COPY ./ ./

# build go app
RUN go mod download
RUN go build -o ./cmd/main ./cmd/

CMD ["./cmd/main"]