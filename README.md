[![Travis CI](https://img.shields.io/travis/skx/overseer/master.svg?style=flat-square)](https://travis-ci.org/skx/overseer)
[![Go Report Card](https://goreportcard.com/badge/github.com/skx/overseer)](https://goreportcard.com/report/github.com/skx/overseer)
[![license](https://img.shields.io/github/license/skx/overseer.svg)](https://github.com/skx/overseer/blob/master/LICENSE)
[![Release](https://img.shields.io/github/release/skx/overseer.svg)](https://github.com/skx/overseer/releases/latest)


# Overseer

Overseer is a simple and scalable [golang](https://golang.org/)-based remote protocol tester, which allows you to monitor the state of your network, and the services running upon it.  The results of each test are posted to a redis-host, where they can be processed by external systems.  (Sample processors are included, but the intention is that by using a queue the alerting mechanism is decoupled from the core of the project; allowing you to integrate with your preferred in-house choice.)

"Remote Protocol Tester" sounds a little vague, so to be more concrete this application lets you test services are running and has built-in support for testing:

* DNS-servers
  * via lookups of A, AAAA, MX, NS, or TXT records.
* FTP
* HTTP & HTTPS fetches.
   * HTTP basic-authentication is supported.
   * Requests may be DELETE, GET, HEAD, POST, PATCH, POST, & etc.
   * SSL certificate validation and expiration warnings are supported.
* IMAP & IMAPS
* MySQL
* ping / ping6
* POP3 & POP3S
* Postgres
* redis
* rsync
* SMTP
* SSH
* VNC
* XMPP

(The implementation of the protocol-handlers can be found beneath the top-level [protocols/](protocols/) directory in this repository.)

Tests to be executed are defined in a simple text-based format which has the general form:

     $TARGET must run $SERVICE [with $OPTION_NAME $VALUE] ..

You can see what the available tests look like in [the sample test-file](input.txt), and each of the available protocol-handlers are self-documenting which means you can view example usage via:

     ~$ overseer examples [pattern]

All of the protocol-tests transparently support testing IPv4 and IPv6 targets, although you may globally disable either address family if you prefer.



## Installation & Dependencies

The following command should fetch/update `overseer`, and install it upon
your system, assuming you have a working golang setup:

     $ go get -u github.com/skx/overseer
     $ go install github.com/skx/overseer

Beyond the compile-time dependencies overseer requires a [redis](https://redis.io/) server which is used for two things:

* As the storage-queue for parsed-jobs.
* As the storage-queue for test-results.

Because `overseer` can be executed in a distributed fashion tests are not
executed as they are parsed/read, instead they are inserted into a redis-queue.
Workers then poll the queue, and fetch/execute jobs as they become available.

In small-scale deployments it is probably sufficient to have a single worker,
and all the software running upon a single host.  For a larger number of
tests (1000+) it might make more sense to have a pool of hosts each running
a worker.

Because we don't want to be tied to a specific notification-system results
of each test are also posted to the same redis-host, which allows results to be retrieved and transmitted to your preferred notifier.

You can see more details of the [notification](#notification) later in this document.


## Executing Tests

Executing tests is a two-step process:

* First of all tests are parsed and inserted into a redis-queue.
* Secondly the tests are pulled from that queue and executed.

This might seem a little convoluted, however it is a great design if you
have a lot of tests to be executed because it allows you to deploy multiple
workers.  Instead of having a single host executing all the tests you can
can have 10 hosts, each watching the redis-queue pulling jobs, & executing
them as they become available.

In short using a central queue allows you to scale out the testing horizontally, ensuring that all the jobs are executed as quickly as they can be.

To add the jobs to the queue you should run:

       $ overseer enqueue \
           -redis-host=queue.example.com:6379 [-redis-pass='secret.here'] \
           test.file.1 test.file.2 .. test.file.N

This will parse the tests contained in the specified files, adding each of them to the (shared) redis queue.  Once all of the jobs have been parsed and inserted into the queue the process will terminate.

To drain the queue you can should now start a worker, which will fetch the tests and process them:

       $ overseer worker -verbose \
          -redis-host=queue.example.com:6379 [-redis-pass='secret']

To run jobs in parallel simply launch more instances of the worker, on the same host, or on different hosts.  Remember that you'll need to specify the MQ host upon which to publish the results too:

       $ overseer worker \
          -verbose \
          -redis-host=queue.example.com:6379 [-redis-pass=secret] \
          -mq=mq.example.com:1883

Beneath [systemd/](systemd/) you will find some sample service-files which can be used to deploy overseer upon a single host:

* A service to start a single worker, fetching jobs from a queue on the localhost.
* A service & timer to regularly populate the queue with fresh jobs to be executed.



## Smoothing Test Failures

To avoid triggering false alerts due to transient (network/host) failures
tests which fail are retried several times before triggering a notification.

This _smoothing_ is designed to avoid raising an alert, which then clears
upon the next overseer run, but the downside is that flapping services might
not necessarily become visible.

If you're absolutely certain that your connectivity is good, and that
alerts should always be raised for failing services you can disable this
retry-logic via the command-line flag `-retry=false`.



## Notification

The result of each executed tests is published as a simple JSON message to the `overseer.results` set of the specified redis-server.

Results are added to the list as the tests are executed, and it is assumed a
notifier will pop them off to trigger alerts to humans.

You can check the size of the results list at any time via `redis-cli` like so:

    $ redis-cli llen overseer.results
    (integer) 0

Each test result is submitted as a JSON object, with the following fields:

| Field Name | Field Value                                                     |
| ---------- | --------------------------------------------------------------- |
| `input`    | The input as read from the configuration-file.                  |
| `result`   | Either `passed` or `failed`.                                    |
| `error`    | If the test failed this will explain why.                       |
| `time`     | The time the result was posted, in seconds past the epoch.      |
| `target`   | The target of the test, either an IPv4 address or an IPv6 one.  |
| `type`     | The type of test (ssh, ftp, etc).                               |

**NOTE**: The `input` field will be updated to mask any password options which have been submitted with the tests.

Included in this repository are two simple "[bridges](bridges/)", which poll results and forward the alerts to more useful systems:

* `irc-bridge.go`
  * This posts test-failures to an IRC channel.
  * Tests which pass are not reported, to avoid undue noise on your channel.
* `purppura-bridge.go`
  * This forwards each test-result to a [purppura host](https://github.com/skx/purppura/)
  * From there alerts will reach a human via pushover.


## Configuration File

If you prefer to use a configuration-file over the command-line arguments
that is supported.  Each of the subcommands can process an optional JSON-based
configuration file.

The configuration file will override the default arguments, and thus
cannot easily be set by a command-line flag itself.  Instead you should
export the environmental variable `OVERSEER` with the path to a suitable
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
