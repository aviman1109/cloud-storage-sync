FROM golang:alpine AS build
WORKDIR /app
COPY . .
RUN go mod tidy && \
    echo "start building..." && \
    GOOS=linux GOARCH=amd64 go build -o /cloud-storage-sync .

FROM gcr.io/google.com/cloudsdktool/google-cloud-cli:alpine
RUN apk add --no-cache tzdata jq
ENV TZ=Asia/Taipei
COPY --from=build /cloud-storage-sync /cloud-storage-sync
RUN chmod +x /cloud-storage-sync
ENTRYPOINT ["/cloud-storage-sync"]