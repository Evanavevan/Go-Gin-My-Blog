FROM golang:1.13

ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct

WORKDIR /app

COPY . .

RUN go build -o main .

EXPOSE 8090

ENTRYPOINT ["./main"]