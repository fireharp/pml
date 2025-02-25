# : directives

### Mainly for LLM-related calls

Should encapsulate LLM and similar calls. Ones that we don't have clear expression in current programming languages.

That's almost everything we have on LLM APIs and a bit more.

It could be more complex with tool calling or even some agentic stuff in it.

### Interactive

Another sign there should be directive: need for compilation / execution.

E.g. when we&#x20;

```
:ask
What's capital of France?
:--
```

We expect `:--` to be postfixed with result link.

There could be potentially needed human interaction if we're unsure on our respose.

