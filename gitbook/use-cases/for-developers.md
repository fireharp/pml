# PML for Developers

PML provides developers with a seamless way to integrate AI assistance directly into their coding workflow. This document showcases practical examples of how developers can use PML to enhance their productivity and code quality.

## Code Generation

Generate complete, executable code directly in your workflow:

```python
def calculate_fibonacci(n):
    :do write_fibonacci
    Write an efficient Python function that calculates the nth Fibonacci number iteratively.
    :--

# After processing, the directive becomes:
def calculate_fibonacci(n):
    :do write_fibonacci
    Write an efficient Python function that calculates the nth Fibonacci number iteratively.
    :--(happy_panda:
        a, b = 0, 1
        for _ in range(n):
            a, b = b, a + b
        return a
    )

# The function is now usable in your code:
result = calculate_fibonacci(10)  # Returns 55
```

## Code Refactoring

Refactor existing code with direct improvements:

```python
# Original slow implementation
def slow_function(n):
    result = []
    for i in range(n):
        result.append(i**2)
    return result

# Use PML to refactor
:do refactor_function
Refactor this function to be more efficient:

def slow_function(n):
    result = []
    for i in range(n):
        result.append(i**2)
    return result
:--

# After processing, you get an optimized implementation:
:do refactor_function
Refactor this function to be more efficient:

def slow_function(n):
    result = []
    for i in range(n):
        result.append(i**2)
    return result
:--(happy_panda:
def fast_function(n):
    # List comprehension is more efficient than appending in a loop
    return [i**2 for i in range(n)]
)

# Now you can directly use the optimized function
data = fast_function(1000)
```

## Interactive Debugging

Debug problems with immediate suggestions and fixes:

```python
# Problem code with a bug
def merge_dicts(dict1, dict2):
    result = dict1
    result.update(dict2)
    return result

# Strange behavior observed
a = {'x': 1}
b = {'y': 2}
c = merge_dicts(a, b)
print(a)  # Unexpectedly prints {'x': 1, 'y': 2}

:do fix_bug
Debug this function that's causing unexpected side effects:

def merge_dicts(dict1, dict2):
    result = dict1
    result.update(dict2)
    return result
:--

# After processing:
:do fix_bug
Debug this function that's causing unexpected side effects:

def merge_dicts(dict1, dict2):
    result = dict1
    result.update(dict2)
    return result
:--(happy_panda:
def merge_dicts(dict1, dict2):
    # Create a new dictionary instead of modifying the input
    result = dict1.copy()
    result.update(dict2)
    return result
)

# Now you can use the fixed function
a = {'x': 1}
b = {'y': 2}
c = merge_dicts(a, b)
print(a)  # Correctly prints {'x': 1}
print(c)  # Prints {'x': 1, 'y': 2}
```

## API Development

Build APIs with minimal boilerplate:

```python
from fastapi import FastAPI

app = FastAPI()

:do create_api_endpoint
Create a FastAPI endpoint that processes user reviews, extracts sentiment,
and returns a summary with sentiment score.
:--

# After processing:
:do create_api_endpoint
Create a FastAPI endpoint that processes user reviews, extracts sentiment,
and returns a summary with sentiment score.
:--(happy_panda:
from pydantic import BaseModel
from typing import List
import nltk
from nltk.sentiment import SentimentIntensityAnalyzer

# Download necessary NLTK data
nltk.download('vader_lexicon')

# Initialize sentiment analyzer
sia = SentimentIntensityAnalyzer()

class Review(BaseModel):
    text: str

class ReviewResponse(BaseModel):
    sentiment_score: float
    sentiment_category: str
    summary: str

@app.post("/analyze-review", response_model=ReviewResponse)
async def analyze_review(review: Review):
    # Get sentiment scores
    sentiment = sia.polarity_scores(review.text)

    # Determine sentiment category
    if sentiment['compound'] >= 0.05:
        category = "positive"
    elif sentiment['compound'] <= -0.05:
        category = "negative"
    else:
        category = "neutral"

    # Generate summary (in real app, you might use an LLM here)
    summary = review.text[:100] + "..." if len(review.text) > 100 else review.text

    return ReviewResponse(
        sentiment_score=sentiment['compound'],
        sentiment_category=category,
        summary=summary
    )
)

# The endpoint is now available in your FastAPI app
```

## Testing Code Generation

Generate test cases for your functions:

```python
def validate_email(email):
    import re
    pattern = r'^[\w\.-]+@[\w\.-]+\.\w+$'
    return bool(re.match(pattern, email))

:do generate_tests
Write pytest test cases for the validate_email function,
including edge cases and various valid/invalid formats.
:--

# After processing:
:do generate_tests
Write pytest test cases for the validate_email function,
including edge cases and various valid/invalid formats.
:--(happy_panda:
import pytest

def test_validate_email_valid_formats():
    # Test standard valid email formats
    assert validate_email("user@example.com") == True
    assert validate_email("first.last@example.com") == True
    assert validate_email("user-name@domain.com") == True
    assert validate_email("user123@domain.co.uk") == True

def test_validate_email_invalid_formats():
    # Test invalid email formats
    assert validate_email("user@domain") == False
    assert validate_email("user@.com") == False
    assert validate_email("@domain.com") == False
    assert validate_email("user@domain.") == False
    assert validate_email("user domain.com") == False

def test_validate_email_edge_cases():
    # Test edge cases
    assert validate_email("a@b.c") == True  # Minimal valid format
    assert validate_email("") == False
    assert validate_email(None) == False  # This would fail unless function handles None

def test_validate_email_special_characters():
    # Test emails with special characters
    assert validate_email("user+tag@domain.com") == True  # Gmail-style tagging
    assert validate_email("user!#$%@domain.com") == False  # Invalid special chars
)

# The tests can be run directly with pytest
```

## Benefits for Developers

For developers, PML offers significant advantages:

1. **Reduced Boilerplate**: Generate common code patterns without writing them from scratch
2. **Constant Learning**: Get real-time suggestions for best practices and optimizations
3. **Documentation Integration**: Generate and maintain documentation alongside code
4. **Testing Acceleration**: Create comprehensive test suites with minimal effort
5. **Tool Augmentation**: Extend your IDE capabilities with AI assistance that works the way you do
