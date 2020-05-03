BINARIES=$(shell ls cmd)

build: $(BINARIES)

$(BINARIES):
	GOOS=linux go build -ldflags="-s -w" -o build/$@ cmd/$@/main.go
	zip -j build/$@.zip build/$@
	rm build/$@