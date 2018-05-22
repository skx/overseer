# Bridges

Overseer only posts the results of the tests to a redis host, so if
you wish to raise alerts to people you will need something to watch
that queue and take the appropriate action.

This directory contains two utilities:

* [purppura-bridge](purppura-bridge/)
   * Posts test results to a [purppura](https://github.com/skx/purppura/)-instance.
* [irc-bridge](irc-bridge/)
   * Posts test-failures to an IRC channel.
     * Test results which succeed are discarded.
