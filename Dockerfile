# syntax=docker/dockerfile:1
FROM golang:1.18 as builder
RUN mkdir /source
COPY ./src /source
WORKDIR /source
RUN go build -o buildpackpackbuilder .

FROM golang:1.18
RUN apt-get update
RUN apt-get -y install ruby-full
RUN gem install bundler
RUN apt-get -y install direnv
RUN apt-get -y install jq
RUN mkdir /build /app
RUN go install github.com/cloudfoundry/libbuildpack/packager/buildpack-packager@master
COPY runapp.sh /app/runapp.sh
COPY --from=builder /source/buildpackpackbuilder /app/buildpackpackbuilder
RUN chmod +x /app/runapp.sh
RUN export GOPATH=/build
WORKDIR /build
ENTRYPOINT [ "/app/runapp.sh"]