
GOBUILDFLAGS += -ldflags -s

OUTPUT_DIR=./_output

build:
	go build ${GOBUILDFLAGS} -o ${OUTPUT_DIR}/inceptions ./cmd/

product:
	env GOOS=linux GOARCH=amd64 go build ${GOBUILDFLAGS} -o ${OUTPUT_DIR}/inceptions.linux ./

test:
	go test $(GOBUILDFLAGS)

clean:
	rm -rf ./${OUTPUT_DIR}
