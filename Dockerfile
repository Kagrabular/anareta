FROM golang:1.24 AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Disable CGO, target Linux, and produce a statically linked executable. don't need clib bindings, so we can be self-contained and static. If for some reason you need CGO you can simply flip this.
RUN CGO_ENABLED=0 GOOS=linux go build -a -o manager .

FROM gcr.io/distroless/static:nonroot

WORKDIR /

COPY --from=builder /workspace/manager /manager

EXPOSE 8080 9443

USER nonroot:nonroot

ENTRYPOINT ["/manager"]
