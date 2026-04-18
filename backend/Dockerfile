FROM golang:1.25-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o server .

FROM alpine:3.20

WORKDIR /app

RUN mkdir -p /app/data /data

COPY --from=build /app/server /app/server

ENV PORT=8080
ENV DATA_DIR=/app/data

EXPOSE 8080

CMD ["/app/server"]
