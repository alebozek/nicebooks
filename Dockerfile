FROM golang:latest AS builder

WORKDIR /build

COPY . .
RUN go mod download
RUN go build -o ./nicebooks

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /build/nicebooks ./nicebooks
COPY --from=builder /build/creds.env ./creds.env
COPY --from=builder /build/templates/ ./templates
COPY --from=builder /build/static ./static
CMD ["/app/nicebooks"]