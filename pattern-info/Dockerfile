FROM golang:latest
WORKDIR /common

COPY common .

WORKDIR /app

COPY pattern-info .

# Get packages
RUN go get .

#Build
RUN go build .

#Execute
CMD ["./patterninfo"]

#Expose
EXPOSE 8080