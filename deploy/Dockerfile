FROM golang:1.27.4-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/orbit ./cmd/app

FROM public.ecr.aws/lambda/provided:al2023 AS lambda
COPY --from=build /out/orbit /var/task/bootstrap
ENTRYPOINT ["/usr/local/bin/aws-lambda-rie", "/var/task/bootstrap"]

FROM debian:bookworm-slim AS runner
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl && rm -rf /var/lib/apt/lists/*
COPY --from=build /out/orbit /usr/local/bin/orbit
ENTRYPOINT ["orbit"]
