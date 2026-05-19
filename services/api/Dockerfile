FROM golang:1.26-alpine AS build

WORKDIR /src/services/api

COPY services/api/go.mod services/api/go.sum ./
RUN go mod download

COPY services/api ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api

FROM alpine:3.23

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

ENV PORT=8080
ENV DATABASE_PATH=/data/app.db

RUN mkdir -p /data && chown -R app:app /data

COPY --from=build /out/api /app/api

USER app

EXPOSE 8080

CMD ["/app/api"]
