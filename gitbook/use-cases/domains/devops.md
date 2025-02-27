# DevOps with PML

PML revolutionizes DevOps workflows by automating infrastructure as code, CI/CD pipelines, and monitoring solutions with an interactive, code-efficient approach.

## Infrastructure Automation

Use PML to generate and manage infrastructure as code:

```python
import boto3
import json

# Set up AWS credentials and region
region = "us-west-2"

:do create_infrastructure
Generate Terraform code for a scalable AWS infrastructure with:
1. VPC with multiple availability zones
2. Application load balancer
3. Auto-scaling EC2 instances
4. RDS database
:--

# After processing:
:do create_infrastructure
Generate Terraform code for a scalable AWS infrastructure with:
1. VPC with multiple availability zones
2. Application load balancer
3. Auto-scaling EC2 instances
4. RDS database
:--(happy_panda:
# main.tf
provider "aws" {
  region = "us-west-2"
}

# VPC and networking
module "vpc" {
  source = "terraform-aws-modules/vpc/aws"
  version = "3.14.0"

  name = "app-vpc"
  cidr = "10.0.0.0/16"

  azs             = ["us-west-2a", "us-west-2b", "us-west-2c"]
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  public_subnets  = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]

  enable_nat_gateway = true
  single_nat_gateway = false
  one_nat_gateway_per_az = true

  tags = {
    Environment = "production"
    Project     = "MyApp"
  }
}

# Security groups
resource "aws_security_group" "alb" {
  name        = "alb-sg"
  description = "ALB Security Group"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "app" {
  name        = "app-sg"
  description = "Application Security Group"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port       = 80
    to_port         = 80
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "db" {
  name        = "db-sg"
  description = "Database Security Group"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port       = 3306
    to_port         = 3306
    protocol        = "tcp"
    security_groups = [aws_security_group.app.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# Application Load Balancer
resource "aws_lb" "app" {
  name               = "app-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = module.vpc.public_subnets
}

resource "aws_lb_target_group" "app" {
  name     = "app-tg"
  port     = 80
  protocol = "HTTP"
  vpc_id   = module.vpc.vpc_id

  health_check {
    path                = "/health"
    port                = "traffic-port"
    healthy_threshold   = 3
    unhealthy_threshold = 3
    timeout             = 5
    interval            = 30
  }
}

resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.app.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.app.arn
  }
}

# Auto Scaling Group
resource "aws_launch_template" "app" {
  name_prefix   = "app-"
  image_id      = "ami-0c55b159cbfafe1f0" # Amazon Linux 2 AMI
  instance_type = "t3.micro"

  vpc_security_group_ids = [aws_security_group.app.id]

  user_data = base64encode(<<-EOF
    #!/bin/bash
    amazon-linux-extras install -y nginx
    systemctl start nginx
    systemctl enable nginx
    echo '<html><body><h1>Hello, World!</h1></body></html>' > /usr/share/nginx/html/index.html
    echo '<html><body><h1>Health Check OK</h1></body></html>' > /usr/share/nginx/html/health
  EOF
  )
}

resource "aws_autoscaling_group" "app" {
  desired_capacity    = 2
  max_size            = 5
  min_size            = 2
  vpc_zone_identifier = module.vpc.private_subnets

  launch_template {
    id      = aws_launch_template.app.id
    version = "$Latest"
  }

  target_group_arns = [aws_lb_target_group.app.arn]

  tag {
    key                 = "Name"
    value               = "app-instance"
    propagate_at_launch = true
  }
}

# RDS Database
resource "aws_db_subnet_group" "app" {
  name       = "app-db-subnet-group"
  subnet_ids = module.vpc.private_subnets
}

resource "aws_db_instance" "app" {
  allocated_storage    = 20
  storage_type         = "gp2"
  engine               = "mysql"
  engine_version       = "8.0"
  instance_class       = "db.t3.micro"
  db_name              = "appdb"
  username             = "admin"
  password             = "Password123!" # In production, use AWS Secrets Manager
  parameter_group_name = "default.mysql8.0"
  db_subnet_group_name = aws_db_subnet_group.app.name
  vpc_security_group_ids = [aws_security_group.db.id]
  skip_final_snapshot  = true

  tags = {
    Environment = "production"
  }
}

# Outputs
output "alb_dns_name" {
  description = "The DNS name of the load balancer"
  value       = aws_lb.app.dns_name
}

output "rds_endpoint" {
  description = "The endpoint of the database"
  value       = aws_db_instance.app.endpoint
}
)

# Execute the infrastructure code generation
with open('main.tf', 'w') as f:
    f.write(create_infrastructure.result)
```

## CI/CD Pipeline Generation

Automate your CI/CD pipeline creation and management:

```python
:do create_github_actions
Generate GitHub Actions workflows for a Python microservice with:
1. Unit and integration testing
2. Code quality checks
3. Container builds
4. Deployment to Kubernetes
:--

# After processing:
:do create_github_actions
Generate GitHub Actions workflows for a Python microservice with:
1. Unit and integration testing
2. Code quality checks
3. Container builds
4. Deployment to Kubernetes
:--(happy_panda:
# .github/workflows/ci.yml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.10'
          cache: 'pip'

      - name: Install dependencies
        run: |
          python -m pip install --upgrade pip
          pip install -r requirements.txt
          pip install -r requirements-dev.txt

      - name: Run linting
        run: |
          flake8 .
          black --check .
          isort --check .

      - name: Run unit tests
        run: pytest tests/unit

      - name: Run integration tests
        run: pytest tests/integration

      - name: Generate test coverage report
        run: |
          pytest --cov=app --cov-report=xml

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.xml

  build:
    needs: test
    if: github.event_name == 'push' || github.event.pull_request.merged == true
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2

      - name: Login to DockerHub
        uses: docker/login-action@v2
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_PASSWORD }}

      - name: Build and push Docker image
        uses: docker/build-push-action@v4
        with:
          context: .
          push: true
          tags: |
            yourcompany/microservice:latest
            yourcompany/microservice:${{ github.sha }}
          cache-from: type=registry,ref=yourcompany/microservice:latest
          cache-to: type=inline

  deploy:
    needs: build
    if: github.ref == 'refs/heads/main' && (github.event_name == 'push' || github.event.pull_request.merged == true)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up kubectl
        uses: azure/setup-kubectl@v3

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v2
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-west-2

      - name: Update kubeconfig
        run: aws eks update-kubeconfig --name production-cluster --region us-west-2

      - name: Deploy to Kubernetes
        run: |
          # Update image tag in deployment yaml
          sed -i "s|image: yourcompany/microservice:.*|image: yourcompany/microservice:${{ github.sha }}|g" k8s/deployment.yaml

          # Apply Kubernetes manifests
          kubectl apply -f k8s/

          # Wait for deployment to complete
          kubectl rollout status deployment/microservice

      - name: Run post-deployment tests
        run: |
          # Wait for service to be ready
          sleep 30

          # Run smoke tests
          pytest tests/smoke
)

# Save the GitHub Actions workflow file
import os
os.makedirs('.github/workflows', exist_ok=True)
with open('.github/workflows/ci.yml', 'w') as f:
    f.write(create_github_actions.result)
```

## Monitoring Setup

Automate the creation of monitoring and observability solutions:

```python
:do create_monitoring_stack
Generate a Docker Compose configuration for a complete monitoring stack with:
1. Prometheus for metrics collection
2. Grafana for visualization
3. Loki for logs aggregation
4. Alertmanager for alerting
Include basic configurations and dashboards.
:--

# After processing:
:do create_monitoring_stack
Generate a Docker Compose configuration for a complete monitoring stack with:
1. Prometheus for metrics collection
2. Grafana for visualization
3. Loki for logs aggregation
4. Alertmanager for alerting
Include basic configurations and dashboards.
:--(happy_panda:
# docker-compose.yml
version: '3.8'

networks:
  monitoring:
    driver: bridge

volumes:
  prometheus_data: {}
  grafana_data: {}
  loki_data: {}

services:
  prometheus:
    image: prom/prometheus:v2.42.0
    container_name: prometheus
    volumes:
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
      - ./prometheus/rules:/etc/prometheus/rules
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--web.enable-lifecycle'
    ports:
      - "9090:9090"
    networks:
      - monitoring
    restart: unless-stopped

  alertmanager:
    image: prom/alertmanager:v0.25.0
    container_name: alertmanager
    volumes:
      - ./alertmanager/alertmanager.yml:/etc/alertmanager/alertmanager.yml
    command:
      - '--config.file=/etc/alertmanager/alertmanager.yml'
      - '--storage.path=/alertmanager'
    ports:
      - "9093:9093"
    networks:
      - monitoring
    restart: unless-stopped

  node_exporter:
    image: prom/node-exporter:v1.5.0
    container_name: node_exporter
    volumes:
      - /proc:/host/proc:ro
      - /sys:/host/sys:ro
      - /:/rootfs:ro
    command:
      - '--path.procfs=/host/proc'
      - '--path.sysfs=/host/sys'
      - '--collector.filesystem.mount-points-exclude=^/(sys|proc|dev|host|etc)($$|/)'
    ports:
      - "9100:9100"
    networks:
      - monitoring
    restart: unless-stopped

  loki:
    image: grafana/loki:2.8.0
    container_name: loki
    volumes:
      - ./loki/loki-config.yml:/etc/loki/loki-config.yml
      - loki_data:/loki
    command:
      - '-config.file=/etc/loki/loki-config.yml'
    ports:
      - "3100:3100"
    networks:
      - monitoring
    restart: unless-stopped

  promtail:
    image: grafana/promtail:2.8.0
    container_name: promtail
    volumes:
      - ./promtail/promtail-config.yml:/etc/promtail/promtail-config.yml
      - /var/log:/var/log
    command:
      - '-config.file=/etc/promtail/promtail-config.yml'
    networks:
      - monitoring
    restart: unless-stopped

  grafana:
    image: grafana/grafana:9.5.0
    container_name: grafana
    volumes:
      - grafana_data:/var/lib/grafana
      - ./grafana/provisioning:/etc/grafana/provisioning
      - ./grafana/dashboards:/var/lib/grafana/dashboards
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=secure_password
      - GF_USERS_ALLOW_SIGN_UP=false
      - GF_INSTALL_PLUGINS=grafana-piechart-panel,grafana-worldmap-panel
    ports:
      - "3000:3000"
    networks:
      - monitoring
    restart: unless-stopped
)

# Create configuration files for the monitoring stack
import os

# Create directories
directories = [
    'prometheus',
    'prometheus/rules',
    'alertmanager',
    'loki',
    'promtail',
    'grafana/provisioning/datasources',
    'grafana/provisioning/dashboards',
    'grafana/dashboards'
]

for directory in directories:
    os.makedirs(directory, exist_ok=True)

# Create docker-compose.yml
with open('docker-compose.yml', 'w') as f:
    f.write(create_monitoring_stack.result)

# Create prometheus.yml
prometheus_config = '''
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "rules/*.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
            - alertmanager:9093

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'node_exporter'
    static_configs:
      - targets: ['node_exporter:9100']

  - job_name: 'app'
    static_configs:
      - targets: ['app:8000']
'''

with open('prometheus/prometheus.yml', 'w') as f:
    f.write(prometheus_config)
```

## Deployment Scripts

Automate complex deployment scenarios:

```python
:do generate_kubernetes_manifests
Create Kubernetes manifests for a microservices application with:
1. Deployment for 3 distinct services
2. Service and Ingress resources for external access
3. ConfigMap for environment-specific configuration
4. Secrets management
:--

# After processing:
:do generate_kubernetes_manifests
Create Kubernetes manifests for a microservices application with:
1. Deployment for 3 distinct services
2. Service and Ingress resources for external access
3. ConfigMap for environment-specific configuration
4. Secrets management
:--(happy_panda:
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: microservices
---
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: microservices
data:
  API_ENDPOINT: "https://api.example.com"
  LOG_LEVEL: "info"
  ENVIRONMENT: "production"
  DATABASE_HOST: "db-service.microservices.svc.cluster.local"
  REDIS_HOST: "redis-service.microservices.svc.cluster.local"
---
# secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
  namespace: microservices
type: Opaque
data:
  DATABASE_PASSWORD: cGFzc3dvcmQxMjM= # password123 in base64
  API_KEY: c2VjcmV0LWFwaS1rZXkxMjM= # secret-api-key123 in base64
  JWT_SECRET: anN0LXNlY3JldC1rZXkxMjM= # jst-secret-key123 in base64
---
# auth-service.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
  namespace: microservices
spec:
  replicas: 2
  selector:
    matchLabels:
      app: auth-service
  template:
    metadata:
      labels:
        app: auth-service
    spec:
      containers:
      - name: auth-service
        image: yourcompany/auth-service:latest
        imagePullPolicy: Always
        ports:
        - containerPort: 8000
        resources:
          limits:
            cpu: "0.5"
            memory: "512Mi"
          requests:
            cpu: "0.2"
            memory: "256Mi"
        env:
        - name: API_ENDPOINT
          valueFrom:
            configMapKeyRef:
              name: app-config
              key: API_ENDPOINT
        - name: LOG_LEVEL
          valueFrom:
            configMapKeyRef:
              name: app-config
              key: LOG_LEVEL
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: JWT_SECRET
        readinessProbe:
          httpGet:
            path: /health
            port: 8000
          initialDelaySeconds: 10
          periodSeconds: 5
        livenessProbe:
          httpGet:
            path: /health
            port: 8000
          initialDelaySeconds: 15
          periodSeconds: 20
---
apiVersion: v1
kind: Service
metadata:
  name: auth-service
  namespace: microservices
spec:
  selector:
    app: auth-service
  ports:
  - port: 80
    targetPort: 8000
  type: ClusterIP
---
# api-service.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-service
  namespace: microservices
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api-service
  template:
    metadata:
      labels:
        app: api-service
    spec:
      containers:
      - name: api-service
        image: yourcompany/api-service:latest
        imagePullPolicy: Always
        ports:
        - containerPort: 8000
        resources:
          limits:
            cpu: "1"
            memory: "1Gi"
          requests:
            cpu: "0.5"
            memory: "512Mi"
        env:
        - name: DATABASE_HOST
          valueFrom:
            configMapKeyRef:
              name: app-config
              key: DATABASE_HOST
        - name: DATABASE_PASSWORD
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: DATABASE_PASSWORD
        - name: API_KEY
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: API_KEY
        - name: ENVIRONMENT
          valueFrom:
            configMapKeyRef:
              name: app-config
              key: ENVIRONMENT
        readinessProbe:
          httpGet:
            path: /health
            port: 8000
          initialDelaySeconds: 10
          periodSeconds: 5
        livenessProbe:
          httpGet:
            path: /health
            port: 8000
          initialDelaySeconds: 15
          periodSeconds: 20
---
apiVersion: v1
kind: Service
metadata:
  name: api-service
  namespace: microservices
spec:
  selector:
    app: api-service
  ports:
  - port: 80
    targetPort: 8000
  type: ClusterIP
---
# frontend-service.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend-service
  namespace: microservices
spec:
  replicas: 2
  selector:
    matchLabels:
      app: frontend-service
  template:
    metadata:
      labels:
        app: frontend-service
    spec:
      containers:
      - name: frontend-service
        image: yourcompany/frontend-service:latest
        imagePullPolicy: Always
        ports:
        - containerPort: 80
        resources:
          limits:
            cpu: "0.5"
            memory: "512Mi"
          requests:
            cpu: "0.2"
            memory: "256Mi"
        env:
        - name: API_ENDPOINT
          valueFrom:
            configMapKeyRef:
              name: app-config
              key: API_ENDPOINT
        - name: ENVIRONMENT
          valueFrom:
            configMapKeyRef:
              name: app-config
              key: ENVIRONMENT
        readinessProbe:
          httpGet:
            path: /health
            port: 80
          initialDelaySeconds: 10
          periodSeconds: 5
        livenessProbe:
          httpGet:
            path: /health
            port: 80
          initialDelaySeconds: 15
          periodSeconds: 20
---
apiVersion: v1
kind: Service
metadata:
  name: frontend-service
  namespace: microservices
spec:
  selector:
    app: frontend-service
  ports:
  - port: 80
    targetPort: 80
  type: ClusterIP
---
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: microservices-ingress
  namespace: microservices
  annotations:
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - api.example.com
    - auth.example.com
    - app.example.com
    secretName: microservices-tls
  rules:
  - host: api.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: api-service
            port:
              number: 80
  - host: auth.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: auth-service
            port:
              number: 80
  - host: app.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: frontend-service
            port:
              number: 80
)

# Create Kubernetes manifest files
k8s_files = {
    'namespace.yaml': generate_kubernetes_manifests.result.split('---')[0],
    'configmap.yaml': generate_kubernetes_manifests.result.split('---')[1],
    'secrets.yaml': generate_kubernetes_manifests.result.split('---')[2],
    'auth-service.yaml': generate_kubernetes_manifests.result.split('---')[3] + '\n---\n' + generate_kubernetes_manifests.result.split('---')[4],
    'api-service.yaml': generate_kubernetes_manifests.result.split('---')[5] + '\n---\n' + generate_kubernetes_manifests.result.split('---')[6],
    'frontend-service.yaml': generate_kubernetes_manifests.result.split('---')[7] + '\n---\n' + generate_kubernetes_manifests.result.split('---')[8],
    'ingress.yaml': generate_kubernetes_manifests.result.split('---')[9]
}

# Create k8s directory
os.makedirs('k8s', exist_ok=True)

# Write Kubernetes manifest files
for filename, content in k8s_files.items():
    with open(f'k8s/{filename}', 'w') as f:
        f.write(content)
```

## Benefits for DevOps

PML offers unique advantages for DevOps workflows:

1. **Infrastructure as Code Automation**: Generate complex infrastructure code in seconds
2. **Pipeline Generation**: Create complete CI/CD pipelines with best practices built-in
3. **Configuration Management**: Easily create and update service configurations
4. **Observability Solutions**: Set up comprehensive monitoring with minimal effort
5. **Deployment Automation**: Generate deployment scripts for various environments
6. **Documentation Generators**: Auto-document infrastructure and deployment processes

PML transforms how DevOps engineers work by:

- Eliminating repetitive boilerplate in infrastructure code
- Automating the creation of pipeline configurations
- Streamlining deployment processes
- Ensuring best practices in infrastructure setups
- Maintaining consistency across environments

By integrating PML into your DevOps workflow, you can focus more on architectural decisions and optimization rather than writing and maintaining complex configuration files.
