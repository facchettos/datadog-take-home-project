# Context
This is a take home project given to me by datadog a while back.
It got me past the take-home project stage. Feel free to clone/use it :) 

# Build

To build the project run `go build`, you need go 1.18 to build it, not that
I used anything specific from it, but I don't have anything else on my laptop
so you'll have to make do.

# Usage

To see all the flags, use the '-help' flags.

# Design

 ┌────────────────┐     ┌───────────────┐     ┌───────────────┐
 │                │     │               │     │               │
 │    ingestor    ├─────►   watcher     ├─────►   alerter     │
 │                │     │               │     │               │
 └────────────────┘     └───────────────┘     └───────────────┘

The ingestor and alerter are interfaces. That allows us to do two things:
-1 we can replace them easily to allow for different ones, which would allow
for different formats or protocols (for example).
-2 easier testing because we don't need to deal with io in testing and can 
just mock the interfaces. 

For the computations , I only do one pass over the file. The counting mechanism
is a sliding window, so that for each second we know how many hits we have and 
can asses if we have to trigger the alarm. To compute the number of hits over the
last period, we keep track of the hits that we had over the last window so we don't
have to compute it again.

As the csv file is not ordered, because it is updated after the request is
processed we need to set some timeout to know when to stop looking for
potential requests. This mechanism may use some memory so we need to be 
careful we setting the request timeout. A better way to deal with it would
have been to log as soon as the requests come in so we could have kept them
ordered and have a separate log file for status code , size and other
informations. 

# Testing

I used table driven test for most of the tests. 

# Logging

There is logging inside the program to report for csv errors or
output errors, it is sent to stderr and you can remove it from the 
output by running `./datadog-stuff 2&> /dev/null`  

# Dockerfile 

There is a simple dockerfile, though I doubt that it will be useful


# Time

I dont have a proper estimate because I had to scatter the work
over multiple days, but I would say that it took me between
6 and 8 hours. I had to redo a few things because I first started 
without the sliding window but with fixed window but I did not like
it. Also I didn't notice that the log file wasn't ordered when I 
first started the ingestor. Overall I think I should have been faster,
but I guess everyone thinks that :) 

