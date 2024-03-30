FROM golang:1.22.1-alpine

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    APP_NAME=tinyurlsvc \
    PORT=8000

# Move to working directory /app
WORKDIR /app

# Copy the code into the container
COPY . .

# Build the application
RUN go mod tidy
RUN go mod vendor
RUN go build -o ${APP_NAME}

# Move to /dist directory as the place for resulting binary folder
WORKDIR /dist

# Copy binary from build to main folder
RUN cp /app/${APP_NAME} .

# Command to run when starting the container
CMD ./${APP_NAME}
EXPOSE ${PORT}
