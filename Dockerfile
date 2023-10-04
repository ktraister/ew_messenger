FROM ghcr.io/ktraister/fyne_builder_nix:latest 

ENV TEST=true

WORKDIR /src/
ADD . /src/
RUN bash -c 'if [[ $TEST ]]; then echo "Performing Test" && go test -v; else echo "Skipping test"; fi'
RUN bash -c 'if [[ $BUILD ]]; then echo "Performing Build" && go build -a -v .; else echo "Skipping Build"; fi'

