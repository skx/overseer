[![Travis CI](https://img.shields.io/travis/skx/overseer/master.svg?style=flat-square)](https://travis-ci.org/skx/overseer)
[![Go Report Card](https://goreportcard.com/badge/github.com/skx/overseer)](https://goreportcard.com/report/github.com/skx/overseer)
[![license](https://img.shields.io/github/license/skx/overseer.svg)](https://github.com/skx/overseer/blob/master/LICENSE)
[![Release](https://img.shields.io/github/release/skx/overseer.svg)](https://github.com/skx/overseer/releases/latest)


# Overseer

Overseer is a simple golang based remote protocol tester, which allows you to monitor the health of a network.

When tests fail, because hosts/services are down, alerts can be generated via a simple plugin-based system.  Currently there is only a single notification plugin distributed with the project, which uses the [purppura](https://github.com/skx/purppura/) notification system.

"Remote Protocol Tester" sounds a little vague, so to be more concrete this application lets you test services are running and has built-in support for testing:

* http-servers
* rsync-servers
* smtp-servers
* etc
* ..

Adding new protocols to be tested is simple.



## Installation

The following command should get/update `overseer` upon your system, assuming
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


### Running Locally

Assuming you have a "small" network you can then execute your tests
directly like this:

      $ overseer local -verbose test.file.1 test.file.2 .. test.file.N

Each specified file will then be parsed and the tests executed one by one.

Because `-verbose` has been specified the tests, and their results, will be output to the console.

In real-world situation you'd also define a [purppura](https://github.com/skx/purppura) end-point to submit notifications to:

     $ overseer local \
        -notifier=purppura \
        -notifier-data=http://localhost:8080/events \
        -verbose \
        test.file.1 test.file.2

I'd be happy to accept notification-modules for other systems, but for the
moment only `purppura` is available.

(It is assumed you'd add a cronjob to run the tests every few minutes.)


### Running from multiple hosts

If you have a large network the expectation is that the tests will take a long time to execute serially, so to speed things up you might want to run the tests
in parallel.   Overseer supports this via the use of a shared [redis](https://redis.io/) queue.

On __one__ host run the following to add your tests to the redis queue:

       $ overseer enqueue \
           -redis-host=queue.example.com:6379 \
           [-redis-pass='secret.here'] \
           test.file.1 test.file.2 .. test.file.N

This will parse the tests and add them to the redis queue, now on as many hosts as you wish you can now run an instance of the worker:

       $ overseer worker -verbose \
          -redis-host=queue.example.com:6379 \
          [-redis-pass='secret']

The `worker` sub-command watches the redis-queue, and executes tests as they become available.  Again note that you'll need to configure your notification too, as shown previously on the simpler setup.  Something like this should be sufficient:

       $ overseer worker \
          -verbose \
          -redis-host=queue.example.com:6379 \
          [-redis-pass=secret] \
          -notifier=purppura \
          -notifier-data=http://localhost:8080/events

(It is assumed you'd leave the workers running, under systemd or similar, and run `overseer enqueue ...` via cron to ensure the queue was constantly refilled with tests for the worker(s) to execute.)


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
   * This is required because we connect by IP address rather than hostname.
* ping
* POP3 & POP3S
   * Use `mail.example.com must run pop3s insecure` to skip TLS validation.
   * This is required because we connect by IP address rather than hostname.
* redis
* rsync
* SMTP
* SSH
* XMPP

Tests for other protocols will be added based upon need & demand.


## Address Families

Because we're living in exciting and modern times the `overseer` application
will handle both IPv4 and IPv6 connections.

This is achieved by duplicating tests at parse-time.  For example given the following input:

     mail.steve.org.uk must run smtp

What actually happens is that __two__ tests are generated:

     176.9.183.102 must run smtp
     2a01:4f8:151:6083::102 must run smtp

This is achieved by resolving the target, `mail.steve.org.uk` in this case, and running the test against each result.

If your host is not a dual-stacked host you can disable a particular family via:

     # IPv6 only
     $ overseer local -4=false

     # IPv4 only
     $ overseer local -6=false

**NOTE**: The default is to enable both IPv6 and IPv4 testing, and the same options are supported for the `overseer local` and `overseer worker` mode.
