# matchmaker simulation

## Goal

Find low latency matches as quickly as possible.

## Strategy

For each new player entering the matchmaking pool:

* Set the player state to "NEW".

For players in state "NEW":

* Find all datacenters less than ideal threshold (50ms). If one or more datacenters are found, go to the "IDEAL" state.

* If no datacenters found under ideal threshold, find all datacenters less than expand threshold (100ms). If one or more datacenters are found, go to the "EXPAND" state.

* If no datacenters are found under the expand threshold, go to the "WARMBODY" state.

Players in "IDEAL" state should quickly find a game and go to "PLAYING" state. 

* Any players that don't find a game within 10 seconds, go to "EXPAND" state.

Players in "EXPAND" state should quickly find a game and go to "PLAYING" state.

* Any players that don't find a game within 10 seconds, go to "WARMBODY" state.

Players in "WARMBODY" state donate themselves to all matchmaking queues. They just need to play somewhere.

* Look across all datacenters in the region for games that need extra players to start, regardless of latency, and volunteer the warm body.

* Any players that don't find a game within 10 seconds, go to "FAILED" state.

Players in "PLAYING" state go to "BETWEEN MATCH" state at the end of the match.

Players in "BETWEEN MATCH" go to "NEW" state once the time between matches has elapsed.
