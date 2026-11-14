# Сборка: статический бинарь, чтобы в рантайм не тащить libc.
FROM golang:1.27-bookworm AS build
WORKDIR /src
COPY . .
WORKDIR /src/note-service
RUN go mod download
RUN CGO_ENABLED=0 go build -trimpath -ldflags '-s -w' -o /note-service ./cmd/note-service && CGO_ENABLED=0 go build -trimpath -o /import-onboarding ./cmd/import-onboarding

# Рантайм: alpine ради wget (в distroless его нет) и минимального размера.
FROM alpine:3.22
# База зон вшита в бинарь (import _ "time/tzdata"), поэтому пакет tzdata и сеть на сборке не нужны.
ENV TZ=Europe/Moscow
WORKDIR /app
COPY --from=build /note-service /note-service
COPY --from=build /import-onboarding /import-onboarding
RUN adduser -D -u 10001 note && mkdir -p /app/out/logs && chown -R note:note /app
USER note
EXPOSE 8079
ENTRYPOINT ["/note-service"]
