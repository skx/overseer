[![Go Report Card](https://goreportcard.com/badge/github.com/skx/overseer)](https://goreportcard.com/report/github.com/skx/overseer)
[![license](https://img.shields.io/github/license/skx/overseer.svg)](https://github.com/skx/overseer/blob/master/LICENSE)
[![Release](https://img.shields.io/github/release/skx/overseer.svg)](https://github.com/skx/overseer/releases/latest)


Table of Contents
=================

* [Overseer](#overseer)
* [Installation &amp; Dependencies](#installation--dependencies)
  * [Source Installation go &lt;=  1.11](#source-installation-go---111)
  * [Source installation go  &gt;= 1.12](#source-installation-go---112)
  * [Dependencies](#dependencies)
* [Executing Tests](#executing-tests)
  * [Running Automatically](#running-automatically)
  * [Smoothing Test Failures](#smoothing-test-failures)
* [Notifications](#notifications)
* [Metrics](#metrics)
* [Redis Specifics](#redis-specifics)
* [Docker](#docker)
* [Github Setup](#github-setup)


# Overseer

Overseer is a simple and scalable [golang](https://golang.org/)-based remote protocol tester, which allows you to monitor the state of your network, and the services running upon it.

"Remote Protocol Tester" sounds a little vague, so to be more concrete this application lets you test that (remote) services are running, and has built-in support for performing testing against:

* DNS-servers
   * Test lookups of A, AAAA, MX, NS, and TXT records.
* Finger
* FTP
* HTTP & HTTPS fetches.
   * HTTP basic-authentication is supported.
   * Requests may be DELETE, GET, HEAD, POST, PATCH, POST, & etc.
   * SSL certificate validation and expiration warnings are supported.
* IMAP & IMAPS
* MySQL
* NNTP
* ping / ping6
* POP3 & POP3S
* Postgres
* redis
* rsync
* SMTP
* SSH
* Telnet
* VNC
* XMPP

(The implementation of the protocol-handlers can be found beneath the top-level [protocols/](protocols/) directory in this repository.)

Tests to be executed are defined in a simple text-based format which has the general form:

     $TARGET must run $SERVICE [with $OPTION_NAME $VALUE] ..

You can see what the available tests look like in [the sample test-file](input.txt), and each of the included protocol-handlers are self-documenting which means you can view example usage via:

     ~$ overseer examples [pattern]

All protocol-tests transparently support testing IPv4 and IPv6 targets, although you may globally disable either address family if you wish.



## Installation & Dependencies

There are two ways to install this project from source, which depend on the version of the [go](https://golang.org/) version you're using.

If you just need the binaries you can find them upon the [project release page](https://github.com/skx/overseer/releases).


### Source Installation go <=  1.11

If you're using `go` before 1.11 then the following command should fetch/update `overseer`, and install it upon your system:

     $ go get -u github.com/skx/overseer

### Source installation go  >= 1.12

If you're using a more recent version of `go` (which is _highly_ recommended), you need to clone to a directory which is not present upon your `GOPATH`:

    git clone https://github.com/skx/overseer
    cd overseer
    go install


### Dependencies

Beyond the compile-time dependencies overseer requires a [redis](https://redis.io/) server which is used for two things:

* As the storage-queue for parsed-jobs.
* As the storage-queue for test-results.

Because overseer is executed in a distributed fashion tests are not executed
as they are parsed/read, instead they are inserted into a redis-queue. A worker,
or number of workers, poll the queue fetching & executing jobs as they become
available.

In small-scale deployments it is probably sufficient to have a single worker,
and all the software running upon a single host.  For a larger number of
tests (1000+) it might make more sense to have a pool of hosts each running
a worker.

Because we don't want to be tied to a specific notification-system results
of each test are also posted to the same redis-host, which allows results to be retrieved and transmitted to your preferred notifier.

More details about [notifications](#notifications) are available later in this document.



## Executing Tests

As mentioned already executing tests a two-step process:

* First of all tests are parsed and inserted into a redis-based queue.
* Secondly the tests are pulled from that queue and executed.

This might seem a little convoluted, however it is a great design if you
have a lot of tests to be executed, because it allows you to deploy multiple
workers.  Instead of having a single host executing all the tests you can
can have 10 hosts, each watching the same redis-queue pulling jobs, & executing
them as they become available.

In short using a central queue allows you to scale out the testing horizontally.

To add your tests to the queue you should run:

       $ overseer enqueue \
           -redis-host=queue.example.com:6379 [-redis-pass='secret.here'] \
           test.file.1 test.file.2 .. test.file.N

This will parse the tests contained in the specified files, adding each of them to the (shared) redis queue.  Once all of the jobs have been parsed and inserted into the queue the process will terminate.

To drain the queue you can should now start a worker, which will fetch the tests and process them:

       $ overseer worker -verbose \
          -redis-host=queue.example.com:6379 [-redis-pass='secret']

The worker will run constantly, not terminating unless manually killed.  With
the worker running you can add more jobs by re-running the `overseer enqueue`
command.

To run tests in parallel simply launch more instances of the worker, on the same host, or on different hosts.



### Running Automatically

Beneath [systemd/](systemd/) you will find some sample service-files which can be used to deploy overseer upon a single host:

* A service to start a single worker, fetching jobs from a redis server.
  * The redis-server is assumed to be running on `localhost`.
* A service & timer to regularly populate the queue with fresh jobs to be executed.
  * i.e. The first service is the worker, this second one feeds the worker.



### Smoothing Test Failures

To avoid triggering false alerts due to transient (network/host) failures
tests which fail are retried several times before triggering a notification.

This _smoothing_ is designed to avoid raising an alert, which then clears
upon the next overseer run, but the downside is that flapping services might
not necessarily become visible.

If you're absolutely certain that your connectivity is good, and that
alerts should always be raised for failing services you can disable this
retry-logic via the command-line flag `-retry=false`.



## Notifications

The result of each test is submitted to the central redis-host, from where it can be pulled and used to notify a human of a problem.

Sample result-processors are [included](bridges/) in this repository which post
test-results to Telegram, a [purppura instance](https://github.com/skx/purppura), or via email.

The sample bridges are primarily included for demonstration purposes, the
expectation is you'll prefer to process the results and issue notifications to
humans via your favourite in-house tool - be it pagerduty, or something similar.

The results themselves are published as JSON objects to the `overseer.results` set.   Your notifier should remove the results from this set, as it generates alerts to prevent it from growing indefinitely.

You can check the size of the results set at any time via `redis-cli` like so:

    $ redis-cli llen overseer.results
    (integer) 0

The JSON object used to describe each test-result has the following fields:

| Field Name | Field Value                                                     |
| ---------- | --------------------------------------------------------------- |
| `input`    | The input as read from the configuration-file.                  |
| `result`   | Either `passed` or `failed`.                                    |
| `error`    | If the test failed this will explain why.                       |
| `time`     | The time the result was posted, in seconds past the epoch.      |
| `target`   | The target of the test, either an IPv4 address or an IPv6 one.  |
| `type`     | The type of test (ssh, ftp, etc).                               |

**NOTE**: The `input` field will be updated to mask any password options which have been submitted with the tests.

As mentioned this repository contains some demonstration "[bridges](bridges/)", which poll the results from Redis, and forward them to more useful systems:

* `email-bridge/main.go`
  * This posts test-failures via email.
  * Tests which pass are not reported.
* `purppura-bridge/main.go`
  * This forwards each test-result to a [purppura host](https://github.com/skx/purppura/).
  * From there alerts will reach a human via pushover.
* `telegram-bridge/main.go`
  * This forwards each test-failure as a message to a Telegram user.



## Metrics

Overseer has built-in support for exporting metrics to a remote carbon-server:

* Details of the system itself.
   * Via the [go-metrics](https://github.com/skx/golang-metrics) package.
* Details of the tests executed.
   * Including the time to run tests, perform DNS lookups, and retry-counts.

To enable this support simply export the environmental variable `METRICS`
with the hostname of your remote metrics-host prior to launching the worker.



## Redis Specifics

We use Redis as a queue as it is simple to deploy, stable, and well-known.

Redis doesn't natively operate as a queue, so we replicate this via the "list"
primitives.  Adding a job to a queue is performed via a "[rpush](https://redis.io/commands/rpush)" operation, and pulling a job from the queue is achieved via an "[blpop](https://redis.io/commands/blpop)" command.

We use the following two lists as queues:

* `overseer.jobs`
    * For storing tests to be executed by a worker.
* `overseer.results`
    * For storing results, to be processed by a notifier.

You can examine the length of either queue via the [llen](https://redis.io/commands/llen) operation.

* To view jobs pending execution:
   * `redis-cli lrange overseer.jobs 0 -1`
   * Or to view just the count
      * `redis-cli llen overseer.jobs`
* To view test-results which have yet to be notified:
   * `redis-cli lrange overseer.results 0 -1`
   * Or to view just the count
      * `redis-cli llen overseer.results`




## Docker

There are a series of Dockerfiles contained within this repository, they're designed to allow you to test things in a simple fashion.  However they do have the notification bridge hardcoded.

You can build the images like so:

```
docker build -t overseer:bridge  -f Dockerfile.bridge .
docker build -t oversser:enqueue -f Dockerfile.enqueue .
docker build -t overseer:worker  -f Dockerfile.worker .
```

Once built the supplied [docker-compose.yml](docker-compose.yml) file will let you launch them, using a shared redis instance.  The notifications will go via telegram by default, so you'll need to populate a token for a bot and setup your recipient user-ID.



## Github Setup

This repository is configured to run tests upon every commit, and when pull-requests are created/updated.  The testing is carried out via [.github/run-tests.sh](.github/run-tests.sh) which is used by the [github-action-tester](https://github.com/skx/github-action-tester) action.
Releases are automated in a similar fashion via [.github/build](.github/build), and the [github-action-publish-binaries](https://github.com/skx/github-action-publish-binaries) action.

Steve
--
