FROM golang:1.24-bookworm AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN go build -o /medication './cmd/medication'

FROM gcr.io/distroless/base-debian12:nonroot
USER nonroot:nonroot
WORKDIR /
COPY --from=builder /medication /medication
CMD ["/medication"]
