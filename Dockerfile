FROM golang:latest as builder
LABEL maintainer="orensimple"
WORKDIR /app
COPY . .
RUN env GIT_TERMINAL_PROMPT=1 go get github.com/orensimple/otus_events_scheduler
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .


######## Start a new stage #######
FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/config/config.yml ./config/config.yml
#EXPOSE 8088
CMD ["./main"]