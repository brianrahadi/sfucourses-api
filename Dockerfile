# The build stage
FROM golang:1.23 as builder
WORKDIR /app
COPY . .
# Build the API
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api cmd/api/*.go
# Build all the scripts and put them in bin directory
RUN mkdir -p bin
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/fetch-sections scripts/fetchSections/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/fetch-outlines scripts/fetchOutlines/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/sync-offerings scripts/syncOfferings/main.go || true
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/sync-instructors scripts/syncInstructors/main.go || true

# The run stage
FROM alpine:latest
WORKDIR /app
# Copy CA certificates and install bash
RUN apk --no-cache add ca-certificates bash

# Copy binaries and make them executable
COPY --from=builder /app/api .
COPY --from=builder /app/bin /app/bin
COPY --from=builder /app/docs/swagger.json /app/docs/

# Copy the data directories
COPY --from=builder /app/internal/store/json /app/internal/store/json

RUN chmod +x /app/api
RUN chmod +x /app/bin/fetch-sections
RUN chmod +x /app/bin/fetch-outlines
RUN chmod +x /app/bin/sync-offerings || true
RUN chmod +x /app/bin/sync-instructors || true

# Don't use EXPOSE - let Render handle the port
CMD ["./api"]