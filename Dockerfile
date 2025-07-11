FROM golang:1.24.4-alpine AS builder

WORKDIR /app

# Add group and user
RUN addgroup --system --gid 700 jag && adduser --system --uid 700 --ingroup jag jag

# Copy modules files
COPY go.mod ./
COPY go.sum ./

# Download Go modules
RUN go mod download

# Copy the rest of the files
COPY . .

# Build binary
RUN CGO_ENABLED=0 go build -o jag main.go

FROM scratch

ENV LISTEN_ADDRESS="0.0.0.0"
ENV LIBRARY_PATH="/library"
ENV THUMBNAILS_PATH="/thumbnails"

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/jag /usr/bin/jag

USER jag

EXPOSE 8080

ENTRYPOINT ["/usr/bin/jag"]
