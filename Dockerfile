FROM ghcr.io/ktraister/ew_messenger_tester:latest 

ENV TEST=true
ENV CGO_ENABLED=1
ENV GOOS=linux

WORKDIR /src/
ADD . /src/
RUN bash -c 'if [[ $TEST ]]; then echo "Performing Test" && go test -v; else echo "Skipping test"; fi'
RUN bash -c 'if [[ $BUILD ]]; then echo "Performing Build" && go build -a -v .; else echo "Skipping Build"; fi'

