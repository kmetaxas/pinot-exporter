# Build
FROM golang:alpine AS builder
# Install git.
RUN apk update && apk add --no-cache git
#WORKDIR $GOPATH/pinot-exporter
RUN mkdir /build
ADD . /build
WORKDIR /build
# Fetch dependencies.# Using go get.
#RUN go get -d -v
# Build the binary.
RUN CGO_ENABLED=0 go build
# STEP 2 build an image from scratch using the binary only
FROM scratch
# Copy our static executable.
COPY --from=builder /build/pinot-exporter /pinot-exporter
# Run the hello binary.
CMD ["/pinot-exporter"]
