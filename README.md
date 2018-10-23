# tm-tools

集成多个子命令功能，提供数据升级、数据查看功能。

支持的子命令：

* **migrate**: tendermint data from `v0.18.0` to `v0.23.1`
* **view**: view tendermint data in `blockstore.db` or `state.db`
* **recover**: recover tendermint block height

## tm_tools migrate
```
Usage: tm_tools migrate -old tmroot -new tmroot [-h startHeight]

	-old tmroot: dir of old tendermint root
	-new tmroot: dir of new tendermint root to store converted data
	-s startHeight: from which height to convert tendermint data, default is `1`
```

## tm_tools view
```
Usage: $ tm_tools view -db /path/of/db [-a get|getall|block] [-q key] [-d] [-v] [-h height]

    -db : db，Note: the db path cannot end with "/"
    [-a get|getall|block]： read the value of a key | output all keyes | read block info
    [-q key] ：key format, please see following "Tendermint data" section
    [-d]: whether decode value，default is "false"
    [-v new|old] ：new(0.23.1), old(0.18.0), default is "new"
    [-h height]: block height，workes with "-a block" arg to read block info at height "N"

examples：
$ tm_tools view -db /path/of/blockstore.db -a getall 
$ tm_tools view -db /path/of/blockstore.db -a block -t 1 -d 
$ tm_tools view -db /path/of/blockstore.db -q "H:1" -d
$ tm_tools view -db /path/of/state.db -q "stateKey" -d -v 
```

##### view 中参数key的说明
* "blockStore": blockchain height info. {"Height":32}
* "H:1"   ... "H:32": block meta info, 
* "P:1:0" ... "P:32:0": block part info. Block may be sliced to several parts, for each part, the key is "P:{height}:{partIndex}", partIndex start from `0`.
* "SC:1"  ... "SC:32": block seen commit info.
* "C:0"   ... "C:31": block commit info.

| key format            | value type                    | examples                    | 
| --------------------- | ----------------------------- | --------------------------- |
| `stateKey`            | raw byte of state             |                             | 
| `abciResponsesKey`    | raw byte of ABCI Responses    |                             | 
| `blockStore`          | raw json                      | "blockStore": {"Height":32} | 
| `H:{height}`          | raw byte of block meta        | H:1                         |
| `P:{height}:{index}`  | raw byte of block part        | P:1:0, P:32:0, P:32:1       |
| `SC:{height}`         | raw byte of block seen commit | SC:1, SC:32                 | 
| `C:{height-1}`        | raw byte of block commit      | C:0, SC:31                  | 


## tm_tools recover
```
tm_tools recover --db /home/share/chaindata/peer4/tendermint --h 100
```

## 新版本数据目录
```
root@mint:/home/share/chaindata# tree -L 3 peer1
peer1
├── ethermint
│   ├── chaindata
│   │   ├── 000009.log
│   │   ├── 000011.ldb
│   │   ├── CURRENT
│   │   ├── LOCK
│   │   ├── LOG
│   │   └── MANIFEST-000010
│   ├── LOCK
│   ├── nodekey
│   └── transactions.rlp
└── tendermint
    ├── addr_book.json
    ├── config
    │   ├── config.toml
    │   ├── genesis.json
    │   ├── node_key.json
    │   └── priv_validator.json
    └── data
        ├── blockstore.db
        ├── cs.wal
        ├── state.db
```


## 代码说明
`tm-tools/older` 存放的是老版本的数据结构及相应的转换方法
