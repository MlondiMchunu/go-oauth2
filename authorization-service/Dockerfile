FROM  golang:1.25-alpine as builder

WORKDIR /app

COPY . .

RUN go build -o /app/authorization-service 

FROM scratch

COPY --from=builder /app/authorization-service /authorization-service

ENTRYPOINT ["/app/authorization-service"]