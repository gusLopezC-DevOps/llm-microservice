FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /app/managed-llm-service .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /app/managed-llm-service /app/managed-llm-service
EXPOSE 8080
USER 65532:nonroot
ENTRYPOINT ["/app/managed-llm-service"]
