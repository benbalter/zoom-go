# zoom-go

Port of https://github.com/benbalter/zoom-launcher to zoom. Go easy, it's my first Go project.

## Installation

To install, download the tarball for your OS and architecture from the [latest release](https://github.com/benbalter/zoom-go/releases/latest). Extract the archive and copy the `zoom` binary somewhere on your `${PATH}`. :tada:

```bash
$ cd ~/Downloads
$ tar zxvf zoom_0.2.0_macOS-64bit.tar.gz
$ cp zoom ~/bin
```

If you want to live on the edge and run the latest master instead, [install Go](https://golang.org/doc/install) ([also on homebrew](https://formulae.brew.sh/formula/go)), then run:

```bash
$ go install github.com/benbalter/zoom-go/cmd/zoom@master
```

This will install a `zoom` executable file into `$(go env GOPATH)/bin` (or `$GOBIN`, if set).

## Usage

Ensure the `zoom` binary is in your `$PATH`, and run `zoom`! That's all.

## Authorization

The first time you run `zoom`, you will see instructions for how to create a Google app in the Developer Console, authorize it to access your calendar, download credentials, then import the credentials into `zoom`. After you import, you should be walked through the process of authorizing in the browser. Paste the authorization code back into your terminal, and vòila, `zoom` will be all configured for your next run.
