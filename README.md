SORACOM Endorse client (Golang version)
=======================================

Client library and CLI tool for SORACOM Endorse.
This library provides SIM authentication and key agreement (AKA) feature using SORACOM Endorse.


## How to use CLI

1. Install

```
go install github.com/soracom/endorse-client-go/cmd/endorse-cli
```

or download executable file from [release page](https://github.com/soracom/endorse-client-go/releases).

On Linux, `libpcsclite` should be installed before using endorse-cli.
  Ubuntu: `sudo apt install pcscd libpcsclite1 libpcsclite-dev && sudo reboot`


2. Plug a smart card reader / USB modem with SORACOM Air SIM to your computer


3. Run

```
endorse-cli
```

You will get a key (CK) if successfully authenticated.

Use `-h` option for more details.

On Linux, you might need to run the command as root. i.e. `sudo endorse-cli` to access USB modems or smart card readers.


## How to build from source code

```
go get -u github.com/soracom/endorse-client-go/...
cd $GOPATH/src/github.com/soracom/endorse-client-go/endorse
go test
cd ../cmd/endorse-client-go
go build
```
