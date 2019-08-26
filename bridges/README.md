# Bridges

Overseer only submits the results of the tests it executes to a redis queue,
in order to actually inform a human about a failure you need to process the
result-queue, and pass the messages on.

Note that each test `overseer` executes is stateless, so if you have a failing
test the notification will be repeated.

To give a concrete example, assume the following test:

    http://example.com/ must run http

If the remote host is offline _every_ time that overseer executes that
test it will record a fresh failure so if you're using the email bridge
you'll receive a fresh email each time the test is executed.

> (The purppura-bridge keeps local state, so it will ensure that humans are only notified once - even though it itself is updated at the end of every run.)

The following bridges are distributed with `overseer`:

* [email-bridge](email-bridge/)
   * Submits test-failures via email.
     * Test results which succeed are discarded.
* [purppura-bridge](purppura-bridge/)
   * Posts test results to a [purppura](https://github.com/skx/purppura/)-instance.
