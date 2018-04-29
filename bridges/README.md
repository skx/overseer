# MQ Bridges

Overseer only posts the results of the tests to an MQ instance, so if
you wish to raise alerts to people you will need something to watch
that bridge and take the appropriate action.

This directory contains two utilities:

* `purppura-bridge.go`
   * Posts test results to a purppura-instance
* `irc-bridge.go`
   * Posts tests to an IRC instance
