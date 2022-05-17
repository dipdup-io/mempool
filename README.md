# Mempool indexer

[![Tests](https://github.com/dipdup-net/mempool/workflows/Tests/badge.svg?)](https://github.com/dipdup-net/mempool/actions?query=workflow%3ATests)
[![Docker images](https://github.com/dipdup-net/mempool/workflows/Release/badge.svg?)](https://hub.docker.com/r/dipdup/mempool)
[![Made With](https://img.shields.io/badge/made%20with-dipdup-blue.svg?)](https://dipdup.net)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Selective Tezos mempool indexer based on DipDup framework.

## Configuration

Fully compatible with DipDup YAML configuration file format.  
Mempool indexer reuses `datasources`, `contracts`, `database`, `hasura` sections, and reads its own settings from `mempool` top-level section.  

Read more [in the docs](https://docs.dipdup.net/config-file-reference/plugins/mempool).

## GQL Client

```
npm i @dipdup/mempool
```

Read [how to use](./build/client/README.md) the GraphQL client for the Mempool service.