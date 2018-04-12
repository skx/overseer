[![Go Report Card](https://goreportcard.com/badge/github.com/skx/overseer)](https://goreportcard.com/report/github.com/skx/overseer)
[![license](https://img.shields.io/github/license/skx/overseer.svg)](https://github.com/skx/overseer/blob/master/LICENSE)
[![Release](https://img.shields.io/github/release/skx/overseer.svg)](https://github.com/skx/overseer/releases/latest)
[![gocover store](http://gocover.io/_badge/github.com/skx/overseer)](http://gocover.io/github.com/skx/overseer)

# Overseer

This is a toy repository which contains some code designed to allow me
to experiment with class-factories in golang.

It might look like a remote protocol-tester, in [golang](https://golang.org),
but it isn't really.



## Usage

Build the application as per usual golang standards.  Then launch
via:

     ./overseer config.file config.file.two config.file.three ... config.file.N

Each test will be parsed and executed one by one.  The results of the
tests will be output to the console and _also_ posted to  [purppura](https://github.com/skx/purppura), which is assumed to be running on the localhost.


## Status

The tests defined in [input.txt](input.txt) each work, demonstrating
the successful registration and lookup of protocol tests for:

* HTTP
* HTTPS
* Redis
* SSH

The test for FTP is deliberately broken, and tests for `rsync`, `imap`,
`smtp`, `ping`, etc, are missing.


## TODO

* Make purppura optional.
  * But right now it is, since submission-errors are ignored.
* Command-line flags.
  * For verbosity, global timeout, and the target purppura server.
