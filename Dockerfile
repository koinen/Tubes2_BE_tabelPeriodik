FROM golang:latest
 
WORKDIR /app
 
# Effectively tracks changes within your go.mod file
COPY go.mod .
COPY go.sum .
# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
 
RUN go mod download
 
# Copies your source code into the app directory
COPY ./src .
 
RUN go build -o /godocker
 
EXPOSE 8080
 
CMD [ "/godocker" ]