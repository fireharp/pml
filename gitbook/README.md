# PML

## Why PML?

* all llm/near llm calls -> wrap in special `:` directives [https://fireharp.gitbook.io/pml/directives](https://fireharp.gitbook.io/pml/directives)&#x20;
* make all code bit more interactive / easier to work with
* don't think much about agentic frameworks at all -> just write yor python code, and think about `:` directives as async func calls&#x20;
* under the hood PML becomes smarter as technology evolves – migrating to newer APIs and protocols, embracing most resent AI tech out there

## Quick start

```
:ask
what's the capital of France?
:--
```

will be rendered as

```
:ask
what's the capital of France?
:--(happy_panda:"Paris")
```

