# PML for Data Analysts

PML empowers data analysts to streamline their workflows by integrating AI assistance directly into their data analysis process. This document showcases practical examples of how analysts can use PML to enhance their productivity and insights.

## Dynamic SQL Generation

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

## Automated Data Visualization

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

## Data Exploration Assistance

Get guidance on exploratory data analysis:

```python
import pandas as pd

customer_data = pd.read_csv('customer_data.csv')

:do explore_data
Provide a comprehensive exploratory data analysis plan for this customer dataset.
The dataset contains demographic information, purchase history, and customer satisfaction scores.
:--

# After processing:
:do explore_data
Provide a comprehensive exploratory data analysis plan for this customer dataset.
The dataset contains demographic information, purchase history, and customer satisfaction scores.
:--(happy_panda:
# 1. Basic Data Overview
print("Dataset shape:", customer_data.shape)
print("\nData types:")
print(customer_data.dtypes)
print("\nBasic statistics:")
print(customer_data.describe(include='all'))

# 2. Missing Value Analysis
print("\nMissing values per column:")
missing_values = customer_data.isnull().sum()
print(missing_values[missing_values > 0])

# 3. Demographic Analysis
print("\nCustomer demographics:")
for col in ['age_group', 'gender', 'location', 'income_bracket']:
    if col in customer_data.columns:
        print(f"\n{col.title()} distribution:")
        print(customer_data[col].value_counts(normalize=True).round(3) * 100)

# 4. Purchase Behavior
print("\nPurchase behavior:")
for col in ['total_purchases', 'avg_order_value', 'preferred_category']:
    if col in customer_data.columns:
        if customer_data[col].dtype in ['int64', 'float64']:
            print(f"\n{col.replace('_', ' ').title()} statistics:")
            print(customer_data[col].describe())
        else:
            print(f"\n{col.replace('_', ' ').title()} distribution:")
            print(customer_data[col].value_counts(normalize=True).round(3) * 100)

# 5. Satisfaction Analysis
if 'satisfaction_score' in customer_data.columns:
    print("\nSatisfaction score distribution:")
    print(customer_data['satisfaction_score'].value_counts(normalize=True).round(3) * 100)

    # Correlation with other metrics
    print("\nCorrelation with satisfaction:")
    numeric_cols = customer_data.select_dtypes(include=['int64', 'float64']).columns
    correlations = customer_data[numeric_cols].corr()['satisfaction_score'].sort_values(ascending=False)
    print(correlations)

# 6. Visualizations (would be implemented with matplotlib/seaborn)
print("\nRecommended visualizations:")
print("- Age group vs. Satisfaction (bar chart)")
print("- Purchase frequency vs. Average order value (scatter plot)")
print("- Satisfaction distribution by preferred category (box plot)")
print("- Income bracket vs. Total spend (bar chart)")
)

# Execute the exploration plan
exec(explore_data.result)
```

## Statistical Analysis

Generate statistical analysis code:

```python
import pandas as pd
import numpy as np

sales_data = pd.read_csv('sales_performance.csv')

:do statistical_analysis
Perform a statistical analysis to determine if there's a significant difference
in sales performance between our two marketing strategies (A and B).
The 'strategy' column contains the strategy name, and 'revenue' contains the results.
:--

# After processing:
:do statistical_analysis
Perform a statistical analysis to determine if there's a significant difference
in sales performance between our two marketing strategies (A and B).
The 'strategy' column contains the strategy name, and 'revenue' contains the results.
:--(happy_panda:
import scipy.stats as stats
import matplotlib.pyplot as plt
import seaborn as sns

# Split data by strategy
strategy_a = sales_data[sales_data['strategy'] == 'A']['revenue']
strategy_b = sales_data[sales_data['strategy'] == 'B']['revenue']

# Basic statistics
print("Strategy A statistics:")
print(strategy_a.describe())
print("\nStrategy B statistics:")
print(strategy_b.describe())

# Visualization
plt.figure(figsize=(12, 6))

# Box plot
plt.subplot(1, 2, 1)
sns.boxplot(x='strategy', y='revenue', data=sales_data)
plt.title('Revenue by Marketing Strategy')

# Distribution plot
plt.subplot(1, 2, 2)
sns.histplot(strategy_a, color='blue', alpha=0.5, label='Strategy A')
sns.histplot(strategy_b, color='red', alpha=0.5, label='Strategy B')
plt.title('Revenue Distribution by Strategy')
plt.legend()

plt.tight_layout()
plt.show()

# Normality test
print("\nNormality test (Shapiro-Wilk):")
print("Strategy A:", stats.shapiro(strategy_a))
print("Strategy B:", stats.shapiro(strategy_b))

# Choose appropriate statistical test based on normality
# If both are normally distributed, use t-test
# Otherwise, use non-parametric Mann-Whitney U test
alpha = 0.05

if stats.shapiro(strategy_a)[1] > alpha and stats.shapiro(strategy_b)[1] > alpha:
    # T-test for normally distributed data
    t_stat, p_value = stats.ttest_ind(strategy_a, strategy_b, equal_var=False)
    test_name = "Independent samples t-test"
else:
    # Non-parametric test for non-normal data
    t_stat, p_value = stats.mannwhitneyu(strategy_a, strategy_b)
    test_name = "Mann-Whitney U test"

print(f"\n{test_name} results:")
print(f"Test statistic: {t_stat:.4f}")
print(f"P-value: {p_value:.4f}")

# Interpretation
if p_value < alpha:
    print(f"\nResult: There is a statistically significant difference between strategies (p < {alpha}).")
    better_strategy = 'A' if strategy_a.mean() > strategy_b.mean() else 'B'
    print(f"Strategy {better_strategy} appears to perform better.")
else:
    print(f"\nResult: There is no statistically significant difference between strategies (p > {alpha}).")
)

# Execute the statistical analysis
exec(statistical_analysis.result)
```

## Benefits for Data Analysts

For data analysts, PML offers significant advantages:

1. **Accelerated Analysis**: Generate complex SQL queries and analysis code without writing from scratch
2. **Visualization Assistance**: Create effective data visualizations with minimal effort
3. **Statistical Guidance**: Get help with appropriate statistical tests and interpretations
4. **Reproducible Research**: Document your analysis process alongside your code
5. **Continuous Learning**: Discover new analysis techniques and best practices through AI assistance
