# FROM golang:alpine AS builder

# LABEL stage=gobuilder

# ENV CGO_ENABLED 0
# ENV GOOS linux
# ENV GOPROXY https://goproxy.cn,direct

# WORKDIR app

# ADD go.mod .
# ADD go.sum .
# RUN go mod download
# RUN go build -ldflags="-s -w" -o table-demo-app 

FROM alpine:3.14

WORKDIR /home/
COPY ./app/app ./
CMD ["./app", "-f", "./etc/app.yml"]