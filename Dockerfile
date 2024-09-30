FROM golang:1.23-alpine
RUN apk add --no-cache gcc musl-dev sqlite-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN mkdir -p /app/database
RUN go build -o main ./cmd/server
EXPOSE 3030

CMD ["./main"]