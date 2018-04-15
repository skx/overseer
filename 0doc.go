// overseer is a remote protocol tester.
//
// It is designed to allow you to execute tests against remote hosts,
// raising alerts when they are down, or failing.
//
// The application is written in an extensible fashion, allowing new
// test-types to be added easily, and the notification of failures is
// handled in a flexible fashion too.
//
// There are two ways the application can run:
//
// Simple
//
// overseer can be used in a simple way by directly giving it a list
// of tests to process.  Each test will be executed locally and appropriate
// notifications triggered.
//
// The `local` sub-command allows this:
//
//    overseer local test.file.1 test.file.2 .. test.file.N
//
//
// Distributed
//
// If you have a lot of tests to run it might make sense to run the tests
// from a small pool of hosts.  This can be achieved by storing tests in
// a central redis queue.
//
// From the redis queue multiple workers can each fetch tests, and execute
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
// Breif documentation for the available sub-commands now follows.
//
package main
