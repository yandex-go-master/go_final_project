FROM golang:1.21.6

WORKDIR /app

COPY . ./

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /todo ./cmd/todo/todo.go

ENV TODO_PORT=7540
ENV TODO_DBFILE=scheduler.db
ENV TODO_PASSWORD=secret123

EXPOSE 7540

CMD ["/todo"]