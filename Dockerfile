FROM golang:1.21.6-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o "./bin"

EXPOSE 8000

WORKDIR /app

RUN chmod +x /bin

CMD [ "./bin" ]
