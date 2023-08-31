FROM golang:1.19

RUN go version

ENV GOPATH=/

ENV POSTGRES_DB=avitodb
ENV POSTGRES_USER=avito
ENV POSTGRES_PORT=5432
ENV POSTGRES_PASSWORD=avitosecret
ENV DATABASE_HOST=database

COPY ./ ./

# install psql
RUN apt-get update
RUN apt-get -y install postgresql-client

# make wait-for-postgres.sh executable
RUN chmod +x wait-for-postgres.sh

# build go app
RUN go mod download
RUN go build -o ./cmd/main ./cmd/

CMD ["./cmd/main"]