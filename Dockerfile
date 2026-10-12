FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /rickroll-tui .
FROM alpine:3.20
COPY --from=build /rickroll-tui /usr/local/bin/rickroll-tui
EXPOSE 2222
ENV RICKROLL_BIND_ADDRESS=0.0.0.0
ENTRYPOINT ["/usr/local/bin/rickroll-tui"]
