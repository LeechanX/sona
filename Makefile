all:
	make -C common/net/protocol
	make -C protocol
	mkdir -p bin
	go build agent/sona_agent.go
	go build broker/sona_broker.go
	go build cli/sona_cli.go
	mv sona_agent bin/
	mv sona_broker bin/
	mv sona_cli bin/
	go build admin/sona_web_server.go
	mv sona_web_server admin/
clean:
	rm -rf bin/
	rm -f admin/sona_web_server
	rm -f protocol/*.pb.go
