FROM golang:1.22.4-alpine

ENV APP_HOME /app

WORKDIR $APP_HOME
RUN mkdir -p "$APP_HOME"

# Download all dependencies
RUN go install github.com/cosmtrek/air@latest
# RUN go mod download

COPY . .

EXPOSE 8000

#ENTRYPOINT ["air"]
CMD ["air", "-c", ".air.toml"]
