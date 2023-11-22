proto-gen:
	mkdir -p ./protogen && protoc -I=. --go_out=./protogen proto/person.proto
