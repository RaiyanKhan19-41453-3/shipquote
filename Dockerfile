FROM golang:1.23-bookworm AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /shipquote .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /shipquote /shipquote
EXPOSE 8080
ENTRYPOINT ["/shipquote"]
