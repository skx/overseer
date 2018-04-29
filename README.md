[![Travis CI](https://img.shields.io/travis/skx/overseer/master.svg?style=flat-square)](https://travis-ci.org/skx/overseer)
[![Go Report Card](https://goreportcard.com/badge/github.com/skx/overseer)](https://goreportcard.com/report/github.com/skx/overseer)
[![license](https://img.shields.io/github/license/skx/overseer.svg)](https://github.com/skx/overseer/blob/master/LICENSE)
[![Release](https://img.shields.io/github/release/skx/overseer.svg)](https://github.com/skx/overseer/releases/latest)


# Overseer

Overseer is a [golang](https://golang.org/) based remote protocol tester, which allows you to monitor the state of your network, and the services running upon it.  The results of each test are posted to an MQ-host, where they can be processed by external systems.  (Sample processors are included, but the intention is that by using a message-queue the alerting mechanism is decoupled from the core of the project; allowing you to integrate with your preferred in-house choice.)

"Remote Protocol Tester" sounds a little vague, so to be more concrete this application lets you test services are running and has built-in support for testing:

* DNS-servers
  * via lookups of A, AAAA, MX, NS, or TXT records.
* FTP
* HTTP & HTTPS fetches.
   * HTTP basic-authentication is supported.
   * Requests may be GET or POST.
   * SSL certificate validation and expiration warning is supported.
* IMAP & IMAPS
* MySQL
* ping
* POP3 & POP3S
* Postgres
* redis
* rsync
* SMTP
* SSH
* VNC
* XMPP

(The existing protocol-handlers can be found beneath the top-level [protocols/](protocols/) directory in this repository.)

Tests to be carried out are defined in a simple format which has the general
form:

     $target must run $service [with $option_name $option_value] ..

You can see what the available tests look like in [the example test-file](input.txt), and the protocol-handlers are also self-documenting so you can read example usage via:

     ~$ overseer examples [filter]

All of the protocol-tests transparently support both IPv4 and IPv6 hosts, although you may disable either address family if you prefer.



## Installation & Dependencies

The following command should get/update `overseer` upon your system, assuming
you have a working golang setup:

     $ go get -u github.com/skx/overseer

Rather than being tied to a specific notification system overseer submits the
result of each test to a message-queue.  (i.e. An instance of [mosquitto](http://mosquitto.org/) or similar.)

This allows you to quickly and easily hook up your own local notification
system, without the need to modify the overseer application itself.


## Usage

There are two ways you can use overseer:

* Locally.
   * For small networks, or a small number of tests.
* Via a queue
   * For huge networks, or a huge number of tests.

In both cases the way that you get started is to write a series of tests, which describe the hosts & services you wish to monitor.  You can look at the [sample tests](input.txt) to get an idea of what is permitted.


### Running Locally

Assuming you have a "small" network you can then execute your tests
directly like this:

      $ overseer local -verbose test.file.1 test.file.2 .. test.file.N

Each specified file will then be parsed and the tests executed one by one.

Because `-verbose` has been specified the tests, and their results, will be output to the console.

In real-world situation you'd also define an MQ-host too, such that the results
would be reported to it:

     $ overseer local \
        -mq=localhost:1883 \
        -verbose \
        test.file.1 test.file.2

(It is assumed you'd add a cronjob to run the tests every few minutes.)


### Running from multiple hosts

If you have a large network the expectation is that the tests will take a long time to execute serially, so to speed things up you might want to run the tests
in parallel.  Overseer supports distributed/parallel operation via the use of
a shared [redis](https://redis.io/) queue.

On __one__ host run the following to add your tests to the redis queue:

       $ overseer enqueue \
           -redis-host=queue.example.com:6379 [-redis-pass='secret.here'] \
           test.file.1 test.file.2 .. test.file.N

This will parse the tests contained in the specified input files, and add each of them to the (shared) redis queue.  Once the jobs have been inserted into the queue the process will terminate.

To drain the queue you can now start a worker, which will fetch the tests to be executed from the queue, and process them:

       $ overseer worker -verbose \
          -redis-host=queue.example.com:6379 [-redis-pass='secret']

To run jobs in parallel simply launch more instances of the worker, on the same host, or on different hosts.  Remember that you'll need to specify the MQ host upon which to publish the results:

       $ overseer worker \
          -verbose \
          -redis-host=queue.example.com:6379 [-redis-pass=secret] \
          -mq=localhost:1883

It is assumed you'd leave the workers running, under systemd or similar, and run a regular `overseer enqueue ...` via cron to ensure the queue is constantly refilled with tests for the worker(s) to execute.



## Smoothing Test Failures

To avoid triggering false alerts due to transient (network/host) failures
tests which fail are retried several times before triggering a notification.

This _smoothing_ is designed to avoid raising an alert, which then clears
shortly afterwards - on the next overseer run - but the downside is that
flapping services might not necessarily become visible.

If you're absolutely certain that your connectivity is good, and that
services should never fail _ever_ you can disable this via the command-line
flag `-retry=false`.



## Notification

The result of each of the tests is published as a simple JSON message to MQ.

If you're using the mosquitto-queue (recommended) you can use the included  `mosquitto_sub` command to watch the `overseer` channel in real-time like so:

    $ mosquitto_sub -h 127.0.0.1 -p 1883 -t overseer
    {"input":"http://www.steve.fi/ must run http with content 'https://steve.fi' with status '302'",
     "result":"passed",
     "target":"176.9.183.100",
     "time":"1525017261",
     "type":"http"}
    {"input":"localhost must run ssh with port '2222'",
     "result":"passed",
     "target":"127.0.0.1",
     "time":"1525017262",
     "type":"ssh"}

Each result is posted as a JSON object, with the following fields being used:

| Field Name | Field Value                                                     |
| ---------- | --------------------------------------------------------------- |
| `input`    | The input as read from the configuration-file.                  |
| `result`   | Either `passed` or `failed`.                                    |
| `error`    | If the test failed this will explain why.                       |
| `time`     | The time the result was posted, in seconds past the epoch.      |
| `target`   | The target of the test, either an IPv4 address or an IPv6 one.  |
| `type`     | The type of test (ssh, ftp, etc).                               |

Beneath the [bridges/](bridges/) directory you'll find some sample code
which can connect to an MQ host, read the test-results as they arrive, and
act upon them:

* `irc-bridge.go`
  * This posts test-failures to an IRC channel.
  * Tests which pass are not reported, to avoid undue noise on your channel.
* `purppura-bridge.go`
  * This forwards each test-result to a [purppura host](https://github.com/skx/purppura/)


## Configuration File

If you prefer to use a configuration-file over the command-line arguments
that is supported.  Each of the subcommands can process a JSON-based
configuration file, if it is present.

The configuration file will override the default arguments, and thus
cannot easily be set by a command-line flag itself.  Instead you should
export the environmental variable OVERSEER with the path to a suitable
file.

For example you might run:

     export OVERSEER=$(pwd)/overseer.json

Where the contents of that file are:

     {
         "IPV6": true,
         "IPv4": true,
         "MQ": "localhost:1883",
         "RedisHost": "localhost:6379",
         "RedisPassword": "",
         "Retry": true,
         "Timeout": 10,
         "Verbose": true
     }



## Future Changes / Development?

This application was directly inspired by previous work upon the [Custodian](https://github.com/BytemarkHosting/custodian) monitoring system.

Compared to custodian overseer has several improvements:

* All optional parameters for protocol tests are 100% consistent.
  * i.e. Any protocol specific arguments are defined via "`with $option_name $option_value`"
  * In custodian options were added in an ad-hoc fashion as they became useful/necessary.
* The parsing of optional arguments is handled outside the protocol-tests.
   * In overseer the protocol test doesn't need to worry about parsing options, they're directly available.
* Option values are validated at parse-time, in addition to their names
   * i.e. Typos in input-files will be detected as soon as possible.
* Protocol tests provide _real_ testing, as much as possible.
   * e.g. If you wish to test an IMAP/POP3/MySQL service this application doesn't just look for a banner response on the remote port, but actually performs a login.

Currently overseer is regarded as stable and reliable.  I'd be willing to implement more notifiers and protocol-tests based upon user-demand and submissions.

Steve
--
