# web3-rpc-monitor
Updates a graph every 2 seconds to show a comparison across 5 RPC endpoints the difference in max block number

![rpc monitor](./rpc_monitor.png)

## go-socket
availble via wscat
```
$ npx wscat -c ws://137.220.54.108:5000/live
< {"blocks":[44711461,44711461,44711462,44711461,44711461],"max":44711462,"time":"2023-07-05T08:08:21.975518733-04:00"}

$ npx wscat -c ws://137.220.54.108:5000/counts
< {"counts":[4,2,0,10,6],"duration":10}
```

## go-rest

#### http://137.220.54.108:8000/api/rpc/latest
return one of the following strings based on latest block: {SEQUENCE_RPC|ALCHEMY_RPC|QUICKNODE_RPC|POLYGON_RPC|ANKR_RPC}
```
{"provider":"ALCHEMY_RPC"}
```

#### http://137.220.54.108:8000/api/1hr
return 1 hr of data from across 5 different providers
```
{"blocks":{"0":[...], "1": [...], ... , "time":[...]}
```

#### http://137.220.54.108:8000/api/notify/:number
add a phone number to the list of numbers to text when Sequence is behind the threshold
```
$ curl http://localhost:8000/api/notify/+16475555555
```

