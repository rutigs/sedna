from golang:1.16-alpine as builder
WORKDIR /bin
RUN apk update && apk upgrade && apk add --no-cache ca-certificates

# Run these steps separately from building so that imports can be cached
# into docker layers on separate from code changes to speed up docker build
COPY go.mod go.sum ./
RUN go mod download

# We need to build a static binary because the application won't have
# libs available to link to in the final image
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app .

# Use a 2 stage from scratch to minimize image size
FROM scratch
COPY --from=builder bin/app .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
CMD ["./app"]
