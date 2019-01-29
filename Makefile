all: prepare udpb

prepare: 
	if [ ! -d bin ]; then mkdir bin; fi;

udpb:
	go build -o bin/udpb main/main.go

clean:
	rm bin/*
