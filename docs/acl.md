# GizClaw ACL Matrix

This file lists the target ACL-managed subjects, resources, permissions, and
business actions for Server Service and Admin Service.

The current ACL schema already supports `workspace`, `workflow`, `model`,
`credential`, `voice`, and `view`. The other resource kinds below are planned
business extensions and must be added to the ACL schema before implementation.

## Target ACL-Controlled Resource Kinds

```text
workspace
workflow
model
credential
voice
view
pet
wallet
contact
friend
friend_request
group
call
game_result
reward
```

## Subjects

| Subject kind | ID | Meaning |
| --- | --- | --- |
| `pk` | peer public key | One peer identity. |
| `view` | view name | A grouped subject for curated access. |
| `all_peers` | empty | Default subject that every connected peer can inherit. |

## Resource Matrix

| Resource kind | ID | Owner | Permissions | Server Service usage | Admin Service usage |
| --- | --- | --- | --- | --- | --- |
| `workspace` | workspace name | peer or admin | `workspace.{read,use,admin}` | `server.workspace.{list,get,create,put,delete}`, `server.run.reload` | `/workspaces/{name}` |
| `workflow` | workflow name | peer or admin | `workflow.{read,use,admin}` | `server.workflow.{list,get,create,put,delete}`, `server.run.reload` | `/workflows/{name}` |
| `model` | model id | peer or admin | `model.{read,use,admin}` | `server.model.{list,get,create,put,delete}`, `server.run.reload`, `server.run.say` | `/models/{id}` |
| `credential` | credential name | peer or admin | `credential.{read,use,admin}` | `server.credential.{list,get,create,put,delete}`, `server.run.reload`, `server.run.say` | `/credentials/{name}` |
| `voice` | voice id | admin | `voice.{read,use,admin}` | `server.run.say`, voice selection, and runtime use | `/voices/{id}` |
| `view` | view name | admin | `view.{read,use,admin}` | read/use resources exposed by a view | `/acl/views/{name}` |
| `pet` | pet id | peer | `pet.{read,use,admin}` | `server.pet.{list,get,create,put,delete}`, `server.pet.feed`, `server.pet.play`, `server.pet.level-up` | `/peers/{publicKey}/pets/{id}` |
| `wallet` | peer public key | peer | `wallet.{read,use,admin}` | `server.wallet.get`, `server.wallet.transactions.list` | `/peers/{publicKey}/wallet` |
| `contact` | contact id | peer | `contact.{read,use,admin}` | `server.contact.{list,get,create,put,delete}`, `server.contact.block`, `server.contact.unblock` | `/peers/{publicKey}/contacts/{id}` |
| `friend` | friend relation id | peer pair | `friend.{read,use,admin}` | `server.friend.{list,delete}` | `/peers/{publicKey}/friends/{id}` |
| `friend_request` | request id | peer pair | `friend_request.{read,use,admin}` | `server.friend.requests.{list,create}`, `server.friend.requests.accept`, `server.friend.requests.reject` | `/peers/{publicKey}/friend-requests/{id}` |
| `group` | group id | peer or admin | `group.{read,use,admin}` | `server.group.{list,get,create,put,delete}`, `server.group.members.{list,add,delete}` | `/groups/{id}` |
| `call` | call id | peer pair or group | `call.{read,use,admin}` | `server.call.{list,get,create}`, `server.call.answer`, `server.call.reject`, `server.call.end` | `/calls/{id}` |
| `game_result` | result id | peer | `game_result.{read,use,admin}` | `server.game.results.create` | `/game-results/{id}` |
| `reward` | reward id | peer or system | `reward.{read,use,admin}` | `server.reward.{list,get,create}`, `server.reward.claim` | `/rewards/{id}` |

## Permission Mapping

| Permission suffix | Meaning |
| --- | --- |
| `read` | List or get metadata/state for a resource. |
| `use` | Use the resource at runtime or perform normal owner actions. |
| `admin` | Create, update, delete, or manage ACL for the resource. |

## Runtime ACL Checks

| Operation | Required ACL checks |
| --- | --- |
| `server.run.agent.set` | target `workspace.use` |
| `server.run.reload` | current pending agent workspace `workspace.use`, `workflow.use`, referenced `model.use`, referenced `credential.use` |
| `server.run.say` | selected `voice.use`, selected TTS `model.use`, referenced `credential.use` |
| `server.group.messages.list` | `group.read` |
| `server.group.messages.send` | `group.use` |
| reward settlement into wallet | `reward.use`, target `wallet.use` |

## Default Ownership Rules

| Create path | Subject to bind | Resource to bind | Role |
| --- | --- | --- | --- |
| Peer creates workspace | `pk:{peerPublicKey}` | `workspace:{name}` | workspace owner/admin |
| Peer creates workflow | `pk:{peerPublicKey}` | `workflow:{name}` | workflow owner/admin |
| Peer creates model | `pk:{peerPublicKey}` | `model:{id}` | model owner/admin |
| Peer creates credential | `pk:{peerPublicKey}` | `credential:{name}` | credential owner/admin |
| Peer creates pet | `pk:{peerPublicKey}` | `pet:{id}` | pet owner/admin |
| Peer starts wallet | `pk:{peerPublicKey}` | `wallet:{peerPublicKey}` | wallet owner/admin |
| Peer creates contact | `pk:{peerPublicKey}` | `contact:{id}` | contact owner/admin |
| Peer creates group | `pk:{peerPublicKey}` | `group:{id}` | group owner/admin |
| Peer creates game result | `pk:{peerPublicKey}` | `game_result:{id}` | game result owner/admin |
| Peer creates reward | `pk:{peerPublicKey}` | `reward:{id}` | reward owner/admin |
| System creates reward for peer | `pk:{peerPublicKey}` | `reward:{id}` | reward owner/user |

## Shared Resource Rules

| Shared resource | Subject | Resource | Role |
| --- | --- | --- | --- |
| Built-in model for everyone | `all_peers` | `model:{id}` | model reader/user |
| Built-in model for a group | `view:{name}` | `model:{id}` | model reader/user |
| Shared credential for one peer | `pk:{peerPublicKey}` | `credential:{name}` | credential user |
| Shared credential for a group | `view:{name}` | `credential:{name}` | credential user |
| Shared voice for everyone | `all_peers` | `voice:{id}` | voice reader/user |
