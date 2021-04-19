# Mempool DipDup
DipDup for mempool indexing

## Config

Fully compatible config with default DipDup YAML config. To index mempool you should add section `mempool` in your config. For example:

``` yaml
version: 0.0.1

datasources:
  tzkt_mainnet:
    kind: tzkt
    url: https://api.tzkt.io
  node_mainnet:
    kind: tezos-node
    url: https://rpc.tzkt.io/mainnet

contracts:
  myaccount:
    address: tz1...

mempool:
  settings:
    keep_operations_seconds: 172800
    expired_after_blocks: 60
    keep_in_chain_blocks: 10
    mempool_request_interval_seconds: 10
    rpc_timeout_seconds: 10
  indexers:
    mainnet:
      filters:
        kinds:
          - transaction
        accounts:
          - myaccount
      datasources:
          tzkt: tzkt_mainnet
          rpc: 
            - node_mainnet

database:
  kind: sqlite
  path: mempool.db
```

### Config in details

Section `mempool` contains two keys: `settings` and `indexers`. `settings` part is general settings for the dipdup service. `indexers` is set of used indexers.

#### `settings` section

* `keep_operations_seconds` - how long store operations in seconds. When operations store more than the `keep_operations_seconds` period it will be removed from database. Exclude `in_chain` status. For `in_chain` operations use parameter `keep_in_chain_blocks`. Default: 172800 (2 days).


* `expired_after_blocks` - after blocks count pointed in `expired_after_blocks` paramater operation's status will be set to `expired`. Default: 60.

* `keep_in_chain_blocks` - how long store operations with status `in_chain` in blocks. Default: 10.

* `mempool_request_interval_seconds` - request interval to tezos node for receiving mempoool operations in seconds. Default: 10.

* `rpc_timeout_seconds` - timeout for request to tezos node. Default: 10.

#### `indexers` section

This section is used to create custom mempool indexers. You can create 1 indexer per network. Network is the primary key of indexer. For example:

```yaml
 indexers:
    mainnet:
        ...
    edonet:
        ...
    delphinet:
        ...
```

Every indexer has 2 settings: `filters` and `datasources`.

`filters` is the your filtration rules.

* `kinds` - array of mempool operation's kinds. 
It may be any of `activate_account`, `ballot`, `delegation`,  `double_baking_evidence`,  `double_endorsement_evidence`, `endorsement`, `origination`, `proposal`, `reveal`, `seed_nonce_revelation`, `transaction`
Default: `transaction`.

* `accounts` - array of tezos tz and KT addresses which will be used for filtering by source, destination and etc.

`datasources` is section for setting URLs of tezos nodes and TzKT.

* `tzkt` - TzKT url

* `rpc` - array of tezos nodes URL.