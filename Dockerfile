FROM sqlc/sqlc AS sqlc
WORKDIR /src

COPY ./sqlc.yaml .
COPY ./internal/db ./internal/db

RUN ["/workspace/sqlc", "generate"]

FROM golang:1.26.3-alpine AS build
WORKDIR /src

COPY go.mod .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /bin/server ./cmd/server

FROM scratch
COPY --from=build /bin/server /bin/url_shortner
CMD ["/bin/url_shortner"]
