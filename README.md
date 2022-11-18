# matchmaker simulation

## Goal

Find low latency matches as quickly as possible.

## Strategy

For each new player entering the matchmaking pool:

Set the player state to "NEW".

For players in state "NEW", 
	
Find the lowest latency datacenters for the player. If this datacenter is < 50ms then it is acceptable. Find any other datacenters within +10ms of the lowest latency datacenter and call these datacenters the set of "ideal datacenters".

* If one or more ideal datacenters exist, put the player into "IDEAL" state and add the player to the queue of each acceptable datacenter. Take the first match that comes back from ANY of the ideal datacenters.

* Players in "IDEAL" state should quickly find a game within 60 seconds and go to "PLAYING" state. 

* Any players in "IDEAL" state that don't find a game within 60 seconds, go to "WARMBODY" state.

If no datacenters are acceptable for the new player, then they possibly outside region, or at they are not near any ideal servers for them (eg. nothing < 50ms). Put these players into "WARMBODY" state

* This is a pincer. People close to server should generally be weighted heavily to find a match there, people far away from any server should be warm bodies to fill games within the region, with some preference towards filling games closer to them of course, but in an iterative fashion, such that they relax over time.

* Potentially, the players in the warmbody queue will iterate across all datacenters within their current (expanding) acceptable latency range, and look for any partial games they can join that let those games start.

* Players in "WARMBODY" state that do not find a match within n seconds, go into "BOTS" state.

* Players in "PLAYING" or "BOTS" state return to "NEW" state at the end of the match.
