FROM golang:1.27-alpine AS build
WORKDIR /src
COPY . .
WORKDIR /src/chat-service
RUN go mod download
RUN CGO_ENABLED=0 go build -trimpath -o /chat-service ./cmd/chat-service
FROM alpine:3.22
RUN apk add --no-cache ca-certificates
COPY --from=build /chat-service /usr/local/bin/chat-service
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["chat-service"]
