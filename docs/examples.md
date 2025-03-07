---
layout: default
title: Examples
nav_order: 7
description: "Practical examples of using HttpProbe for different API testing scenarios."
---

# Examples
{: .no_toc }

Explore practical examples of using HttpProbe for different API testing scenarios.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Basic API Tests

### Simple GET Request

```yaml
name: "Simple API Test"
description: "Basic GET request test"
variables:
  base_url:
    type: string
    value: "https://api.example.com"
suites:
  - name: "Basic Tests"
    cases:
      - title: "Get API Status"
        request:
          method: GET
          url: "${base_url}/status"
          headers:
            - key: Accept
              value: application/json
          body:
            type: json
            data: null
          assertions:
            status: 200
            body:
              "$.status": "ok"
```

### Simple POST Request

```yaml
name: "User Creation Test"
description: "Test user creation API"
variables:
  base_url:
    type: string
    value: "https://api.example.com"
suites:
  - name: "User Management"
    cases:
      - title: "Create New User"
        request:
          method: POST
          url: "${base_url}/users"
          headers:
            - key: Content-Type
              value: application/json
            - key: Accept
              value: application/json
          body:
            type: json
            data: |
              {
                "name": "John Doe",
                "email": "john@example.com",
                "password": "securePassword123"
              }
          assertions:
            status: 201
            body:
              "$.success": true
              "$.user.name": "John Doe"
              "$.user.email": "john@example.com"
```

## Authentication Tests

### JWT Authentication

```yaml
name: "Authentication API Tests"
description: "Test authentication endpoints"
variables:
  base_url:
    type: string
    value: "https://api.example.com"
  username:
    type: string
    value: "testuser"
  password:
    type: string
    value: "${env:TEST_PASSWORD}"
suites:
  - name: "Authentication"
    cases:
      - title: "Login and Get Token"
        request:
          method: POST
          url: "${base_url}/auth/login"
          headers:
            - key: Content-Type
              value: application/json
          body:
            type: json
            data: |
              {
                "username": "${username}",
                "password": "${password}"
              }
          assertions:
            status: 200
            body:
              "$.success": true
              "$.token": "contains eyJ"  # JWT tokens start with eyJ
              "$.expiresIn": "> 0"
```

### OAuth2 Authentication

```yaml
name: "OAuth2 Authentication Tests"
description: "Test OAuth2 flow"
variables:
  auth_url:
    type: string
    value: "https://auth.example.com"
  client_id:
    type: string
    value: "${env:OAUTH_CLIENT_ID}"
  client_secret:
    type: string
    value: "${env:OAUTH_CLIENT_SECRET}"
suites:
  - name: "OAuth Flow"
    cases:
      - title: "Get Access Token"
        request:
          method: POST
          url: "${auth_url}/oauth/token"
          headers:
            - key: Content-Type
              value: application/x-www-form-urlencoded
          body:
            type: json
            data: |
              {
                "grant_type": "client_credentials",
                "client_id": "${client_id}",
                "client_secret": "${client_secret}"
              }
          assertions:
            status: 200
            body:
              "$.access_token": "contains "
              "$.token_type": "Bearer"
              "$.expires_in": "> 0"
```

## CRUD API Tests

### Complete CRUD Example

```yaml
name: "Product API CRUD Tests"
description: "Test Create, Read, Update, Delete operations for products"
variables:
  base_url:
    type: string
    value: "https://api.example.com"
  auth_token:
    type: string
    value: "${env:AUTH_TOKEN}"
suites:
  - name: "Product CRUD"
    cases:
      - title: "Create Product"
        request:
          method: POST
          url: "${base_url}/products"
          headers:
            - key: Authorization
              value: Bearer ${auth_token}
            - key: Content-Type
              value: application/json
          body:
            type: json
            data: |
              {
                "name": "Test Product",
                "price": 99.99,
                "category": "Electronics",
                "inStock": true
              }
          assertions:
            status: 201
            body:
              "$.success": true
              "$.product.id": "length > 0"  # Check that ID was generated
              "$.product.name": "Test Product"
              
      - title: "Get Product"
        request:
          method: GET
          url: "${base_url}/products/{{product_id}}"  # This would come from the previous test
          headers:
            - key: Authorization
              value: Bearer ${auth_token}
          body:
            type: json
            data: null
          assertions:
            status: 200
            body:
              "$.name": "Test Product"
              "$.price": 99.99
              "$.category": "Electronics"
              "$.inStock": true
              
      - title: "Update Product"
        request:
          method: PUT
          url: "${base_url}/products/{{product_id}}" 
          headers:
            - key: Authorization
              value: Bearer ${auth_token}
            - key: Content-Type
              value: application/json
          body:
            type: json
            data: |
              {
                "name": "Updated Product",
                "price": 149.99
              }
          assertions:
            status: 200
            body:
              "$.success": true
              "$.product.name": "Updated Product"
              "$.product.price": 149.99
              "$.product.category": "Electronics"  # Unchanged field
              
      - title: "Delete Product"
        request:
          method: DELETE
          url: "${base_url}/products/{{product_id}}"
          headers:
            - key: Authorization
              value: Bearer ${auth_token}
          body:
            type: json
            data: null
          assertions:
            status: 200
            body:
              "$.success": true
              "$.message": "contains deleted"
```

## Advanced Features

### Environment-Specific Tests

```yaml
name: "Environment-Specific Tests"
description: "Tests that adapt to different environments"
variables:
  env_name:
    type: string
    value: "${env:TARGET_ENV}"  # dev, staging, prod
  base_url:
    type: string
    value: "${env:${env_name}_API_URL}"  # Uses dynamic environment variables
suites:
  - name: "Core API"
    cases:
      - title: "Check Environment"
        request:
          method: GET
          url: "${base_url}/info"
          headers:
            - key: Accept
              value: application/json
          body:
            type: json
            data: null
          assertions:
            status: 200
            body:
              "$.environment": "${env_name}"
```

### Dynamic Data Generation

```yaml
name: "Dynamic Data Tests"
description: "Tests using dynamically generated data"
variables:
  base_url:
    type: string
    value: "https://api.example.com"
  request_id:
    type: string
    value: "${random(16)}"  # 16-character random string
  timestamp:
    type: string
    value: "${timestamp(2006-01-02T15:04:05Z)}"  # Current ISO timestamp
suites:
  - name: "Logging API"
    cases:
      - title: "Log Event"
        request:
          method: POST
          url: "${base_url}/events"
          headers:
            - key: Content-Type
              value: application/json
            - key: X-Request-ID
              value: ${request_id}
          body:
            type: json
            data: |
              {
                "event": "TEST_EVENT",
                "timestamp": "${timestamp}",
                "data": {
                  "requestId": "${request_id}",
                  "source": "httpprobe_test"
                }
              }
          assertions:
            status: 201
            body:
              "$.success": true
```

### Schema Validation

```yaml
name: "Schema Validation Tests"
description: "Tests that validate response schema"
variables:
  base_url:
    type: string
    value: "https://api.example.com"
suites:
  - name: "User API"
    cases:
      - title: "Get User with Schema Validation"
        request:
          method: GET
          url: "${base_url}/users/123"
          headers:
            - key: Accept
              value: application/json
          body:
            type: json
            data: null
          assertions:
            status: 200
            schema: |
              {
                "type": "object",
                "required": ["id", "name", "email", "createdAt"],
                "properties": {
                  "id": { 
                    "type": "integer",
                    "minimum": 1
                  },
                  "name": { 
                    "type": "string",
                    "minLength": 1
                  },
                  "email": { 
                    "type": "string",
                    "format": "email"
                  },
                  "createdAt": {
                    "type": "string",
                    "format": "date-time"
                  },
                  "role": {
                    "type": "string",
                    "enum": ["admin", "user", "guest"]
                  },
                  "settings": {
                    "type": "object",
                    "properties": {
                      "notifications": { "type": "boolean" },
                      "theme": { "type": "string" }
                    }
                  }
                }
              }
```

## Integration Tests

### Microservices Integration

```yaml
name: "Order Processing Flow"
description: "Test the complete order flow across multiple services"
variables:
  auth_url:
    type: string
    value: "https://auth.example.com"
  order_url:
    type: string
    value: "https://orders.example.com"
  payment_url:
    type: string
    value: "https://payments.example.com"
  username:
    type: string
    value: "${env:TEST_USERNAME}"
  password:
    type: string
    value: "${env:TEST_PASSWORD}"
suites:
  - name: "E2E Order Flow"
    cases:
      - title: "1. Login"
        request:
          method: POST
          url: "${auth_url}/login"
          headers:
            - key: Content-Type
              value: application/json
          body:
            type: json
            data: |
              {
                "username": "${username}",
                "password": "${password}"
              }
          assertions:
            status: 200
            body:
              "$.token": "length > 0"
              
      - title: "2. Create Order"
        request:
          method: POST
          url: "${order_url}/orders"
          headers:
            - key: Authorization
              value: Bearer {{auth_token}}  # From previous test
            - key: Content-Type
              value: application/json
          body:
            type: json
            data: |
              {
                "items": [
                  {"productId": "prod-123", "quantity": 2},
                  {"productId": "prod-456", "quantity": 1}
                ],
                "shippingAddress": {
                  "street": "123 Main St",
                  "city": "Anytown",
                  "zipCode": "12345"
                }
              }
          assertions:
            status: 201
            body:
              "$.orderId": "length > 0"
              "$.total": "> 0"
              
      - title: "3. Process Payment"
        request:
          method: POST
          url: "${payment_url}/process"
          headers:
            - key: Authorization
              value: Bearer {{auth_token}}
            - key: Content-Type
              value: application/json
          body:
            type: json
            data: |
              {
                "orderId": "{{order_id}}",
                "paymentMethod": "credit_card",
                "cardDetails": {
                  "number": "4111111111111111",
                  "expiry": "12/25",
                  "cvv": "123"
                }
              }
          assertions:
            status: 200
            body:
              "$.success": true
              "$.transactionId": "length > 0"
              
      - title: "4. Check Order Status"
        request:
          method: GET
          url: "${order_url}/orders/{{order_id}}"
          headers:
            - key: Authorization
              value: Bearer {{auth_token}}
          body:
            type: json
            data: null
          assertions:
            status: 200
            body:
              "$.status": "paid"
              "$.paymentDetails.transactionId": "{{transaction_id}}"
```

These examples cover a wide range of API testing scenarios and showcase HttpProbe's features for creating comprehensive and maintainable test suites.