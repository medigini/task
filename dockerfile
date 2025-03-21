
FROM golang:1.23-alpine as builder

# install ca certificates
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

WORKDIR /app

# copy the go module files and downloadd dependencies
COPY go.mod go.sum ./
RUN go mod download

# copy the applicatioon source code and build the binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o myapp


FROM alpine:latest


RUN apk update && apk add --no-cache ca-certificates

# copy only the binary. from the build stage to the final image
COPY --from=builder /app/myapp /
EXPOSE 4000

# set the entry point for the container
ENTRYPOINT ["/myapp"]





