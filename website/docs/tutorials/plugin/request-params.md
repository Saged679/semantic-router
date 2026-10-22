# Request Parameters

## Overview

`request_params` is a route-local plugin that validates and trims OpenAI Chat Completions request bodies before they are forwarded to backends.

## Key Advantages

- Caps expensive parameters (`max_tokens`, `n`) per route.
- Blocks request parameters such as `verbosity` for tiers that must not carry them.
- Optionally strips unknown top-level JSON fields to reduce surprise passthrough behavior.
- Can remove a client `prompt_cache_key` with `blocked_params: [prompt_cache_key]` before backend dispatch.

## What Problem Does It Solve?

Model routing can restrict which backend serves a request, but clients can still amplify cost or extract logits via request parameters. This plugin enforces per-decision limits on the request body after a decision matches.

## When to Use

- a tier or route must not carry a request parameter the backend should not see
- you need a hard ceiling on `max_tokens` or `n` independent of client input
- unknown JSON fields should not be forwarded to backends

## Configuration

Add the plugin under `routing.decisions[].plugins`:

```yaml
plugins:
  - type: request_params
    configuration:
      blocked_params:
        - verbosity
      max_tokens_limit: 500
      max_n: 1
      strip_unknown: true
```

In DSL, the same plugin can appear as:

```dsl
PLUGIN request_params {
  blocked_params: ["verbosity"]
  max_tokens_limit: 500
  max_n: 1
  strip_unknown: true
}
```

A `blocked_params` entry names a request field the Router forwards, in the name
the client sends. A name the Router cannot block is refused when the
configuration loads, so a policy never looks applied while the parameter still
reaches the backend. Fields the Router refuses outright, such as `logprobs`,
never reach this plugin: the codec rejects a request that carries them.

This plugin enforces a bounded set of OpenAI Chat Completions fields. It is not
a general JSON-schema firewall and does not authorize a caller. Test
`strip_unknown` against clients that add provider-specific fields before
enabling it. See a complete example:
[`config/fragments/plugin/request-params/budget-tier.yaml`](https://github.com/vllm-project/semantic-router/blob/main/config/fragments/plugin/request-params/budget-tier.yaml).
