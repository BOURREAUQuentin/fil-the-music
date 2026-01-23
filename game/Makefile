export PATH := $(PATH):$(shell go env GOPATH)/bin

# Installe les outils nécessaires (à faire une fois)
install-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Génère le code Go à partir du fichier .proto
gen:
	protoc --proto_path=. \
	--go_out=. --go_opt=paths=source_relative \
	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
	api/proto/v1/*.proto
	@echo "Code gRPC généré avec succès !"

# Nettoie le projet
clean:
	rm -f api/proto/v1/*.pb.go