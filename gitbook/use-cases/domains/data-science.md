# Data Science with PML

PML transforms data science workflows by making machine learning, statistical analysis, and data visualization more interactive and code-efficient.

## Statistical Analysis

Use PML to generate statistical analyses with clean, executable code:

```python
import pandas as pd
import numpy as np
import scipy.stats as stats

# Load dataset
housing_data = pd.read_csv('housing_data.csv')

:do dataset_exploration
Generate code to explore this housing dataset with:
1. Summary statistics
2. Distribution analysis
3. Correlation analysis between price and other features
:--

# After processing:
:do dataset_exploration
Generate code to explore this housing dataset with:
1. Summary statistics
2. Distribution analysis
3. Correlation analysis between price and other features
:--(happy_panda:
# 1. Summary statistics
print("=== Summary Statistics ===")
summary = housing_data.describe(include='all')
print(summary)

# 2. Distribution analysis
import matplotlib.pyplot as plt
import seaborn as sns

# Set up the matplotlib figure
plt.figure(figsize=(15, 10))

# Plot histograms for numerical features
numeric_features = housing_data.select_dtypes(include=[np.number]).columns.tolist()
for i, feature in enumerate(numeric_features):
    plt.subplot(3, 3, i+1)
    sns.histplot(housing_data[feature], kde=True)
    plt.title(f'Distribution of {feature}')
    plt.tight_layout()

# 3. Correlation analysis
correlation_matrix = housing_data.corr()
plt.figure(figsize=(12, 10))
sns.heatmap(correlation_matrix, annot=True, cmap='coolwarm', linewidths=0.5)
plt.title('Correlation Matrix')

# Price correlations specifically
price_correlations = correlation_matrix['price'].sort_values(ascending=False)
print("\n=== Price Correlations ===")
print(price_correlations)

# Scatter plots for top correlated features with price
top_correlated = correlation_matrix['price'].abs().sort_values(ascending=False)[1:4].index
plt.figure(figsize=(15, 5))
for i, feature in enumerate(top_correlated):
    plt.subplot(1, 3, i+1)
    sns.scatterplot(x=feature, y='price', data=housing_data)
    plt.title(f'Price vs {feature}')

plt.tight_layout()
)

# Execute the exploration
exec(dataset_exploration.result)
```

## Machine Learning Model Development

Rapidly implement and evaluate machine learning models:

```python
import pandas as pd
import numpy as np
from sklearn.model_selection import train_test_split
from sklearn.preprocessing import StandardScaler
from sklearn.metrics import mean_squared_error, r2_score

# Load and prepare data
housing_data = pd.read_csv('housing_data.csv')
X = housing_data.drop('price', axis=1)
y = housing_data['price']

# Handle categorical features and missing values
X = pd.get_dummies(X)
X.fillna(X.mean(), inplace=True)

# Split the data
X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, random_state=42)

:do model_selection
Analyze the data and suggest appropriate regression models for predicting housing prices.
Implement the top 3 models with proper hyperparameter tuning.
:--

# After processing:
:do model_selection
Analyze the data and suggest appropriate regression models for predicting housing prices.
Implement the top 3 models with proper hyperparameter tuning.
:--(happy_panda:
from sklearn.linear_model import LinearRegression, Ridge, Lasso
from sklearn.ensemble import RandomForestRegressor, GradientBoostingRegressor
from sklearn.model_selection import GridSearchCV
import matplotlib.pyplot as plt

# 1. Linear Regression (baseline)
print("Training Linear Regression model...")
lr_model = LinearRegression()
lr_model.fit(X_train, y_train)
lr_pred = lr_model.predict(X_test)
lr_rmse = np.sqrt(mean_squared_error(y_test, lr_pred))
lr_r2 = r2_score(y_test, lr_pred)
print(f"Linear Regression RMSE: {lr_rmse:.2f}")
print(f"Linear Regression R²: {lr_r2:.2f}")

# 2. Ridge Regression (with regularization)
print("\nTraining Ridge Regression model with hyperparameter tuning...")
ridge_params = {'alpha': [0.01, 0.1, 1.0, 10.0, 100.0]}
ridge_grid = GridSearchCV(Ridge(), ridge_params, cv=5, scoring='neg_mean_squared_error')
ridge_grid.fit(X_train, y_train)
ridge_model = ridge_grid.best_estimator_
ridge_pred = ridge_model.predict(X_test)
ridge_rmse = np.sqrt(mean_squared_error(y_test, ridge_pred))
ridge_r2 = r2_score(y_test, ridge_pred)
print(f"Best Ridge alpha: {ridge_grid.best_params_['alpha']}")
print(f"Ridge Regression RMSE: {ridge_rmse:.2f}")
print(f"Ridge Regression R²: {ridge_r2:.2f}")

# 3. Random Forest (more complex, non-linear relationships)
print("\nTraining Random Forest model with hyperparameter tuning...")
rf_params = {
    'n_estimators': [100, 200],
    'max_depth': [None, 10, 20],
    'min_samples_split': [2, 5]
}
rf_grid = GridSearchCV(RandomForestRegressor(random_state=42),
                       rf_params, cv=3, scoring='neg_mean_squared_error')
rf_grid.fit(X_train, y_train)
rf_model = rf_grid.best_estimator_
rf_pred = rf_model.predict(X_test)
rf_rmse = np.sqrt(mean_squared_error(y_test, rf_pred))
rf_r2 = r2_score(y_test, rf_pred)
print(f"Best Random Forest parameters: {rf_grid.best_params_}")
print(f"Random Forest RMSE: {rf_rmse:.2f}")
print(f"Random Forest R²: {rf_r2:.2f}")

# Feature importance from the Random Forest model
feature_importance = pd.DataFrame(
    {'feature': X_train.columns, 'importance': rf_model.feature_importances_}
).sort_values('importance', ascending=False)
print("\nTop 10 important features:")
print(feature_importance.head(10))

# Visualize model performance
models = ['Linear Regression', 'Ridge Regression', 'Random Forest']
rmse_scores = [lr_rmse, ridge_rmse, rf_rmse]
r2_scores = [lr_r2, ridge_r2, rf_r2]

plt.figure(figsize=(12, 5))
plt.subplot(1, 2, 1)
plt.bar(models, rmse_scores)
plt.title('RMSE Comparison (lower is better)')
plt.ylabel('RMSE')

plt.subplot(1, 2, 2)
plt.bar(models, r2_scores)
plt.title('R² Comparison (higher is better)')
plt.ylabel('R²')

plt.tight_layout()
plt.show()
)

# Execute the model selection code
exec(model_selection.result)
```

## Data Visualization and Insight Generation

Create sophisticated visualizations and extract insights:

```python
import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns

# Load your dataset
sales_data = pd.read_csv('sales_by_region.csv')

:do advanced_visualization
Create an interactive dashboard-style visualization for regional sales data.
Include:
1. Time series analysis showing trends
2. Regional comparisons
3. Key performance indicators
4. Seasonality analysis
:--

# After processing:
:do advanced_visualization
Create an interactive dashboard-style visualization for regional sales data.
Include:
1. Time series analysis showing trends
2. Regional comparisons
3. Key performance indicators
4. Seasonality analysis
:--(happy_panda:
import plotly.express as px
import plotly.graph_objects as go
from plotly.subplots import make_subplots
import numpy as np
from datetime import datetime

# Ensure date is in datetime format
sales_data['date'] = pd.to_datetime(sales_data['date'])
sales_data['month'] = sales_data['date'].dt.month
sales_data['year'] = sales_data['date'].dt.year
sales_data['quarter'] = sales_data['date'].dt.quarter

# Create a dashboard layout
fig = make_subplots(
    rows=3, cols=2,
    subplot_titles=(
        "Sales Trends Over Time", "Regional Sales Comparison",
        "Key Performance Indicators", "Monthly Sales by Region",
        "Year-over-Year Growth", "Quarterly Performance"
    ),
    specs=[
        [{"type": "scatter"}, {"type": "bar"}],
        [{"type": "indicator"}, {"type": "heatmap"}],
        [{"type": "bar"}, {"type": "bar"}]
    ],
    vertical_spacing=0.1,
    horizontal_spacing=0.1
)

# 1. Time series analysis
for region in sales_data['region'].unique():
    region_data = sales_data[sales_data['region'] == region]
    fig.add_trace(
        go.Scatter(
            x=region_data['date'],
            y=region_data['sales_amount'],
            mode='lines',
            name=region
        ),
        row=1, col=1
    )

# 2. Regional comparison (total sales by region)
regional_sales = sales_data.groupby('region')['sales_amount'].sum().reset_index()
fig.add_trace(
    go.Bar(
        x=regional_sales['region'],
        y=regional_sales['sales_amount'],
        marker_color=px.colors.qualitative.Plotly
    ),
    row=1, col=2
)

# 3. Key Performance Indicators
total_sales = sales_data['sales_amount'].sum()
avg_sales = sales_data['sales_amount'].mean()
max_region = regional_sales.loc[regional_sales['sales_amount'].idxmax(), 'region']

fig.add_trace(
    go.Indicator(
        mode="number+delta",
        value=total_sales,
        title={"text": "Total Sales"},
        delta={'reference': total_sales * 0.9, 'relative': True},
        domain={'row': 0, 'column': 0}
    ),
    row=2, col=1
)

# 4. Seasonality analysis (heatmap of monthly sales by region)
monthly_region_sales = sales_data.groupby(['region', 'month'])['sales_amount'].mean().reset_index()
monthly_region_pivot = monthly_region_sales.pivot(index='region', columns='month', values='sales_amount')

fig.add_trace(
    go.Heatmap(
        z=monthly_region_pivot.values,
        x=['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'],
        y=monthly_region_pivot.index,
        colorscale='Viridis'
    ),
    row=2, col=2
)

# 5. Year-over-year growth
yearly_sales = sales_data.groupby('year')['sales_amount'].sum().reset_index()
fig.add_trace(
    go.Bar(
        x=yearly_sales['year'],
        y=yearly_sales['sales_amount'],
        marker_color='darkblue'
    ),
    row=3, col=1
)

# 6. Quarterly performance
quarterly_sales = sales_data.groupby(['year', 'quarter'])['sales_amount'].sum().reset_index()
quarterly_labels = quarterly_sales.apply(lambda x: f"Q{x['quarter']} {x['year']}", axis=1)
fig.add_trace(
    go.Bar(
        x=quarterly_labels,
        y=quarterly_sales['sales_amount'],
        marker_color='darkgreen'
    ),
    row=3, col=2
)

# Update layout and configuration
fig.update_layout(
    height=900,
    width=1200,
    title_text="Sales Performance Dashboard",
    showlegend=False
)

fig.show()
)

# Execute the visualization code
import plotly.io as pio
pio.renderers.default = "browser"
exec(advanced_visualization.result)
```

## Natural Language Processing

Apply NLP techniques to analyze text data:

```python
import pandas as pd
import numpy as np
import re
import nltk
from nltk.corpus import stopwords

# Download NLTK resources
nltk.download('punkt')
nltk.download('stopwords')
nltk.download('wordnet')

# Sample customer feedback dataset
customer_feedback = pd.read_csv('customer_feedback.csv')

:do nlp_analysis
Perform a comprehensive NLP analysis on customer feedback data.
Include sentiment analysis, topic modeling, and key phrase extraction.
:--

# After processing:
:do nlp_analysis
Perform a comprehensive NLP analysis on customer feedback data.
Include sentiment analysis, topic modeling, and key phrase extraction.
:--(happy_panda:
import nltk
from nltk.tokenize import word_tokenize
from nltk.stem import WordNetLemmatizer
from nltk.corpus import stopwords
from nltk.sentiment import SentimentIntensityAnalyzer
from sklearn.feature_extraction.text import CountVectorizer, TfidfVectorizer
from sklearn.decomposition import LatentDirichletAllocation
import matplotlib.pyplot as plt
import seaborn as sns

# Download additional NLTK resources
nltk.download('vader_lexicon')
nltk.download('wordnet')

# Text preprocessing function
def preprocess_text(text):
    if not isinstance(text, str):
        return ""

    # Convert to lowercase and remove punctuation
    text = re.sub(r'[^\w\s]', '', text.lower())

    # Tokenize and remove stopwords
    stop_words = set(stopwords.words('english'))
    tokens = word_tokenize(text)
    tokens = [token for token in tokens if token not in stop_words]

    # Lemmatize
    lemmatizer = WordNetLemmatizer()
    tokens = [lemmatizer.lemmatize(token) for token in tokens]

    return ' '.join(tokens)

# Apply preprocessing
customer_feedback['processed_text'] = customer_feedback['feedback'].apply(preprocess_text)

# 1. Sentiment Analysis
sid = SentimentIntensityAnalyzer()
customer_feedback['sentiment_score'] = customer_feedback['feedback'].apply(
    lambda x: sid.polarity_scores(x)['compound'] if isinstance(x, str) else 0
)

# Map scores to sentiment categories
customer_feedback['sentiment'] = customer_feedback['sentiment_score'].apply(
    lambda score: 'positive' if score > 0.05 else ('negative' if score < -0.05 else 'neutral')
)

# Sentiment distribution
plt.figure(figsize=(10, 6))
sns.countplot(x='sentiment', data=customer_feedback)
plt.title('Sentiment Distribution in Customer Feedback')
plt.xlabel('Sentiment')
plt.ylabel('Count')

# 2. Topic Modeling
# Create document-term matrix
vectorizer = CountVectorizer(max_df=0.95, min_df=2, stop_words='english')
dtm = vectorizer.fit_transform(customer_feedback['processed_text'])
feature_names = vectorizer.get_feature_names_out()

# Apply LDA for topic modeling
n_topics = 5
lda = LatentDirichletAllocation(n_components=n_topics, random_state=42)
lda.fit(dtm)

# Print top terms for each topic
print("Top terms per topic:")
for topic_idx, topic in enumerate(lda.components_):
    top_terms_idx = topic.argsort()[:-11:-1]  # Get indices of top 10 terms
    top_terms = [feature_names[i] for i in top_terms_idx]
    print(f"Topic {topic_idx+1}: {', '.join(top_terms)}")

# Assign topic to each document
topic_results = lda.transform(dtm)
customer_feedback['dominant_topic'] = topic_results.argmax(axis=1) + 1

# Topic distribution
plt.figure(figsize=(10, 6))
sns.countplot(x='dominant_topic', data=customer_feedback)
plt.title('Distribution of Dominant Topics')
plt.xlabel('Topic Number')
plt.ylabel('Count')

# 3. Key Phrase Extraction using TF-IDF
tfidf_vectorizer = TfidfVectorizer(max_df=0.95, min_df=2, stop_words='english')
tfidf = tfidf_vectorizer.fit_transform(customer_feedback['processed_text'])
tfidf_feature_names = tfidf_vectorizer.get_feature_names_out()

# Extract top keywords for each document
def get_top_keywords(x, n_terms=5):
    if x.sum() == 0:
        return []

    # Get indices of top n TF-IDF values
    top_indices = x.toarray().flatten().argsort()[:-n_terms-1:-1]
    # Get corresponding feature names
    return [tfidf_feature_names[i] for i in top_indices]

# This is a bit computation-intensive but provides key phrases for each review
customer_feedback['key_phrases'] = [
    get_top_keywords(tfidf[i]) for i in range(tfidf.shape[0])
]

# 4. Cross analysis: Sentiment by Topic
sentiment_topic_cross = pd.crosstab(
    customer_feedback['dominant_topic'],
    customer_feedback['sentiment']
)

plt.figure(figsize=(12, 6))
sentiment_topic_cross.plot(kind='bar', stacked=True)
plt.title('Sentiment Distribution Across Topics')
plt.xlabel('Topic')
plt.ylabel('Count')
plt.legend(title='Sentiment')

# Return a summary of findings
sentiment_counts = customer_feedback['sentiment'].value_counts()
dominant_topic = customer_feedback['dominant_topic'].value_counts().idxmax()

print(f"Summary of Analysis:")
print(f"Total reviews analyzed: {len(customer_feedback)}")
print(f"Sentiment distribution: {sentiment_counts.to_dict()}")
print(f"Most common topic: Topic {dominant_topic}")
print(f"Most positive topic: Topic {sentiment_topic_cross['positive'].idxmax()}")
print(f"Most negative topic: Topic {sentiment_topic_cross['negative'].idxmax()}")
)

# Execute the NLP analysis code
exec(nlp_analysis.result)
```

## Benefits for Data Science

PML offers unique advantages for data science workflows:

1. **Rapid Prototyping**: Generate complex analytical code in seconds instead of hours
2. **Interactive Model Development**: Quickly test and refine different approaches
3. **Reproducible Analyses**: Create shareable, executable documents with embedded analyses
4. **Automated Visualization**: Generate complex visualizations without memorizing APIs
5. **Methodological Guidance**: Get built-in best practices for statistical analysis
6. **End-to-End Pipelines**: Develop complete data workflows from cleaning to visualization

PML transforms how data scientists work by:

- Minimizing boilerplate code in analytical workflows
- Providing guidance on appropriate statistical methods
- Automating common data science tasks
- Creating reproducible, self-documenting analyses

By integrating PML into your data science workflow, you can focus more on analyzing results and drawing conclusions rather than debugging code or researching API documentation.
