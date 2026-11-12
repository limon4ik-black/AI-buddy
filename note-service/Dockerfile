# Сборка: статический бинарь, чтобы в рантайм не тащить libc.
FROM golang:1.27-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags '-s -w' -o /note-service ./cmd/note-service

# Рантайм: alpine ради wget (в distroless его нет) и минимального размера.
FROM alpine:3.22
# База зон вшита в бинарь (import _ "time/tzdata"), поэтому пакет tzdata и сеть на сборке не нужны.
ENV TZ=Europe/Moscow
WORKDIR /app
COPY --from=build /note-service /note-service
RUN adduser -D -u 10001 note && mkdir -p /app/out/logs && chown -R note:note /app
USER note
EXPOSE 8079
ENTRYPOINT ["/note-service"]
