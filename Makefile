GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod
GOFMT = $(GOCMD) fmt
GOVET = $(GOCMD) vet
PACKAGENAME = easystore-ws
BINNAME = $(PACKAGENAME)

build: darwin 

all: darwin linux

darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -a -tags service -o bin/$(BINNAME).darwin cmd/$(PACKAGENAME)/*.go

linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -a -tags service -installsuffix cgo -o bin/$(BINNAME).linux cmd/$(PACKAGENAME)/*.go

clean:
	$(GOCLEAN) cmd/
	rm -rf bin

dep:
	cd cmd/$(PACKAGENAME); $(GOGET) -u
	$(GOMOD) tidy
	$(GOMOD) verify

fmt:
	cd cmd/$(PACKAGENAME); $(GOFMT)

vet:
	cd cmd/$(PACKAGENAME); $(GOVET)

check:
	go get honnef.co/go/tools/cmd/staticcheck
	~/go/bin/staticcheck -checks all,-S1002,-ST1003 cmd/$(PACKAGENAME)/*.go
