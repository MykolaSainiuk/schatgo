FROM golang:1.21.6-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

# RUN go get ./...
RUN go mod download

COPY . .

EXPOSE 8080

# RUN chmod +x /bin

# CMD [ "./bin" ]
CMD [ "go", "run", "main.go" ]
