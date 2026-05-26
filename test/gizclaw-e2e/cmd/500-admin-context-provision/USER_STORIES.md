# 500 Admin Context Provision

## User Story

As an operator, I want a saved context to be provisioned through peer preparation
and then verified with an admin command, so I know the context
can be used for later control-plane workflows.

## Scenario

1. Start a real server with automatic peer creation enabled.
2. Create a saved CLI context pointing at that server.
3. Prepare the context through the test harness.
4. Run `gizclaw admin peers list --context admin-a`.
5. Verify the admin command can connect and succeeds after provisioning.

## Covered Behaviors

- provisioning a context through peer preparation enables admin command
  access.
- the scenario uses the harness API preparation path instead of the removed
  `play register` CLI.
