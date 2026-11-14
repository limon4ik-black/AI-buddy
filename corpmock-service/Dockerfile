FROM golang:1.27-alpine AS build
WORKDIR /src
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -trimpath -o /service ./corpmock-service/cmd/corpmock
FROM alpine:3.22
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 buddy
COPY --from=build /service /service
USER buddy
ENTRYPOINT ["/service"]
