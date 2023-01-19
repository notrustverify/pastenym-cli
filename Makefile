BINARY_NAME=pastenym
BINARY_FOLDER=bin

all: build

build:
	go build -o ${BINARY_NAME} *.go

compile:
	echo "Compiling for every OS and Platform"
	mkdir -p ${BINARY_FOLDER}
	GOOS=windows GOARCH=amd64 go build -o ${BINARY_FOLDER}/${BINARY_NAME}-amd64-windows.exe *.go
	GOOS=windows GOARCH=386 go build -o ${BINARY_FOLDER}/${BINARY_NAME}-386-windows.exe *.go
	GOOS=freebsd GOARCH=386 go build -o ${BINARY_FOLDER}/${BINARY_NAME}-386-freebsd *.go
	GOOS=linux GOARCH=386 go build -o ${BINARY_FOLDER}/${BINARY_NAME}-386-linux *.go
	GOOS=linux GOARCH=amd64 go build -o ${BINARY_FOLDER}/${BINARY_NAME}-amd64-linux *.go
	GOOS=linux GOARCH=arm go build -o ${BINARY_FOLDER}/${BINARY_NAME}-arm-linux *.go
	GOOS=linux GOARCH=arm64 go build -o ${BINARY_FOLDER}/${BINARY_NAME}-arm64-linux *.go
	GOOS=darwin GOARCH=amd64 go build -o ${BINARY_FOLDER}/${BINARY_NAME}-amd64-darwin *.go
 
run:
	mkdir -p ${BINARY_FOLDER}
	go build -o ${BINARY_FOLDER}/${BINARY_NAME} *.go
	./bin/${BINARY_NAME}

test: build
	
	echo "Test before release"
	echo "Add a paste"
	./${BINARY_FOLDER}/${BINARY_NAME} -text "Add a paste"

	echo "Add a public paste"
	./${BINARY_FOLDER}/${BINARY_NAME} -text "Add a public paste" -public

	echo "Add a ipfs paste"
	./${BINARY_FOLDER}/${BINARY_NAME} -text "Add a ipfs paste" -ipfs

	echo "Add burn after reading paste"
	./${BINARY_FOLDER}/${BINARY_NAME} -text "Add burn after reading paste" -burn
	
	sleep 3

	echo "Get a paste id key"
	./${BINARY_FOLDER}/${BINARY_NAME} -id 5c3g1-uJ -key 6x23zeietv45ho9jlx7o4in1045qh2c3
	
	sleep 3

	echo "Get a paste URL"
	./${BINARY_FOLDER}/${BINARY_NAME} -id "https://pastenym.ch/#/5c3g1-uJ&key=6x23zeietv45ho9jlx7o4in1045qh2c3"

	sleep 3

	echo "Get a public paste"
	./${BINARY_FOLDER}/${BINARY_NAME} -id NL_4bnBT

	sleep 3

	echo "Get a paste from IPFS"
	./${BINARY_FOLDER}/${BINARY_NAME} -id "https://pastenym.ch/#/SPeNlLtY&key=l72u6pj1y2hf26oz1fok9qx4rjdjo7wj"

	sleep 3

	echo "Get a paste with file"
	./${BINARY_FOLDER}/${BINARY_NAME} -id "https://pastenym.ch/#/VQxVQl2d&key=922fc0f59aeb4ae7493e68bd0a252c12"



clean:
	go clean
	rm ${BINARY_FOLDER}/*
