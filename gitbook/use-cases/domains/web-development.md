# Web Development with PML

PML provides powerful capabilities for web development workflows, streamlining API development, frontend implementation, and full-stack application creation.

## RESTful API Development

Quickly scaffold and implement RESTful APIs with built-in validation and documentation:

```python
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

app = FastAPI(title="Product API")

:do create_product_model
Create a Pydantic model for a product with fields for id, name, price, and inventory_count.
:--

# After processing:
:do create_product_model
Create a Pydantic model for a product with fields for id, name, price, and inventory_count.
:--(happy_panda:
class Product(BaseModel):
    id: int
    name: str
    price: float
    inventory_count: int

    class Config:
        schema_extra = {
            "example": {
                "id": 1,
                "name": "Laptop",
                "price": 999.99,
                "inventory_count": 10
            }
        }
)

# Use the generated model in your endpoint definitions
products = []

@app.post("/products/", response_model=Product)
async def create_product(product: Product):
    products.append(product)
    return product

:do create_get_endpoints
Create endpoints to:
1. Get all products
2. Get a product by ID (with proper error handling if product not found)
:--

# After processing:
:do create_get_endpoints
Create endpoints to:
1. Get all products
2. Get a product by ID (with proper error handling if product not found)
:--(happy_panda:
@app.get("/products/", response_model=list[Product])
async def get_products():
    return products

@app.get("/products/{product_id}", response_model=Product)
async def get_product(product_id: int):
    for product in products:
        if product.id == product_id:
            return product
    raise HTTPException(status_code=404, detail="Product not found")
)

# The API is now functional with proper validation and error handling
```

## Frontend Development with React

Generate React components and hooks with PML:

```jsx
import React, { useState, useEffect } from 'react';
import axios from 'axios';

:do create_product_list
Create a React component that fetches and displays products from our API.
Include loading and error states.
:--

# After processing:
:do create_product_list
Create a React component that fetches and displays products from our API.
Include loading and error states.
:--(happy_panda:
const ProductList = () => {
  const [products, setProducts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchProducts = async () => {
      try {
        setLoading(true);
        const response = await axios.get('/api/products');
        setProducts(response.data);
        setError(null);
      } catch (err) {
        setError('Failed to fetch products. Please try again later.');
        console.error('Error fetching products:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchProducts();
  }, []);

  if (loading) return <div className="loading">Loading products...</div>;
  if (error) return <div className="error">{error}</div>;

  return (
    <div className="product-list">
      <h2>Available Products</h2>
      {products.length === 0 ? (
        <p>No products available.</p>
      ) : (
        <ul>
          {products.map((product) => (
            <li key={product.id} className="product-item">
              <h3>{product.name}</h3>
              <p>Price: ${product.price.toFixed(2)}</p>
              <p>In Stock: {product.inventory_count}</p>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};

export default ProductList;
)

// The component can be imported and used in your React application
```

## Full-Stack Integration

PML excels at bridging backend and frontend development:

```python
:do api_client_generator
Create a JavaScript/TypeScript API client for our Product API.
The client should handle authentication and provide methods for all endpoints.
:--

# After processing:
:do api_client_generator
Create a JavaScript/TypeScript API client for our Product API.
The client should handle authentication and provide methods for all endpoints.
:--(happy_panda:
// ProductApiClient.ts
import axios, { AxiosInstance } from 'axios';

interface Product {
  id: number;
  name: string;
  price: number;
  inventory_count: number;
}

class ProductApiClient {
  private client: AxiosInstance;

  constructor(baseURL: string, apiKey?: string) {
    this.client = axios.create({
      baseURL,
      headers: {
        'Content-Type': 'application/json',
        ...(apiKey && { 'Authorization': `Bearer ${apiKey}` }),
      },
    });
  }

  async getAllProducts(): Promise<Product[]> {
    const response = await this.client.get<Product[]>('/products/');
    return response.data;
  }

  async getProductById(id: number): Promise<Product> {
    const response = await this.client.get<Product>(`/products/${id}`);
    return response.data;
  }

  async createProduct(product: Omit<Product, 'id'>): Promise<Product> {
    const response = await this.client.post<Product>('/products/', product);
    return response.data;
  }

  async updateProduct(id: number, product: Partial<Product>): Promise<Product> {
    const response = await this.client.put<Product>(`/products/${id}`, product);
    return response.data;
  }

  async deleteProduct(id: number): Promise<void> {
    await this.client.delete(`/products/${id}`);
  }
}

export default ProductApiClient;
)

// Usage in a full-stack application:
const apiClient = new ProductApiClient('https://api.example.com', 'your-api-key');

// React integration
function ProductManager() {
  const [products, setProducts] = useState([]);

  useEffect(() => {
    async function loadProducts() {
      const data = await apiClient.getAllProducts();
      setProducts(data);
    }
    loadProducts();
  }, []);

  // ...rest of component
}
```

## Authentication Implementation

Secure your web applications with PML-assisted authentication:

```python
from fastapi import FastAPI, Depends, HTTPException, status
from fastapi.security import OAuth2PasswordBearer, OAuth2PasswordRequestForm
from jose import JWTError, jwt
from passlib.context import CryptContext
from datetime import datetime, timedelta
from pydantic import BaseModel

app = FastAPI()

:do auth_setup
Create a JWT authentication system for our API with:
1. User model
2. Password hashing
3. Token generation and verification
4. Protected endpoint decorator
:--

# After processing:
:do auth_setup
Create a JWT authentication system for our API with:
1. User model
2. Password hashing
3. Token generation and verification
4. Protected endpoint decorator
:--(happy_panda:
# Security configuration
SECRET_KEY = "YOUR_SECRET_KEY"  # In production, use a secure key and store in environment variables
ALGORITHM = "HS256"
ACCESS_TOKEN_EXPIRE_MINUTES = 30

# Password context
pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")
oauth2_scheme = OAuth2PasswordBearer(tokenUrl="token")

# Models
class Token(BaseModel):
    access_token: str
    token_type: str

class TokenData(BaseModel):
    username: str | None = None

class User(BaseModel):
    username: str
    email: str | None = None
    full_name: str | None = None
    disabled: bool | None = None

class UserInDB(User):
    hashed_password: str

# User database (in a real app, this would be a database)
fake_users_db = {
    "johndoe": {
        "username": "johndoe",
        "full_name": "John Doe",
        "email": "johndoe@example.com",
        "hashed_password": pwd_context.hash("secret"),
        "disabled": False,
    }
}

# Helper functions
def verify_password(plain_password, hashed_password):
    return pwd_context.verify(plain_password, hashed_password)

def get_user(db, username: str):
    if username in db:
        user_dict = db[username]
        return UserInDB(**user_dict)

def authenticate_user(fake_db, username: str, password: str):
    user = get_user(fake_db, username)
    if not user:
        return False
    if not verify_password(password, user.hashed_password):
        return False
    return user

def create_access_token(data: dict, expires_delta: timedelta | None = None):
    to_encode = data.copy()
    if expires_delta:
        expire = datetime.utcnow() + expires_delta
    else:
        expire = datetime.utcnow() + timedelta(minutes=15)
    to_encode.update({"exp": expire})
    encoded_jwt = jwt.encode(to_encode, SECRET_KEY, algorithm=ALGORITHM)
    return encoded_jwt

async def get_current_user(token: str = Depends(oauth2_scheme)):
    credentials_exception = HTTPException(
        status_code=status.HTTP_401_UNAUTHORIZED,
        detail="Could not validate credentials",
        headers={"WWW-Authenticate": "Bearer"},
    )
    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=[ALGORITHM])
        username: str = payload.get("sub")
        if username is None:
            raise credentials_exception
        token_data = TokenData(username=username)
    except JWTError:
        raise credentials_exception
    user = get_user(fake_users_db, username=token_data.username)
    if user is None:
        raise credentials_exception
    return user

async def get_current_active_user(current_user: User = Depends(get_current_user)):
    if current_user.disabled:
        raise HTTPException(status_code=400, detail="Inactive user")
    return current_user

# Endpoints
@app.post("/token", response_model=Token)
async def login_for_access_token(form_data: OAuth2PasswordRequestForm = Depends()):
    user = authenticate_user(fake_users_db, form_data.username, form_data.password)
    if not user:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Incorrect username or password",
            headers={"WWW-Authenticate": "Bearer"},
        )
    access_token_expires = timedelta(minutes=ACCESS_TOKEN_EXPIRE_MINUTES)
    access_token = create_access_token(
        data={"sub": user.username}, expires_delta=access_token_expires
    )
    return {"access_token": access_token, "token_type": "bearer"}

@app.get("/users/me", response_model=User)
async def read_users_me(current_user: User = Depends(get_current_active_user)):
    return current_user
)

# Now we can protect any endpoint by adding the dependency
@app.get("/protected-resource")
async def protected_resource(current_user: User = Depends(get_current_active_user)):
    return {"message": "This is a protected resource", "user": current_user}
```

## Benefits for Web Development

PML brings several unique advantages to web development workflows:

1. **Rapid Prototyping**: Generate complete API endpoints and frontend components in seconds
2. **Full-Stack Consistency**: Ensure your data models are consistent across frontend and backend
3. **Security Best Practices**: Implement authentication and security measures correctly from the start
4. **Documentation Generation**: Automatically document your APIs as you build them
5. **Cross-Framework Support**: Work with any web framework or library, from Django to React to Vue
6. **Interactive Development**: Refine components and endpoints in real-time as requirements evolve

By integrating PML into your web development workflow, you can dramatically accelerate development time while maintaining high code quality and security standards.
