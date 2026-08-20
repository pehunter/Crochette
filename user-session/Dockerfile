FROM golang:latest
WORKDIR /common

COPY common .

WORKDIR /app

COPY user-session .

# Get packages
RUN go get .

#Build
RUN go build .

#Execute
CMD ["./sessions"]

#Expose
EXPOSE 8080