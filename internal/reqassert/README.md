# Request Assertion Package

The `reqassert` package provides a flexible and extensible way to validate HTTP responses in httpprobe tests. It supports various assertion types and can be extended with custom assertions.

## Basic Usage

The package comes with built-in assertion types for validating:
- HTTP status codes
- Response headers
- Response body contents (using JSONPath)

In test definition files, you can use assertions as follows:

```yaml
assertions:
  status: 200                      # Status code assertion
  headers:
    Content-Type: application/json # Header assertion
    Cache-Control: no-cache
  body:
    $.id: 123                      # Body assertion with JSONPath
    $.name: "John Doe"
    $.items[0].id: 1
    $.count: "> 5"                 # Comparison operator
    $.tags: "contains admin"       # Contains function
```

## Extending with Custom Assertions

You can add custom assertion types by implementing the `Assertion` and `AssertionFactory` interfaces, and registering them with the assertion builder.

### Step 1: Define your assertion type

```go
// CustomAssertion validates something custom
type CustomAssertion struct {
    // Custom fields
    Key      string
    Expected interface{}
}

// Validate implements the Assertion interface
func (a *CustomAssertion) Validate(ctx *reqassert.AssertionContext) error {
    // Implement custom validation logic
    // Return nil if assertion passes, error otherwise
    return nil
}
```

### Step 2: Create a factory for your assertion

```go
// CustomAssertionFactory creates custom assertions
type CustomAssertionFactory struct{}

// Create implements the AssertionFactory interface
func (f *CustomAssertionFactory) Create(key string, expected interface{}) (reqassert.Assertion, error) {
    // Implement logic to create your custom assertion
    return &CustomAssertion{
        Key:      key,
        Expected: expected,
    }, nil
}
```

### Step 3: Register your assertion type

```go
// Create a builder
builder := reqassert.NewBuilder()

// Register your custom assertion type
builder.RegisterType("custom", &CustomAssertionFactory{})
```

### Step 4: Use in test definitions

```yaml
assertions:
  custom:
    somePath: expectedValue
```

## Example: Adding a Response Time Assertion

Here's an example of adding a custom assertion to validate response time:

```go
// ResponseTimeAssertion validates that the response was received within a time limit
type ResponseTimeAssertion struct {
    MaxMilliseconds int
}

func (a *ResponseTimeAssertion) Validate(ctx *reqassert.AssertionContext) error {
    if ctx.ResponseTimeMs > a.MaxMilliseconds {
        return fmt.Errorf("response time %dms exceeded maximum %dms", 
            ctx.ResponseTimeMs, a.MaxMilliseconds)
    }
    return nil
}

// ResponseTimeAssertionFactory creates response time assertions
type ResponseTimeAssertionFactory struct{}

func (f *ResponseTimeAssertionFactory) Create(key string, expected interface{}) (reqassert.Assertion, error) {
    // Convert expected to int
    maxTime, ok := expected.(int)
    if !ok {
        if floatVal, isFloat := expected.(float64); isFloat {
            maxTime = int(floatVal)
        } else {
            return nil, fmt.Errorf("response time must be an integer (milliseconds), got %T", expected)
        }
    }
    
    return &ResponseTimeAssertion{MaxMilliseconds: maxTime}, nil
}
```

Then register it:

```go
builder.RegisterType("responseTime", &ResponseTimeAssertionFactory{})
```

And use in test definitions:

```yaml
assertions:
  responseTime: 500  # Assert response time is under 500ms
  status: 200
  # other assertions...
```

## Available Assertion Types

| Type | Description | Example |
|------|-------------|---------|
| `status` | HTTP status code | `status: 200` |
| `headers` | HTTP response headers | `headers: { Content-Type: application/json }` |
| `body` | Response body content (JSONPath) | `body: { $.id: 123 }` |

## Comparison Operators

For body assertions, you can use the following comparison operators:

| Operator | Description | Example |
|----------|-------------|---------|
| equals, = | Equality (default) | `$.count: 5` |
| contains | String contains | `$.message: "contains success"` |
| >, gt | Greater than | `$.count: "> 5"` |
| >=, gte | Greater than or equal | `$.count: ">= 5"` |
| <, lt | Less than | `$.count: "< 5"` |
| <=, lte | Less than or equal | `$.count: "<= 5"` |