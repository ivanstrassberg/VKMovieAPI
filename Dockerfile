FROM golang:latest

WORKDIR /go/src/app

COPY . .

RUN go get -d -v ./...

RUN go build -o main .

ENV POSTGRES_USER=postgres
ENV POSTGRES_PASSWORD=root
ENV POSTGRES_DB=postgres
ENV POSTGRES_HOST=db
ENV POSTGRES_PORT=5433

EXPOSE 3000

CMD ["./main"]