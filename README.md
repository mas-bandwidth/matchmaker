# matchmaker simulation

Example source code for https://mas-bandwidth.com/creating-a-matchmaker-for-your-multiplayer-game

To run the simulator:

```console
make && ./dist/matchmaker
```

You should see output like this:

```console
2024-02-18 04:57:15:      25312 players    2s average search time    33ms average latency
2024-02-18 04:57:16:      25288 players    2s average search time    33ms average latency
2024-02-18 04:57:17:      25308 players    2s average search time    32ms average latency
2024-02-18 04:57:18:      25300 players    2s average search time    32ms average latency
2024-02-18 04:57:19:      25308 players    2s average search time    33ms average latency
2024-02-18 04:57:20:      25308 players    2s average search time    33ms average latency
2024-02-18 04:57:21:      25296 players    2s average search time    33ms average latency
2024-02-18 04:57:22:      25284 players    2s average search time    33ms average latency
2024-02-18 04:57:23:      25296 players    2s average search time    33ms average latency
2024-02-18 04:57:24:      25288 players    2s average search time    32ms average latency
2024-02-18 04:57:25:      25276 players    2s average search time    33ms average latency
2024-02-18 04:57:26:      25268 players    2s average search time    32ms average latency
2024-02-18 04:57:27:      25256 players    2s average search time    32ms average latency
2024-02-18 04:57:28:      25260 players    2s average search time    32ms average latency
```

To view the real-time visualization of players on a map, just open map/index.html in a browser.

Or you could just watch it on YouTube:

<a href="http://www.youtube.com/watch?v=5QOyvrKB_8Q">
  <img width="2467" alt="image" src="https://github.com/mas-bandwidth/matchmaker/assets/696656/c222d80e-3706-4e87-9ec6-12b563ee57cd">>
</a>

[![Matchmaker Simulator on YouTube]([https://github.com/mas-bandwidth/matchmaker/assets/696656/0cf93c31-422b-47da-b27c-9ad30896ac50](https://github.com/mas-bandwidth/matchmaker/assets/696656/0cf93c31-422b-47da-b27c-9ad30896ac50))]( "Matchmaker Simulation")

The datasets are included under the "data" folder. In particular, players.csv defines the player lat/long coordinates joining each second. 

The rest of the data defines the set of datacenters and the latency maps per-datacenter.

Tested on MacOS. Linux should work. Windows untested.
