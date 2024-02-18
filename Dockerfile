FROM golang:1.21 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /wire-pod-mdns

FROM alpine:3.19 AS RUN

WORKDIR /

COPY --from=build /wire-pod-mdns /wire-pod-mdns

ENTRYPOINT ["/wire-pod-mdns"]