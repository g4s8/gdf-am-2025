+++
title = "Go Beyond Models"
outputs = ["Reveal"]
+++

# Go Beyond Models

Why AI needs a fast backend — and why Go fits so well

---

# The AI hype layer

 * GPT
 * Agents
 * RAG
 * Tools

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

---

# Real production problems

 * {{% fragment %}} Users expect streaming {{% /fragment %}}
 * {{% fragment %}} APIs are expensive {{% /fragment %}}
 * {{% fragment %}} Requests must be cancelled {{% /fragment %}}
 * {{% fragment %}} Load spikes are normal {{% /fragment %}}
 * {{% fragment %}} Latency matters {{% /fragment %}}

---

# AI Gateway

{{% section %}}

{{< mermaid >}}
graph LR
    A(["Clients"])
    A --> B["Go API backend"]
    B --> C("OpenAI")
    B --> D("Something else")
    B --> E("Something else 2")
{{< /mermaid >}}

---

Responsibilities:
 * auth & tenants
 * streaming
 * retries
 * rate limits
 * metrics

{{% /section %}}

---

# Why concurrency

Typical request:
 * user request
 * retrieval
 * user context
 * tool calls
 * LLM

---

# Why Go

 * High concurrency → goroutines
 * Streaming → net/http
 * Cancellation → context
 * Predictability → simple GC
 * Deploy → single binary

---


{{% section %}}

# Streaming

Expectation:
 * token-by-token UX

---

Reality:
 * upstream → downstream
 * client disconnects
 * partial failures

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

---

 * no async/await
 * no framework
 * standard library
 * readable even for non-Go devs

{{% /section %}}

---

# Cancellation = money

User closes tab → Context cancelled → Stop LLM response

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

 * cheap goroutines
 * natural parallelism
 * simple mental model

---

 * Python: async complexity
 * Node: event loop pressure
 * Java: thread cost

{{% /section %}}

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
