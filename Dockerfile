FROM golang:bullseye AS build

ENV TEST=true

WORKDIR /src/
ADD . /src/
RUN apt-get update && apt-get install -y golang gcc libgl1-mesa-dev xorg-dev
RUN bash -c 'if [[ $TEST ]]; then echo "Performing Test" && go test -v; else echo "Skipping test"; fi'
RUN bash -c 'if [[ $BUILD ]]; then echo "Performing Build" && go build -a -v .; else echo "Skipping Build"; fi'

