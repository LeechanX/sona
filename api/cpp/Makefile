all:
	g++ -c base_protocol.pb.cc
	g++ -c conf_memory.c
	g++ -c network.cc
	g++ -c sona_api.cc
	ar cqs libsonaapi.a sona_api.o base_protocol.pb.o conf_memory.o network.o

test:
	g++ -o example example.cc libsonaapi.a -lprotobuf -lpthread

clean:
	rm -f *.o *.a example
