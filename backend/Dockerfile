FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG SERVICE
RUN CGO_ENABLED=0 go build -trimpath -o /service ./cmd/${SERVICE}
FROM alpine:3.22
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 buddy
COPY --from=build /service /service
USER buddy
ENTRYPOINT ["/service"]
