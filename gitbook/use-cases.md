# PML Use Cases

This document outlines practical applications and use cases for PML in various domains.

## AI-Assisted Development

### Code Generation

PML enables developers to get AI assistance with code generation:

```
:ask
Write a Python function to calculate the Fibonacci sequence iteratively.
:--
```

The LLM responds with a well-implemented function that can be directly integrated into your codebase.

### Code Refactoring

Use PML to get suggestions on how to improve your existing code:

```
:ask
How can I refactor this code to be more efficient?

def slow_function(n):
    result = []
    for i in range(n):
        result.append(i**2)
    return result
:--
```

### Documentation Generation

Generate documentation for your code:

```
:ask
Generate docstring for this function:

def process_image(image_path, target_size=(224, 224), normalize=True):
    img = Image.open(image_path)
    img = img.resize(target_size)
    img_array = np.array(img)
    if normalize:
        img_array = img_array / 255.0
    return img_array
:--
```

## Data Analysis Workflows

### Data Exploration Guidance

```
:ask
What exploratory data analysis techniques should I use for this time series data with seasonal patterns?
:--
```

### Query Generation

```
:ask
Generate a SQL query to find customers who purchased more than 5 products in the last 30 days.
:--
```

## Complex Decision Workflows

This is where the future syntax with named directives and return types shines:

```
:ask user_intent
What is the user trying to do with this query: "I want to analyze my sales data from last quarter"?
:return_type UserIntent
:--

:ask tool_selection
Based on the user intent to analyze sales data, what tools should I recommend?
:return_type ToolRecommendation
:--

:do
when tool_selection.primary_tool == "tableau" -> show_tableau_templates
when tool_selection.primary_tool == "python" -> show_python_notebooks
otherwise -> show_general_analytics_options
:--
```

## Nutrition Example

A practical example of using PML for a nutrition application:

```python
async def nutrition_flow(user_message, user_profile=None):
    # Check input against guardrails
    :ask guardrail_check
    Does this user query contain any harmful content or request dangerous advice?
    Input: "{user_message}"
    :return_type GuardrailOutput
    :--

    if not guardrail_check.accept:
        return {"error": "Input blocked", "reason": guardrail_check.reply}

    # Determine routing
    :ask route_decision
    Should this nutrition query be routed to general nutrition advice,
    specialized dietary recommendations, or calorie calculation?
    Input: "{user_message}"
    User profile: {user_profile}
    :return_type RouteDecision
    :--

    # Route to appropriate handler
    if route_decision.route == "calorie_calculation":
        return await calculate_calories(user_message, user_profile)
    elif route_decision.route == "dietary_recommendations":
        return await specialized_diet_recommendations(user_message, user_profile)
    else:
        return await general_nutrition_advice(user_message)
```

## Image Analysis Integration

PML can handle complex inputs like images:

```
:ask image_analysis
Analyze this image of a food plate and identify all ingredients visible.
:input
  image: {base64_image_data}
  user_question: "What's in my meal?"
:return_type FoodAnalysisResult
:--
```

## Benefits Across Use Cases

Regardless of the specific application, PML provides:

1. **Consistency**: A standardized way to interact with LLMs
2. **Workflow Control**: Ability to create complex decision trees
3. **Readability**: Self-documenting code that clearly shows LLM interactions
4. **Modularity**: Easy to reuse and compose AI-assisted components
5. **Progressive Enhancement**: Code that can benefit from LLM improvements without rewriting

These use cases demonstrate how PML can enhance productivity and enable new capabilities in various domains.
