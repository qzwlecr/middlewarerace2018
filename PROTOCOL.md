## PROTOCOL

> \- You can't burn the sea.
>
> \- But we may boom the earth.

### Assumptions

This is the protocol used in inter-agent communications.

The primary workload in provider VM should be mainly the *server process*, so this protocol should require a minimal processing cost.

But there's another problem: the *Service Discovery* procedure only results in a list of available providers and their addresses, without the maximum possible workload of each provider.

This introduces a difficulty in load balancing: how could us know whether a provider is reaching its maximum capacity or not?

In this protocol, we will rely on the *Provider-side Agent* to figure out this for us. So this protocol will be acting more like a **Congestion Control Protocol** instead of simply transfer consumer queries between agents.

### Overview

Instead of requiring extra packets exchanging process besides the normal *Query-Reply* model, this protocol will have extra informations sent in these packets.

That's saying, for each *consumer query*, there will be **only one reply** from the server. There will be no *heartbeats*, *control signals*, every thing that is required for this protocol to work correctly is included in the queries and replies.

But that's not saying that you will only need to send once for a query from consumer. Under certain circumstances, consumer agent will have to rearrange packets.

> Note: this approach may not be the fastest way, and it's **highly experimental**. If anyone has better idea, discuss that before we run out of time.

### The Query

This is simple: since the consumer query have only 4 parameters, and each of them seem to be readable strings, so...

```text
Query Protocol:
POST[interface] + '\n' + POST[method] + '\n' + POST[pmTypeStr] + '\n' + POST[params] + '\n' \
+ POST[attachments]
```

very simple, right?

The *attachments* is for the **universal** requirement of the competition. In all test cases, this field is an empty *Map<string,string>*, serialized by *JSON*.

And the HTTP requests sent by consumers will not send this. You( I mean, the one who wrote *consumer-side agent* ) need to make an empty Map, serialize it, and set the corresponding mapping in the *HttpPacks* structure.

### The Response(reply)

For incoming queries, the provider-side agent should maintain a queue for these queries.

A certain amount of queries will be send concurrently to the provider process. In order to find out that **amount**, we will start from a small value, and increase it gradually until significant lags happened between a query and its corresponding response.

```C
// Normal Response:
struct NResp{
    unsigned long latency; // the time between the query enqueued in agent and response are finally sent by provider process, in us
    VAR_LENGTH response_content; // the response content by provider process.
};
```

In this case, we said this provider has reached its capacity.

When this happens, the provider agent can still enqueue some queries from the consumer agent, but the amount should **not** exceed the capacity of the provider process, and **no more** queries should be sent to the provider process until one of the processing queries returns.

When the provider agent have its queue reached its capacity, any further query send by consumer agent will result in a *rejected response*:

```C
// Rejected Response
struct RResp{
    unsigned long magic_number; // 'rejected', literally
};
```

In this case, the consumer agent are forced to rearrange the query, that is, send it to another provider.