build: build-linux build-windows

build-linux: build-linux-386 build-linux-amd64

build-linux-386:
	GOOS=linux GOARCH=386 go build -o mirach_linux_386

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o mirach_linux_amd64

build-windows: build-windows-386 build-windows-amd64

build-windows-386:
	GOOS=windows GOARCH=386 go build -o mirach_win_386.exe

build-windows-amd64:
	GOOS=windows GOARCH=amd64 go build -o mirach_win_amd64.exe
