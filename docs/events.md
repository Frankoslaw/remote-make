# Master Events

## node.UUID.task.schedule
Body: `task_template_uuid`

Triggers:
- if `is_local` is set to true
    - `node.UUID.task.UUID.start`
- if `is_local` is set to false
    - `task.UUID.provision`

Subscribes:
- `task.UUID.provision` 
- `task.UUID.done`
- `task.UUID.error`

## task.UUID.provision
Body: `task_uuid`

Triggers:
- `task.UUID.ready` (indirectly when node joins cluster it sends this event)

Subscribes:
- `node.UUID.ready`
- `node.UUID.terminate`
- `node.UUID.terminated`
- `node.UUID.interrupted`

## task.UUID.error
Body: `err`

Triggers:
- `node.UUID.terminate` if workers exist and `is_local` is set to false
- `task.UUID.step.0.error` if parent exists

## task.UUID.done
Body: `task_uuid`

Triggers:
- `node.UUID.terminate` if workers exist and `is_local` is set to false
- `task.UUID.step.0.done` if parent exists

## node.UUID.ready
Body: `node_uuid`

Triggers:
- `node.UUID.task.UUID.start`

TODO:
- send machine info for usage in steps and special runtimes

# Worker Events

## node.UUID.task.UUID.start
Body: `task_uuid`

Triggers:
- `task.UUID.step.0.start`
- `task.UUID.done`

Subscribers:
- `task.UUID.step.0.done`
- `task.UUID.step.0.error`

## task.UUID.step.0.start
Body: `step_uuid`

Triggers:
- `task.UUID.step.0.done`
- or `task.UUID.step.0.error`

## task.UUID.step.0.done
Body: `step_uuid`

## task.UUID.step.0.error
Body: `err`

Triggers:
- `task.UUID.error`