FROM golang:latest
WORKDIR /app

COPY . .

# Get packages
RUN go get .

#Build
RUN go build .

#Execute
CMD ["./sessions"]

#Expose
EXPOSE 8080