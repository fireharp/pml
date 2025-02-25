# PML Use Cases

This document outlines practical applications and use cases for PML in various domains, highlighting its interactive nature.

## AI-Assisted Development

### Code Generation

PML enables developers to generate executable code directly in their workflow:

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

### Code Refactoring

Use PML to refactor your existing code with direct improvements:

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

### Interactive Debugging

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

## Interactive Data Analysis Workflows

### Dynamic SQL Generation

Generate and execute SQL queries interactively:

````python
def get_sales_data(date_range, min_purchase_count=5):
    :do generate_sql
    Write a SQL query to find customers who purchased more than {min_purchase_count} products
    between {date_range.start} and {date_range.end}.
    :--

    # After processing:
    :do generate_sql
    Write a SQL query to find customers who purchased more than {min_purchase_count} products
    between {date_range.start} and {date_range.end}.
    :--(happy_panda:
    ```sql
    SELECT
        customer_id,
        customer_name,
        COUNT(order_id) as order_count,
        SUM(quantity) as total_items
    FROM
        orders
    WHERE
        order_date BETWEEN '{date_range.start}' AND '{date_range.end}'
    GROUP BY
        customer_id, customer_name
    HAVING
        SUM(quantity) > {min_purchase_count}
    ORDER BY
        total_items DESC;
    ```
    )

    # The generated SQL can be executed directly
    results = db.execute(generate_sql.result)
    return results
````

### Automated Data Visualization

Create visualizations with minimal code:

```python
import pandas as pd
import matplotlib.pyplot as plt

sales_data = pd.read_csv('quarterly_sales.csv')

:do visualize_data
Create Python code to visualize quarterly sales trends from this DataFrame.
The columns are: date, product_category, revenue, profit_margin.
:--

# After processing:
:do visualize_data
Create Python code to visualize quarterly sales trends from this DataFrame.
The columns are: date, product_category, revenue, profit_margin.
:--(happy_panda:
# Convert date to datetime
sales_data['date'] = pd.to_datetime(sales_data['date'])
sales_data['quarter'] = sales_data['date'].dt.to_period('Q')

# Aggregate by quarter and product category
quarterly_summary = sales_data.groupby(['quarter', 'product_category']).agg({
    'revenue': 'sum',
    'profit_margin': 'mean'
}).reset_index()

# Plot the results
fig, (ax1, ax2) = plt.subplots(2, 1, figsize=(12, 10))

# Revenue by quarter and category
pivot_revenue = quarterly_summary.pivot(index='quarter', columns='product_category', values='revenue')
pivot_revenue.plot(kind='bar', ax=ax1)
ax1.set_title('Quarterly Revenue by Product Category')
ax1.set_ylabel('Revenue ($)')

# Profit margin by quarter and category
pivot_margin = quarterly_summary.pivot(index='quarter', columns='product_category', values='profit_margin')
pivot_margin.plot(kind='line', marker='o', ax=ax2)
ax2.set_title('Quarterly Profit Margin by Product Category')
ax2.set_ylabel('Profit Margin (%)')

plt.tight_layout()
)

# Execute the visualization
exec(visualize_data.result)
plt.show()
```

## End-to-End Development Workflows

### Interactive API Development

Build and test APIs with minimal boilerplate:

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

### Document Generation with Code Integration

Generate complete technical documents with embedded, executable code:

````python
:do technical_doc
Create a comprehensive guide for implementing a Redis caching layer in a Python web application.
Include code examples and performance considerations.
:--

# After processing:
:do technical_doc
Create a comprehensive guide for implementing a Redis caching layer in a Python web application.
Include code examples and performance considerations.
:--(happy_panda:
# Implementing Redis Caching in Python Web Applications

## Introduction

Redis is an in-memory data structure store that can be used as a database, cache, and message broker.
This guide shows how to implement Redis as a caching layer in Python web applications.

## Installation

```python
pip install redis flask
````

## Basic Implementation

```python
import redis
from flask import Flask, request, jsonify
import json
from functools import wraps
import time

app = Flask(__name__)
redis_client = redis.Redis(host='localhost', port=6379, db=0)

def redis_cache(expiration=3600):
    def decorator(f):
        @wraps(f)
        def decorated_function(*args, **kwargs):
            # Create a cache key based on the function name and arguments
            cache_key = f"{f.__name__}:{request.path}:{request.query_string.decode('utf-8')}"

            # Try to get cached response
            cached_response = redis_client.get(cache_key)
            if cached_response:
                return json.loads(cached_response)

            # If no cache, call the original function
            result = f(*args, **kwargs)

            # Cache the result
            redis_client.setex(
                cache_key,
                expiration,
                json.dumps(result)
            )

            return result
        return decorated_function
    return decorator

@app.route('/api/data')
@redis_cache(expiration=60)  # Cache for 60 seconds
def get_data():
    # Simulate expensive operation
    time.sleep(2)
    return jsonify({"data": "This response is now cached"})
```

## Performance Considerations

1. **Appropriate Cache Durations**: Set expiration times based on data volatility
2. **Cache Invalidation**: Implement mechanisms to invalidate cache when data changes
3. **Memory Management**: Monitor Redis memory usage and set appropriate maxmemory policy
4. **Connection Pooling**: Use connection pools for efficient Redis connections

## Advanced Usage: Cache Patterns

### Cache-Aside Pattern

[Implementation details follow...]
)

# The generated documentation can be saved to a file:

with open('redis_caching_guide.md', 'w') as f:
f.write(technical_doc.result)

```

## Benefits of Interactive PML

The interactive nature of PML provides key advantages:

1. **Seamless Code Integration**: Generated code is directly usable in your workflow
2. **Live Execution**: Results can be executed immediately without manual copying/pasting
3. **Iterative Development**: Easily refine generated code through follow-up directives
4. **Context Awareness**: PML directives understand the surrounding code context
5. **First-Class Programming Experience**: AI assistance becomes a native part of the development process

These use cases demonstrate how PML transforms AI assistance from mere suggestions into executable, integrated components of your development workflow.
```
