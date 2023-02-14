# Build - backend
FROM --platform=linux/amd64 docker.io/library/golang:1.17-buster AS backend-build
RUN DEBIAN_FRONTEND="noninteractive" apt-get -y install tzdata
RUN wget https://github.com/swaggo/swag/releases/download/v1.7.1/swag_linux_amd64.tar.gz -O - | tar -xz -C /tmp && cp /tmp/swag_linux_amd64/swag /usr/local/bin

WORKDIR /app/backend
COPY ./ .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./cmd/server/main.go ./cmd/server

ENV TZ=Asia/Seoul

EXPOSE 8080

WORKDIR /app/backend/cmd/server

ENTRYPOINT ["./server"]
CMD ["-webroot","/app/backend/web"]
