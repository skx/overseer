# Systemd Examples

This directory contains some sample systemd configuration-files for running
overseer upon a single host.

The expectation is:

* You have overseer deployed as `/opt/overseer/bin/overseer`
* Your local host is running MQ + Redis
* You wish to execute all the tests available as `/opt/overseer/tests.d/*.conf`

The goal is to:

* Start a single worker to execute tests
  * These tests will be pulled from redis running on localhost
* Have a timer which will populate the redis-queue
  * This will refill the queue every two minutes


## Installation

Copy the files into the correct location:

     cp oveseer* /lib/systemd/system/

Enable the worker:

     # systemctl daemon-reload
     # systemctl enable overseer-worker.service
     # systemctl start overseer-worker.service

Now start the timer:

     # systemctl enable overseer-enqueue.timer
     # systemctl start overseer-enqueue.timer


## Sanity Checking

You can see the state of the worker, and any output it produces, via:

     # systemctl status overseer-worker.service

The cron-job to populate the queue is implemented as a (one-shot) service
and a corresponding timer to trigger it.  To view the status of the timer:

     # systemctl list-timers
     ..
     Tue 2018-05-15 09:42:00 UTC  17s left Tue 2018-05-15 09:40:00 UTC  1min 41s ago overseer-enqueue.timer       overseer-enqueue.service
     ..

Finally you can look for errors parsing the files via:

      # journalctl -u overseer-enqueue.service
