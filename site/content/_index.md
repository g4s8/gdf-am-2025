+++
title = "Go Beyond Models"
outputs = ["Reveal"]
+++

# Go Beyond Models

Why AI needs a fast backend — and why Go fits so well

{{% note %}}
Hi everyone, my name is Kirill. I work mostly on backend systems and infrastructure, mainly using Go.
This talk is not about how AI models work, and it’s not a Go tutorial.
Instead, I want to talk about the part of AI systems that usually gets less attention — the backend infrastructure that makes all this work in production.
{{% /note %}}

---

# The AI hype

 * GPT
 * Agents
 * RAG
 * Tools

{{% note %}}
When we talk about AI today, we usually talk about models.
GPT, Claude, agents, tools, prompts - that's where most discussions stop.

But models don't magically run in production.

Every real AI system depends on a backend that handles traffic, failures, latency, and cost.
And that backend is where most real engineering problems live.
{{% /note %}}

---

# The invisible layer

{{< mermaid >}}
graph LR
    A["Frontend"]
    A --> B["Backend"]
    B --> C["LLM API"]
{{< /mermaid >}}

Backend responsibilities:
 * streaming
 * concurrency
 * retries
 * limits
 * costs

{{% note %}}
Between users and AI models, there is always a backend layer.
This layer does things like authentication, streaming responses, retries, rate limits, logging, and metrics.
If this layer is slow or unstable, users don’t care how smart your model is — the system feels broken.
This talk is about that layer.
{{% /note %}}

---

# Real production problems

 * {{% fragment %}} Users expect streaming {{% /fragment %}}
 * {{% fragment %}} APIs are expensive {{% /fragment %}}
 * {{% fragment %}} Requests must be cancellable {{% /fragment %}}
 * {{% fragment %}} Load spikes are normal {{% /fragment %}}
 * {{% fragment %}} Latency matters {{% /fragment %}}

{{% note %}}
In production, AI backends have very specific problems.

Users expect streaming responses.
APIs are expensive.
Load is bursty.
And latency really matters.

These are not ML problems — these are classic backend engineering problems, just amplified by AI.
{{% /note %}}

---

{{% section %}}

# AI gateway

{{< mermaid >}}
graph LR
    A(["Clients"])
    B["Load balancer"]
    C(["AI API"])
    A --> B
    B --> X["API backend"]
    B --> Y["API backend"]
    B --> Z["API backend"]
    X --> C
    Y --> C
    Z --> C
{{< /mermaid >}}

{{% note %}}
A common solution is to introduce an AI gateway — a single backend service that all clients talk to.

This is usually a small service — often 1–3 replicas — but it sits on the critical path.
{{% /note %}}

---

Responsibilities:
 * auth
 * streaming
 * retries
 * rate limits
 * metrics

{{% note %}}
This gateway handles authentication, streaming, retries, rate limits, metrics, and abstracts different AI providers.

From an architecture point of view, this is a normal backend service.
And that’s exactly why Go fits so well here.
{{% /note %}}

{{% /section %}}

---

# What this backend must handle

 * High concurrency
 * Streaming responses
 * Cancellations
 * Bounded load
 * Predictable latency

---

# Why Go

 * High concurrency → goroutines
 * Streaming → net/http
 * Cancellation → context
 * Predictability → stable memory & GC
 * Deploy → single binary

{{% note %}}
Go was designed for exactly this kind of workload.

You get cheap concurrency with goroutines.
Streaming with the standard HTTP library.
Cancellation built into the language.
Predictable memory usage.
And very simple deployment — a single binary.

None of this is exotic, but together it makes Go very practical.
{{% /note %}}

---


# concurrency

Typical request:
 * user request
 * retrieval
 * user context
 * tool calls
 * LLM

{{% note %}}
A single AI request is rarely just one call.

You might fetch user context, run a retrieval query, call some tools, and then call the LLM.

That means one user request often becomes several concurrent backend requests.

If your concurrency model is awkward, the whole system becomes hard to reason about.
{{% /note %}}

---

# Fan-in / Fan-out

{{< mermaid >}}
graph LR
    A(["Request"])
    E(["LLM request"])

    A --> B["Load data"]
    A --> C["Load user profile"]
    A --> D["Load context"]

    B --> E
    C --> E
    D --> E
{{< /mermaid >}}

---
{{% section %}}

# Streaming

Expectation:
 * token-by-token UX

{{% note %}}
From the user’s point of view, streaming looks simple: text appears token by token.
{{% /note %}}

---

Reality:
 * upstream → downstream
 * client disconnects
 * partial failures

{{% note %}}
From the backend point of view, it’s harder.
You stream from the AI provider, you stream to the client, clients disconnect, networks fail.

Streaming is easy to demo, but surprisingly hard to get right in production.
{{% /note %}}


{{% /section %}}

---

{{% section %}}

# Streaming in Go

```go
for msg := range stream {
    fmt.Fprint(w, msg.Text)
    flusher.Flush()
}
```

{{% note %}}
This is what streaming looks like in Go.
{{% /note %}}

---

 * no async/await
 * no framework
 * standard library
 * readable even for non-Go devs

{{% note %}}
There's no async framework, no callbacks, no magic.
You read from the upstream stream and write to the response.

Even if you don't know Go, this code is readable.
That's one of Go’s biggest strengths.
{{% /note %}}


---
 

 * cheap goroutines
 * natural parallelism
 * simple mental model

---

 * Python: async complexity
 * Node: event loop pressure
 * Java: thread cost

{{% /section %}}

---

# AI backends fail at p99

 * LLM APIs have unpredictable latency
 * One slow request blocks many others
 * Retries amplify the problem
 * Costs grow silently

{{% note %}}

Manager-friendly:
    When we talk about performance, average numbers lie.
    AI systems usually look fine on average — until users start complaining.

Engineer hook:
    LLM APIs are slow, variable, and rate-limited.
    That makes tail latency unavoidable unless you control it.
    It’s like traffic: the average commute is fine, but one accident blocks the whole city.

Key line:
    Most AI outages are p99 problems, not throughput problems.

{{% /note %}}

---

# Why AI makes p99 worse

 * External APIs (network + vendor)
 * Streaming responses
 * Fan-out to multiple services
 * Expensive retries
 * Bursty traffic

{{% note %}}
    Every AI request fans out.
    One user request becomes several backend calls.

    If you don’t put limits somewhere, the system doesn’t slow down — it collapses.
{{% /note %}}

---

# What happens without limits?

 * queues grow
 * latency explodes
 * retries amplify load
 * costs spike
 * system collapses

{{% note %}}
Most AI outages are slow failures, not crashes.
{{% /note %}}

---

# Backpressure keeps systems stable

{{< mermaid >}}
graph LR
    A(["Clients"])
    F("LLM API")

    A --> B["Bounded queue"]

    B --> C["Workers"]

    C --> F
{{< /mermaid >}}


---

# Backpressure keeps systems stable

 * Fixed queue size
 * Fixed concurrency
 * Fast rejection when full

{{% note %}}
Backpressure means we say ‘no’ early, instead of failing slowly.
Bounded queues turn overload into fast failure instead of cascading latency.
Fast rejection keeps p99 low.
{{% /note %}}

---

# Bounded channels

{{< mermaid >}}
graph LR
    A(["Clients"])
    F("LLM API")

    A --> B["Channel  (100)"]

    B --> C["Worker1"]
    B --> D["Worker ..."]
    B --> E["Worker20"]

    C --> F
    D --> F
    E --> F
{{< /mermaid >}}

Queue size = max in-flight requests.
Workers = max parallel calls

{{% note %}}
This channel is a physical limit.
It defines how much load the system is allowed to accept.

When the channel is full, we reject immediately.
That protects latency, memory, and cost.

We’d rather say 'try again' than let everyone wait forever.

No unbounded queues. No surprise latency spikes.
{{% /note %}}


---

{{% section %}}

**Why Go is good at explicit backpressure**

 * Channels are first-class
 * Bounded by default
 * Cheap concurrency
 * Cancellation is built-in

---

```go
select {
case queue <- req:
default:
    return http.StatusTooManyRequests
}
```

{{% note %}}
This pattern exists in many languages. Go just makes it obvious.

Python: async queues + cancellation = complexity
Node: event loop → shared fate
Java: thread pools work, but heavier and more config

Key line:
    In Go, backpressure is visible in the code.
{{% /note %}}

{{% /section %}}

---

**Cancellation = cost control**

User closes tab → Request context cancelled → Stop upstream token generation

{{% note %}}
(Manager angle) In AI systems, cancellation literally saves money.

(Engineer) context.Context propagates intent across the system.

(Tie back to p99) Cancellation prevents slow requests from poisoning the tail.
{{% /note %}}

---

{{% section %}}

# Go vs others

---

Python

 * good for models
 * painful concurrency

---

Node

 * async-friendly
 * memory & CPU under load

---

Java

 * mature
 * heavy for gateways

---

Go

 * simple
 * predictable
 * infra-native

{{% /section %}}

---

# Deployment & ops

  - {{% fragment %}} one binary {{% /fragment %}}
  - {{% fragment %}} small Docker image {{% /fragment %}}
  - {{% fragment %}} fast startup {{% /fragment %}}
  - {{% fragment %}} easy autoscaling {{% /fragment %}}
  - {{% fragment %}} built-in profiling {{% /fragment %}}

---

# Go for AI

 * {{% fragment %}} multiple clients {{% /fragment %}}
 * {{% fragment %}} streaming {{% /fragment %}}
 * {{% fragment %}} cost control {{% /fragment %}}
 * {{% fragment %}} reliability {{% /fragment %}}
 * {{% fragment %}} multi-provider {{% /fragment %}}

{{% note %}}
Go is a good choice when you have multiple AI consumers, streaming UX, strict cost control,
and reliability requirements.

You don’t need Go everywhere - just where predictability matters.
{{% /note %}}

---

AI models may be smurt. AI backends must be reliabile.

---

# Contacts

![qr-g4s8](images/qr-g4s8.png)

 - Me: [github.com/g4s8](https://github.com/g4s8)
 - Slides: [github.com/g4s8/gdf-am-2025](https://github.com/g4s8/gdf-am-2025)

