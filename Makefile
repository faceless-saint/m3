
SRC_LIB := lib/*
SRC_INSTALL := install/*

.PHONY: all build package clean

all : package

build : bin/m3-install bin/m3-install.exe bin/m3-install_osx

bin/m3-install : $(SRC_LIB) $(SRC_INSTALL)
	mkdir -p bin && GOOS=linux GOARCH=amd64 go build -o $@ ./install

bin/m3-install.exe : $(SRC_LIB) $(SRC_INSTALL)
	mkdir -p bin && GOOS=windows GOARCH=amd64 go build -o $@ ./install

bin/m3-install_osx : $(SRC_LIB) $(SRC_INSTALL)
	mkdir -p bin && GOOS=darwin GOARCH=amd64 go build -o $@ ./install

package : dist/m3-install_linux.tar.gz dist/m3-install_windows.zip dist/m3-install_osx.zip

dist/m3-install_linux.tar.gz : bin/m3-install
	mkdir -p dist && cd bin && tar -czf ../dist/m3-install_linux.tar.gz m3-install

dist/m3-install_windows.zip : bin/m3-install.exe
	mkdir -p dist && cd bin && zip -q ../dist/m3-install_windows.zip m3-install.exe

dist/m3-install_osx.zip : bin/m3-install_osx
	mkdir -p dist && cd bin && zip -q ../dist/m3-install_osx.zip m3-install_osx

clean :
	rm -rf bin/ dist/

# Add external dependencies
${GOPATH}/src/github.com/cavaliercoder/grab :
	go get github.com/cavaliercoder/grab
