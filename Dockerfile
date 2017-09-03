FROM golang:1.7.3
WORKDIR /go/src/github.com/chrisurwin/autospotting/
RUN go get -d -v golang.org/x/net/html  
ADD . ./
RUN CGO_ENABLED=0 GOOS=linux go get
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o autospotting .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/chrisurwin/autospotting/autospotting .
CMD ["./autospotting"] 