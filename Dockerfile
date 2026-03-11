FROM golang:1.22 AS build
WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/auditor ./cmd/auditor

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=build /out/auditor /auditor
USER nonroot:nonroot
ENTRYPOINT ["/auditor"]