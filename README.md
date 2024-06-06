# matchmaker simulation

## Goal

Find low latency matches as quickly as possible.

## Strategy

For each new player entering the matchmaking pool:

* Set the player state to "NEW".

For players in state "NEW":

* Find all datacenters less than ideal threshold (50ms). If one or more ideal datacenters are found, go to the "IDEAL" state.

* If no datacenters found under ideal threshold, find all datacenters less than expand threshold (100ms). If one or more expanded datacenters are found, go to the "EXPAND" state.

* If no datacenters are found under the expand threshold, go to the "WARMBODY" state where the player will be donated to any datacenter queue to fill games.

Players in "IDEAL" state should quickly find a game and go to "PLAYING" state. 

* Any players that don't find a game within 10 seconds, go to "EXPAND" state.

Players in "EXPAND" state should quickly find a game and go to "PLAYING" state.

* Any players that don't find a game within 10 seconds, go to "WARMBODY" state.

Players in "WARMBODY" state donate themselves to all matchmaking queues. They just need to play somewhere.

* Any players that don't find a game within 10 seconds, go to "FAILED" state.


* In this state players expand from the nearest datacenter up to 100ms over 60 seconds.

* Any players in "EXPAND" state that don't find a game within 60 seconds, go to "WARMBODY" state.

For players in the "WARMBODY" state, regular matchmaking attempts have failed. Try to fill games with these players.

* Look across all datacenters in the region for games that need extra players to start, regardless of latency, and volunteer the warm body.

Players in "WARMBODY" state that do not find a match within n seconds, go into "BOTS" state.

Players in "PLAYING" or "BOTS" state go to "DELAY" state at the end of the match.

Players in "DELAY" state go to "NEW" state after the delay between matches time has completed.
