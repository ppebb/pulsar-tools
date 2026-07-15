all:
	GODEBUG=embedfollowsymlinks=1 go build

release:
	echo "Building pulsar-tools"
	env GODEBUG=embedfollowsymlinks=1 GOOS=linux GOARCH=amd64 go build -o pulsar-tools

	echo "Building pulsar-tools.exe"
	env GODEBUG=embedfollowsymlinks=1 GOOS=windows GOARCH=amd64 go build -o pulsar-tools.exe

	echo "Building pulsar-tools-osx"
	env GODEBUG=embedfollowsymlinks=1 GOOS=darwin GOARCH=amd64 go build -o pulsar-tools-osx
