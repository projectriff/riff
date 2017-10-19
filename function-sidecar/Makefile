.PHONY: build clean fetch-grpc grpc dockerize
OUTPUT = function-sidecar
GRPC_DIR = pkg/dispatcher/grpc

build: $(OUTPUT)

$(OUTPUT): vendor
	go build cmd/function-sidecar.go

vendor: Gopkg.toml
	dep ensure

clean:
	rm $(OUTPUT)

grpc: $(GRPC_DIR)/fntypes/fntypes.pb.go $(GRPC_DIR)/function/function.pb.go

$(GRPC_DIR)/fntypes/fntypes.pb.go: $(GRPC_DIR)/fntypes/fntypes.proto
	protoc -I $(GRPC_DIR)/fntypes fntypes.proto --go_out=$(GRPC_DIR)/fntypes
$(GRPC_DIR)/function/function.pb.go: $(GRPC_DIR)/function/function.proto $(GRPC_DIR)/fntypes/fntypes.pb.go
	protoc -I $(GRPC_DIR)/fntypes -I $(GRPC_DIR)/function function.proto --go_out=Mfntypes.proto=github.com/sk8sio/function-sidecar/pkg/dispatcher/grpc/fntypes,plugins=grpc:$(GRPC_DIR)/function

fetch-grpc:
	rm -f $(GRPC_DIR)/fntypes/*.proto $(GRPC_DIR)/function/*.proto
	wget https://raw.githubusercontent.com/markfisher/sk8s/master/function-proto/src/main/proto/fntypes.proto \
		-P $(GRPC_DIR)/fntypes
	wget https://raw.githubusercontent.com/markfisher/sk8s/master/function-proto/src/main/proto/function.proto \
	     -P $(GRPC_DIR)/function

dockerize:
	docker build . -t sk8s/function-sidecar:v0001	  