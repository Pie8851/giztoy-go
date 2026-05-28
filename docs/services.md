# GizClaw Service Tree

This document describes business-level services, not transport service IDs.

```text
Peer Service
└── peer
    └── ping

Public Service
├── server.info
│   └── get
├── GET /server-info
└── POST /login

Gear Service
├── device.info
│   └── get
├── device.identifiers
│   └── get
├── peer.info
│   ├── get
│   └── put
├── peer.runtime
│   └── get
├── workflow
│   └── list, get, create, put, delete
├── model
│   └── list, get, create, put, delete
└── credential
    └── list, get, create, put, delete

Admin Service
├── /@apply POST
├── /resources/{kind}/{name} GET, PUT, DELETE
├── /acl/views GET, POST
├── /acl/views/{name} GET, PUT, DELETE
├── /acl/roles GET, POST
├── /acl/roles/{name} GET, PUT, DELETE
├── /acl/policy-bindings GET, POST
├── /acl/policy-bindings/{id} GET, PUT, DELETE
├── /workflows GET, POST
├── /workflows/{name} GET, PUT, DELETE
├── /firmwares GET, POST
├── /firmwares/{name} GET, PUT, DELETE
├── /firmwares/{name}/@release POST
├── /firmwares/{name}/@rollback POST
├── /credentials GET, POST
├── /credentials/{name} GET, PUT, DELETE
├── /models GET, POST
├── /models/{id} GET, PUT, DELETE
├── /dashscope-tenants GET, POST
├── /dashscope-tenants/{name} GET, PUT, DELETE
├── /gemini-tenants GET, POST
├── /gemini-tenants/{name} GET, PUT, DELETE
├── /openai-tenants GET, POST
├── /openai-tenants/{name} GET, PUT, DELETE
├── /minimax-tenants GET, POST
├── /minimax-tenants/{name} GET, PUT, DELETE
├── /minimax-tenants/{name}/@sync-voices POST
├── /volc-tenants GET, POST
├── /volc-tenants/{name} GET, PUT, DELETE
├── /volc-tenants/{name}/@sync-voices POST
├── /voices GET, POST
├── /voices/{id} GET, PUT, DELETE
├── /workspaces GET, POST
├── /workspaces/{name} GET, PUT, DELETE
├── /peers GET
├── /peers/{publicKey} GET, DELETE
├── /peers/{publicKey}/@approve POST
├── /peers/{publicKey}/@block POST
├── /peers/{publicKey}/@refresh POST
├── /peers/{publicKey}/info GET, PUT
├── /peers/{publicKey}/config GET, PUT
├── /peers/{publicKey}/runtime GET
├── /peers/sn/{sn} GET
└── /peers/imei/{tac}/{serial} GET
```
