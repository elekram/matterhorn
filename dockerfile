FROM golang:1.22.5-alpine

ENV APP_HOME /app

WORKDIR $APP_HOME
RUN mkdir -p "$APP_HOME"

# Download all dependencies
RUN go install github.com/air-verse/air@latest
RUN go install github.com/a-h/templ/cmd/templ@latest
# RUN go mod download && go mod verify
# RUN go mod download

COPY . .
RUN go mod download && go mod verify

EXPOSE 8000

#ENTRYPOINT ["air"]
CMD ["air", "-c", ".air.toml"]
