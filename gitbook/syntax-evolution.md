# PML Syntax Evolution

This document outlines the evolution of PML syntax and the rationale behind its design choices. PML's syntax is being continuously refined to better integrate with Python and provide a seamless way to work with LLMs.

## Current Syntax

The current implementation of PML supports the following basic syntax:

```
:ask
  What is the capital of France?
:--
```

Which after processing becomes:

```
:ask
  What is the capital of France?
:--(happy_panda:"Paris")
```

## Syntax Refinements

Based on feedback and practical usage, several refinements to the PML syntax are being explored:

### Named Directives with Return Types

Explicitly naming directives and specifying return types improves referencing and type safety:

```
:ask route_decision
  What category does this query fall into: technical, general, or personal?
:return_type RouteDecision
:--
```

### Advanced `:do` Directives

The `:do` directive is evolving from simple execution to a more declarative, DSL-like branching system:

```
:do
  when guardrail_check.accept == false -> raise BlockedInputError(guardrail_check.reply)
  when route_decision.route == "image_analysis" -> ask image_analysis
  otherwise -> ask nutrition_expert
:--
```

This avoids confusion with actual Python execution and provides clearer workflow control.

### Context Management

A new `:context` directive is being developed to handle state management:

```
:context user_profile
store profile
:--
```

This makes it explicit that the context is available for later calls.

### Reflection and Quality Control

The `:reflect` directive is being designed as a parallel quality control mechanism:

```
:reflect throughout module
  "Ensure all decisions conform to compliance rules and verify that cached results are correct."
:--
```

This allows reflection to occur dynamically during execution rather than as a final step.

## Python Integration

Since Python is the main execution context for many PML applications, newer syntax designs emphasize better Python integration:

1. Removing redundant type quoting (using `RouteDecision` instead of `"RouteDecision"`)
2. Supporting async functionality explicitly:
   ```
   :ask guardrail_check async
   ```
3. Better handling of binary and structured data:
   ```
   :ask image_block
   input:
     image: base64_image_data
     user_message: "{user_message}"
   ```

## Future Direction

The PML syntax is evolving towards:

1. **Clarity**: Making the syntax more intuitive and self-documenting
2. **Integration**: Better integration with Python and other languages
3. **Expressiveness**: Supporting more complex AI interactions and workflows
4. **Type Safety**: Stronger typing for improved reliability
5. **Modularity**: Better support for reusable components and patterns

These refinements aim to make PML a more powerful and intuitive tool for AI-assisted programming.
