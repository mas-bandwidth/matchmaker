# matchmaker simulation

## Goal

Find low latency matches as quickly as possible.

## Strategy

For each new player entering the matchmaking pool:

Set the player state to "NEW".

For players in state "NEW", 
	
Find the set of datacenters for the player that are acceptable (eg. latency < X, and if multiple datacenters are < X, find the set of datacenters within within +Y ms of lowest latency datacenter). 

* If one or more acceptable datacenters exist, put the player into "IDEAL" state and add the player to the queue of each acceptable datacenter.

* Players in "IDEAL" state should quickly find a game within 60 seconds and go to "PLAYING" state. 

* Any players in "IDEAL" state that don't find a game within 60 seconds, go to "WARMBODY" state.

If no datacenters are acceptable for the new player, then they possibly outside region, or at least near any acceptable servers (eg. nothing < 50ms). Put these players into "WARMBODY" state.

* This is a pincer. People close to server should generally be weighted heavily to find a match there, people far away from any server should be warm bodies to fill games within the region, with some preference towards filling games closer to them of course, but in an iterative fashion, such that they relax over time.

* Potentially, the players in the warmbody queue will iterate across all datacenters within their current (expanding) acceptable latency range, and look for any partial games they can join that let those games start.

* Players in "WARMBODY" state that do not find a match within n seconds, go into "BOTS" state.

* Players in "PLAYING" or "BOTS" state return to "NEW" state at the end of the match.
