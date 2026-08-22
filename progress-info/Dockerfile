FROM golang:latest
WORKDIR /common

COPY common .

WORKDIR /app

COPY progress-info .

# Get packages
RUN go get .

#Build
RUN go build .

#Execute
CMD ["./progressinfo"]

#Expose
EXPOSE 8080