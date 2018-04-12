[![Travis CI](https://img.shields.io/travis/skx/overseer/master.svg?style=flat-square)](https://travis-ci.org/skx/overseer)
[![Go Report Card](https://goreportcard.com/badge/github.com/skx/overseer)](https://goreportcard.com/report/github.com/skx/overseer)
[![license](https://img.shields.io/github/license/skx/overseer.svg)](https://github.com/skx/overseer/blob/master/LICENSE)
[![Release](https://img.shields.io/github/release/skx/overseer.svg)](https://github.com/skx/overseer/releases/latest)


# Overseer

This is a toy repository which contains some code designed to allow me
to experiment with class-factories in golang.

It might look like a remote protocol-tester, in [golang](https://golang.org),
but it isn't really.  Specifically compared to the obvious comparison
project, custodian, we lack the ability to pull tests via HTTP(S), and we're single-host only.


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

* FTP
* HTTP & HTTPS
  * Note that no certificate validation is coded explicitly.
* redis
* rsync
* SMTP
* SSH

Tests for `imap`, `ping`, etc, are missing.


## Address Families

Because we're living in exciting and modern times the `overseer` application
will handle both IPv4 and IPv6 connections.

This is achieved by duplicating tests.  For example given the following input-line:

     mail.steve.org.uk must run smtp

What actually happens is that _two_ tests are executed:

     176.9.183.102 must run smtp
     2a01:4f8:151:6083::102 must run smtp

This is achieved by resolving the target, `mail.steve.org.uk` in this case, and running the test against each result.

If your host is not running with dual-stacks you can disable a particular family via:

     # IPv6 only
     overseer -4=false [-6=true]

     # IPv4 only
     overseer [-4=true] -6=false

The default is to enable both IPv6 and IPv4 testing.


## TODO

* Update HTTP-test to work on IPv4 & IPv6
