# Server Resource RPC

## Story

A registered peer can manage workspace, workflow, model, and credential resources through the Server Service RPC surface when ACL bindings explicitly allow those resources. A different peer without matching bindings must be denied.

## Coverage

- Admin seeds existing resources before peer RPC reads them.
- Peer RPC lists and gets seeded resources.
- Peer RPC creates, updates, gets, lists, and deletes its own allowed resources.
- ACL denies access for an unbound peer.
