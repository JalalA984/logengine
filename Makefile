CONFIG_PATH=${HOME}/.logengine/

.PHONY: init
init:
	mkdir -p ${CONFIG_PATH}

.PHONY: gencert
gencert:
	
	cfssl gencert \
		-initca test/ca-csr.json | cfssljson -bare ca

	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=test/ca-config.json \
		-profile=server \
		test/server-csr.json | cfssljson -bare server

	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=test/ca-config.json \
		-profile=client \
		test/client-csr.json | cfssljson -bare client

	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=test/ca-config.json \
		-profile=client \
		-cn="root" \
		test/client-csr.json | cfssljson -bare root-client

	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=test/ca-config.json \
		-profile=client \
		-cn="nobody" \
		test/client-csr.json | cfssljson -bare nobody-client

	mv *.pem *.csr ${CONFIG_PATH}


.PHONY: compile
compile:
	protoc api/v1/*.proto \
	--go_out=. \
	--go-grpc_out=. \
	--go_opt=paths=source_relative \
	--go-grpc_opt=paths=source_relative \
	--proto_path=.


.PHONY: run
run:
	go run cmd/getservers/main.go


.PHONY: fmt
fmt:
	go fmt ./...


$(CONFIG_PATH)/model.conf:
	cp test/model.conf $(CONFIG_PATH)/model.conf

$(CONFIG_PATH)/policy.csv:
	cp test/policy.csv $(CONFIG_PATH)/policy.csv

.PHONY: test
test: $(CONFIG_PATH)/policy.csv $(CONFIG_PATH)/model.conf
	go test -race ./...
# 	cd internal/server && go test -v -debug=true

	
.PHONY: clean
clean:
	rm -rf ${CONFIG_PATH}
	rm -f /tmp/metrics-*.log
	rm -f /tmp/traces-*.log


TAG ?= 0.0.1
build-docker:
	docker build -t github.com/jalala984/logengine:$(TAG) .

kind:
	kind load docker-image github.com/jalala984/logengine:0.0.1

helmunin:
	helm uninstall logengine

helmin:
	helm install logengine deploy/logengine

port:
	kubectl port-forward pod/logengine-0 18000:8400