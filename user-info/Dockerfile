FROM golang:latest
WORKDIR /common

COPY common .

WORKDIR /app

COPY user-info .

# Get packages
RUN go get .

#Build
RUN go build .

#Execute
CMD ["./userinfo"]

#Expose
EXPOSE 8080