# Directives in PML

Directives are special syntax blocks in PML that encapsulate operations related to LLMs and similar AI systems. They provide a structured way to interact with AI capabilities within your code while maintaining clear boundaries and control.

## The Problem Directives Solve

When working with LLMs and large codebases, one significant challenge is that LLMs tend to make widespread changes across the entire codebase. This can lead to:

1. Unintended side effects
2. Difficult-to-review changes
3. Complex debugging when things go wrong
4. Loss of control over specific parts of your code

Directives solve this by creating well-defined boundaries that:

- Restrict the scope of LLM changes to specific blocks
- Prevent modifications to other parts of the codebase
- Make AI interactions explicit and predictable
- Isolate AI-driven code from regular code

## Core Directive Types

### `:ask` - LLM Query Directive

The `:ask` directive sends a query to an LLM and captures the response:

```
:ask
What's the capital of France?
:--
```

When processed, the directive gets postfixed with a result:

```
:ask
What's the capital of France?
:--(happy_panda:"Paris")
```

#### Advanced `:ask` Usage

You can name `:ask` blocks to reference them later and specify return types:

```
:ask route_decision
What category does this query fall into: technical, general, or personal?
:return_type RouteDecision
:--
```

### `:do` - Action Execution Directive

The `:do` directive enables code execution or planning actions:

```
:do
echo "Hello from main.pml"
:--
```

In future implementations, `:do` will support more complex workflow control:

```
:do
when guardrail_check.accept == false -> raise BlockedInputError(guardrail_check.reply)
when route_decision.route == "image_analysis" -> ask image_analysis
otherwise -> ask nutrition_expert
:--
```

### `:context` - State Management Directive

The planned `:context` directive will handle state management across multiple directives:

```
:context user_profile
store profile
:--
```

### `:reflect` - Verification Directive

The planned `:reflect` directive will provide runtime validation and quality control:

```
:reflect throughout module
"Ensure all decisions conform to compliance rules and verify that cached results are correct."
:--
```

## Directives as Boundaries

A key benefit of directives is creating clear boundaries for LLM interactions. When an LLM sees a directive, it understands:

1. "I should only modify content within this block"
2. "The rest of the code should remain unchanged"
3. "My changes should respect the directive's specific purpose"

This boundary-setting is crucial for maintaining control when working with generative AI, as it prevents the "change everything" tendency that LLMs sometimes exhibit.

## Directive Syntax Rules

1. All directives start with a colon (`:`)
2. Directive blocks are terminated with `:--`
3. After processing, directives may be postfixed with results
4. Directives can be named for later reference
5. Some directives accept parameters like `:return_type`

## Implementation Status

- `:ask` - Fully implemented in current version
- `:do` - Basic implementation available, extended functionality planned
- `:context` - Planned for future release
- `:reflect` - Planned for future release

## Use Cases

Directives are particularly useful for:

1. AI-assisted code generation with controlled scope
2. Natural language processing within applications
3. Controlling complex AI workflows
4. Maintaining state across multiple AI interactions
5. Preventing unintended LLM modifications to critical code
6. Creating clear separation between AI and non-AI components

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
