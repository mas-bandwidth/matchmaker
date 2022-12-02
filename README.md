# matchmaker simulation

## Goal

Find low latency matches as quickly as possible.

## Strategy

For each new player entering the matchmaking pool:

* Set the player state to "NEW".

For players in state "NEW":
	
* Find the lowest latency datacenter for the player.

* If this datacenter is < 35ms then it is considered ideal. Find any other datacenters within +10ms of the lowest latency datacenter and call this the set of "ideal datacenters" for the player.

* If one or more ideal datacenters exist, put the player into "IDEAL" state and add the player to the queue of each datacenter in the ideal set. Take the first match that comes back from ANY of the ideal datacenters.

For players in "IDEAL" state:

* Players in "IDEAL" state should quickly find a game within 60 seconds and go to "PLAYING" state. 

* Any players in "IDEAL" state that don't find a game within 60 seconds, go to "EXPAND" state.

If no ideal datacenters are available for the new player, but there are datacenters within 100ms, go to the "EXPAND" state.

* In this state players expand from the nearest datacenter up to 100ms over 60 seconds.

Players in "EXPAND" state should quickly find a game within 60 seconds and go to "PLAYING" state. 

* Any players in "EXPAND" state that don't find a game within 60 seconds, go to "WARMBODY" state.

For players in the "WARMBODY" state, regular matchmaking attempts have failed. Try to fill games with these players.

* Look across all datacenters in the region for games that need extra players to start, regardless of latency, and volunteer the warm body.

Players in "WARMBODY" state that do not find a match within n seconds, go into "BOTS" state.

Players in "PLAYING" or "BOTS" state go to "DELAY" state at the end of the match.

Players in "DELAY" state go to "NEW" state after the delay between matches time has completed.

