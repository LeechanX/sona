all:
	make -C protocol
	mkdir -p bin
	go build agent/sona_agent.go
	go build broker/sona_broker.go
	go build cli/CLI.go
	mv sona_agent bin/
	mv sona_broker bin/
	mv CLI bin/
clean:
	rm -rf bin/
	rm -f protocol/*.pb.go
