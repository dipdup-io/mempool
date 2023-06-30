# Mempool indexer

[![Tests](https://github.com/dipdup-net/mempool/workflows/Tests/badge.svg?)](https://github.com/dipdup-net/mempool/actions?query=workflow%3ATests)
[![Docker images](https://github.com/dipdup-net/mempool/workflows/Release/badge.svg?)](https://hub.docker.com/r/dipdup/mempool)
[![Made With](https://img.shields.io/badge/made%20with-dipdup-blue.svg?)](https://dipdup.net)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Selective Tezos mempool indexer based on DipDup framework.

## Configuration

Fully compatible with DipDup YAML configuration file format.  
Mempool indexer reuses `datasources`, `contracts`, `database`, `hasura`
sections, and reads its own settings from `mempool` top-level section.  

Mempool configuration has two sections: settings and indexers **(required)**.

## Settings

This section is optional so are all the setting keys.

```yaml
mempool:
  settings:
    keep_operations_seconds: 172800
    expired_after_blocks: 60
    keep_in_chain_blocks: 10
    mempool_request_interval_seconds: 10
    rpc_timeout_seconds: 10
  indexers:
    ...
```

### keep_operations_seconds

How long to store operations that did not get into the chain. After that period,
such operations will be wiped from the database. Default value is **172800** **seconds**
(2 days).

### expired_after_blocks

When `level(head) - level(operation.branch) >= expired_after_blocks` and operation is
still on in chain it's marked as expired. Default value is **60 blocks** (~1 hour).

### keep_in_chain_blocks

Since the main purpose of this service is to index mempool operations (actually it's a
rolling index), all the operations that were included in the chain are removed from  the
database after specified period of time. Default value is **10 blocks** (~10 minutes).

### mempool_request_interval_seconds

How often Tezos nodes should be polled for pending mempool operations.
Default value is **10 seconds**.

### rpc_timeout_seconds

Tezos node request timeout. Default value is **10 seconds**.

## Indexers

You can index several networks at once, or index different nodes independently.
Indexer names are not standardized, but for clarity it's better to stick with
some meaningful keys:

```yaml
 mempool:
   settings:
     ...
   indexers:
     mainnet:
       filters:
         kinds:
           - transaction
         accounts:
           - contract_alias
       datasources:
         tzkt: tzkt_mainnet
         rpc: 
           - node_mainnet
     edonet:
     florencenet: 
```

Each indexer object has two keys: `filters` and `datasources` (required).

### Filters

An optional section specifying which mempool operations should be indexed.
By default, all transactions will be indexed.

#### kinds

Array of operations kinds, default value is `transaction` (single item).  
The complete list of values allowed:

* `activate_account`
* `ballot`
* `delegation*`
* `double_baking_evidence`
* `double_endorsement_evidence`
* `endorsement`
* `origination*`
* `proposal`
* `reveal*`
* `seed_nonce_revelation`
* `transaction*`

`*`  — manager operations.

#### accounts

Array of [contract](../config/contracts.md) aliases used to filter operations by source or destination.  
**NOTE**: applied to manager operations only.

### Datasources

Mempool plugin is tightly coupled with [TzKT](../config/datasources.md#tzkt) and [Tezos node](../config/datasources.md#tezos-node) providers.

#### tzkt

An alias pointing to a [datasource](../config/datasources.md) of kind `tzkt` is expected.

#### rpc

An array of aliases pointing to [datasources](../config/datasources.md) of kind `tezos-node`  
Polling multiple nodes allows to detect more refused operations and makes indexing more robust in general.

Read more [in the docs](https://docs.dipdup.net/config-file-reference/plugins/mempool).

## GQL Client

```
npm i @dipdup/mempool
```

Read [how to use](./build/client/README.md) the GraphQL client for the Mempool service.