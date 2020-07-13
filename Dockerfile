FROM golang:alpine AS builder

# Install git.
# Git is required for fetching the dependencies.
# Make is required for installing.
RUN apk update && apk add --no-cache git make

WORKDIR $GOPATH/src/github.com/Shopify/themekit
COPY . .

# Build the binary.
RUN make

FROM scratch

# Copy our static executable.
COPY --from=builder /go/bin/theme /go/bin/theme

# Run the binary.
ENTRYPOINT ["/go/bin/theme"]