# First build the application in a container
FROM alpine:3.11 AS builder

# Install depdendencies
RUN apk add --no-cache go dep git
ENV GOPATH /go

# Copy sources and build
WORKDIR /go/src/multisneeuw
COPY . .
RUN dep ensure && go build


######

# Then run the application in a clean container
FROM alpine:3.11 AS runner

WORKDIR /workdir

COPY --from=builder /go/src/multisneeuw/multisneeuw /workdir/multisneeuw
COPY --from=builder /go/src/multisneeuw/client /workdir/client

# Set up the entry point to the application
ENTRYPOINT ["./multisneeuw"]
