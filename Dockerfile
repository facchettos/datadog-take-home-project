FROM golang as builder

RUN mkdir /src
ADD . /src/

WORKDIR /src
RUN CGO_ENABLED=0 go build -o /app

FROM alpine 

COPY --from=builder /app /app
RUN chmod +x /app
CMD ["/app","-f asdf"]
