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

<iframe width="800" height="350" src="https://www.youtube.com/embed/5QOyvrKB_8Q?si=L0gHfcLv2Lz-shUA" title="Matchmaker Simulation" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>

The datasets are included under the "data" folder. In particular, players.csv defines the player lat/long coordinates joining each second. 

The rest of the data defines the set of datacenters and the latency maps per-datacenter.

Tested on MacOS. Linux should work. Windows untested.
