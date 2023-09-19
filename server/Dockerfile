FROM golang:1.21

WORKDIR /usr/src/server

COPY go.mod go.sum .
RUN go mod download

COPY . .
RUN go build

EXPOSE 8000

CMD ["./server"]



