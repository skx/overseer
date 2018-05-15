// overseer is a remote protocol tester.
//
// It is designed to allow you to execute tests against remote hosts,
// raising alerts when they are down, or failing.
//
// The application is written in an extensible fashion, allowing new
// test-types to be added easily, and the notification of failures is
// handled in a flexible fashion too via the use of MQ.
//
// The application is designed to run in a distributed fashion, although
// it is equally happy to run upon a single node.
//
// Each of the tests to be executed is parsed and stored in a redis
// queue, from where multiple workers can each fetch tests, and execute
// them as they become available.
//
// To get started you'd first add your tests to the queue:
//
//    overseer enqueue  -redis-host=redis.example.com:6379 \
//      test.file.1 test.file.2 .. test.file.N
//
// On a pool of machines you can await tests by starting:
//
//    overseer worker -redis-host=redis.example.com:6379
//
// The workers will run the tests as they become available, raising and
// clearing notifications as appropriaate.
//
//
// Brief documentation for the available sub-commands now follows.
//
package main
