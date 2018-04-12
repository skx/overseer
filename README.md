[![Travis CI](https://img.shields.io/travis/skx/overseer/master.svg?style=flat-square)](https://travis-ci.org/skx/overseer)
[![Go Report Card](https://goreportcard.com/badge/github.com/skx/overseer)](https://goreportcard.com/report/github.com/skx/overseer)
[![license](https://img.shields.io/github/license/skx/overseer.svg)](https://github.com/skx/overseer/blob/master/LICENSE)
[![Release](https://img.shields.io/github/release/skx/overseer.svg)](https://github.com/skx/overseer/releases/latest)


# Overseer

This is a toy repository which contains some code designed to allow me
to experiment with class-factories in golang.

It might look like a remote protocol-tester, in [golang](https://golang.org),
but it isn't really.



## Usage

Build the application as per usual golang standards.  Then launch
via:

      $ ./overseer config.file config.file.two config.file.three ... config.file.N

Each test will be parsed and executed one by one.  The results of the
tests will be output to the console and if the end-point of a [purppura](https://github.com/skx/purppura) server is defined it receive the results of the tests too.

For example:

       $ ./overseer -purppura=http://localhost:8080/events input.txt


## Status

The tests defined in [input.txt](input.txt) each work, demonstrating
the successful registration and lookup of protocol tests for:

* HTTP & HTTPS
  * Note that no certificate validation is coded explicitly.
* Redis
* SSH

The test for FTP is deliberately broken, and tests for `rsync`, `imap`,
`smtp`, `ping`, etc, are missing.
