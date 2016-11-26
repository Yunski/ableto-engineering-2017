FROM joonix/appengine

COPY . /go/src/app
RUN go get golang.org/x/crypto/bcrypt

CMD ["app.yaml", "--runtime=go"]
