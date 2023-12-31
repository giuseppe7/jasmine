FROM golang:alpine as builder

# Build variable
ARG BUILD_NUM

# Grab the ca-certificates
RUN apk update && apk --no-cache add ca-certificates && update-ca-certificates 

# CGO has to be disabled for alpine.
ENV CGO_ENABLED=0

# Copy and download dependency using go mod into the build folder.
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container and build it.
COPY . .
RUN	go build -ldflags "-X main.version=${BUILD_NUM}" ./cmd/jasmine/jasmine.go

# Move to /dist directory as the place for resulting binary folder and copy the binary.
WORKDIR /dist
RUN cp /app/jasmine .

# Build the smaller image
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /dist/jasmine /
EXPOSE 2112
ENTRYPOINT [ "/jasmine" ]