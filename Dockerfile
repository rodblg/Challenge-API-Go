FROM golang:1.19

WORKDIR /usr/src/app

COPY go.mod go.sum ./

RUN go mod download

COPY ./app .

#RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-gs-ping

CMD ["go", "run", "main.go"]