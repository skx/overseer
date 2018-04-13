[![Travis CI](https://img.shields.io/travis/skx/overseer/master.svg?style=flat-square)](https://travis-ci.org/skx/overseer)
[![Go Report Card](https://goreportcard.com/badge/github.com/skx/overseer)](https://goreportcard.com/report/github.com/skx/overseer)
[![license](https://img.shields.io/github/license/skx/overseer.svg)](https://github.com/skx/overseer/blob/master/LICENSE)
[![Release](https://img.shields.io/github/release/skx/overseer.svg)](https://github.com/skx/overseer/releases/latest)


# Overseer

Overseer is a simple golang based remote protocol tester, which allows you to monitor the health of a network, raising/clearing alerts with the [purppura](https://github.com/skx/purppura/) notification system based upon service availability.

"Remote Protocol Tester" sounds a little vague, so to be more concrete this application lets you test services are running and has built-in support for testing:

* http-servers
* rsync-servers
* smtp-servers
* ..

Adding new protocols to be tested is simple.


## Installation

The following command should get/update overseer upon your system, assuming
you have a working golang setup:

     $ go get -u github.com/skx/overseer


## Usage

There are two ways you can use overseer:

* Locally.
   * For small networks.
* Via a queue
   * For huge networks.

In both cases the way that you get started is to write a series of tests,
these are the tests which describe the hosts & services you wish to monitor.

You can look at the [sample tests](input.txt) to get an idea of what is permitted.

Assuming you have a "small" network you can then execute your tests like so:

      $ overseer local test.file.1 test.file.2 .. test.file.N

Each file will be parsed and executed one by one.  The results of the
tests will be output to the console and if the end-point of a [purppura](https://github.com/skx/purppura) server is defined it receive the results of the tests too.

If you have a large network the expectation is that the tests will take a long time to execute serially, so to speed things up you might want to run the tests
in parallel.   Overseer supports this via a shared instance of [redis](https://redis.io/).

On __one__ host run the following to add your tests to the redis queue:

       $ overseer enqueue -redis=queue.example.com:6379 \
           test.file.1 test.file.2 .. test.file.N

On as many hosts as you wish you can now run workers which will await tests, and execute them in turn:

       $ overseer worker -redis=queue.example.com:6379


## Status

The tests defined in [input.txt](input.txt) each work, demonstrating
the successful registration and lookup of protocol tests for:

* DNS
  * Lookups of A, AAAA, MX, NS, and TXT records.
* FTP
* HTTP & HTTPS
   * Note that no certificate validation is coded explicitly.
* IMAP & IMAPS
   * Use `mail.example.com must run imaps insecure` to skip TLS validation.
* ping
* POP3
* redis
* rsync
* SMTP
* SSH

Tests for other protocols will be added based upon need & demand.


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
     overseer local -4=false [-6=true]

     # IPv4 only
     overseer local [-4=true] -6=false

The default is to enable both IPv6 and IPv4 testing.
