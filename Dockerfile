# Build - backend
FROM --platform=linux/amd64 docker.io/library/golang:1.21 AS backend-build
RUN DEBIAN_FRONTEND="noninteractive" apt-get -y install tzdata
RUN wget https://github.com/swaggo/swag/releases/download/v1.16.3/swag_1.16.3_Linux_amd64.tar.gz -O - | tar -xz -C /tmp && cp /tmp/swag /usr/local/bin

WORKDIR /app/backend
COPY ./ .

RUN swag init -g ./cmd/server/main.go --parseDependency --parseInternal -o ./api/swagger

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/server ./cmd/server/main.go

ENV TZ=Asia/Seoul

EXPOSE 8080

WORKDIR /app/backend/bin

ENTRYPOINT ["./server"]
CMD ["-webroot","/app/backend/web"]
