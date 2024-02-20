FROM golang:1.21.6-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

# RUN go get ./...
RUN go mod download

COPY . .

RUN go build -o ./bin

EXPOSE 8080

RUN chmod +x ./bin

CMD [ "./bin" ]
# CMD [ "go", "run", "main.go" ]

# docker build --platform=linux/amd64 -t schatgo .  
# docker run --platform=linux/amd64 --env-file ./.env --network host schatgo 