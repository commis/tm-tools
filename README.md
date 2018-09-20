# tm-tools

集成多个子命令功能，提供数据升级、数据查看功能。

支持的子命令：

* **migrate**: tendermint data from `v0.18.0` to `v0.23.1`
* **view**: view tendermint data in `blockstore.db` or `state.db`
* **reset**: reset tendermint data block height

## tm_tools migrate
```
Usage: tm_migrator -old tmroot -new tmroot -priv priv_dir [-s startHeight]

	-old tmroot: dir of old tendermint root
	-new tmroot: dir of new tendermint root to store converted data
	-priv priv_dir: dir to place other validators's old `priv_validator.json`
	-s startHeight: from which height to convert tendermint data, default is `1`
```

## tm_tools view
```
Usage: $ tm_view -db /path/of/db [-a get|getall|block] [-q key] [-d] [-v new|old] [-t height]

// -db : db，Note: the db path cannot end with "/"
// [-a get|getall|block]： read the value of a key | output all keyes | read block info
// [-q key] ：key format, please see following "Tendermint data" section
// [-d]: whether decode value，default is "false"
// [-v new|old] ：new(0.23.1), old(0.18.0), default is "new"
// [-t height]: block height，workes with "-a block" arg to read block info at height "N"

examples：
$ tm_view -db /path/of/blockstore.db -a getall 
$ tm_view -db /path/of/blockstore.db -a block -t 1 -d 
$ tm_view -db /path/of/blockstore.db -q "H:1" -d -v old 
$ tm_view -db /path/of/state.db -q "stateKey" -d -v old 
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


## tm_tools reset
```
Allen@MacBook-Pro:~ ls -l ~/.tendermint.v0.23.1/data/
drwxr-xr-x  8 Allen  staff  272 Oct 15 20:23 blockstore.db
drwx------  3 Allen  staff  102 Oct 15 20:23 cs.wal
drwx------  3 Allen  staff  102 Oct 15 20:23 mempool.wal
drwxr-xr-x  8 Allen  staff  272 Oct 15 20:23 state.db
drwxr-xr-x  7 Allen  staff  238 Oct 15 20:23 tx_index.db
```

## 新版本数据目录
```
Allen@MacBook-Pro:~ ls -l ~/.tendermint.v0.23.1/data/
drwxr-xr-x  8 Allen  staff  272 Oct 15 20:23 blockstore.db
drwx------  3 Allen  staff  102 Oct 15 20:23 cs.wal
drwx------  3 Allen  staff  102 Oct 15 20:23 mempool.wal
drwxr-xr-x  8 Allen  staff  272 Oct 15 20:23 state.db
drwxr-xr-x  7 Allen  staff  238 Oct 15 20:23 tx_index.db
```


## 代码说明
`tm-tools/types` 存放的是老版本的数据结构
