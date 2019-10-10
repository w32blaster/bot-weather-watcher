#
# Step 1: compile the app
#
FROM golang as builder

WORKDIR /app
COPY . .

# compile app
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -o /app/bot \
    cmd/bot/main.go

# and initiate database, populating it with the locations
RUN mkdir -p storage
RUN go run cmd/create-resources/main.go

#
# Phase 2: prepare the runtime container, ready for production
#
FROM scratch

VOLUME "/storage"
EXPOSE 8444

# copy our bot executable
COPY --from=builder /app/bot /bot
COPY --from=builder /app/storage/weather.db /storage/weather.db

# copy root CA certificate to set up HTTPS connection with Telegram
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# copy timezone databases to be able to find London location zone
COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /usr/local/go/lib/time/zoneinfo.zip

CMD ["/bot"]