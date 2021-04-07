FROM golang:buster
RUN apt-get update && apt-get install -y git
WORKDIR /src
COPY . /src
RUN go build -ldflags "-X main.GITSHA=`git rev-list -1 HEAD`"
EXPOSE 8878
CMD ["./icewall"]
