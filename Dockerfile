FROM golang:1.21.6-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

# RUN go get ./...
RUN go mod download

COPY ./src ./src
COPY ./main.go ./

RUN go build -o ./bin

EXPOSE 8000

RUN chmod +x /bin

CMD [ "./bin" ]

EXPOSE 8080

# CMD [ "go", "run", "main.go" ]