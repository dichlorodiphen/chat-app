FROM golang:1.21

WORKDIR /usr/src/server

COPY . .

RUN go mod download \
    && go build

EXPOSE 8000

CMD ["./server"]



