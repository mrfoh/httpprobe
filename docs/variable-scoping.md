# Variable Scoping

HttpProbe implements a hierarchical variable scoping system that provides isolation and flexibility when defining and using variables in your tests.

## Variable Scope Hierarchy

### Test Definition Scope
Variables defined at the root level of a test definition file are accessible to all test suites within that file.

```yaml
name: "My Test Definition"
variables:
  base_url:
    type: string
    value: "https://api.example.com"
  api_key:
    type: string
    value: "${env:API_KEY}"

suites:
  # All suites have access to base_url and api_key
```

### Test Suite Scope
Variables defined at the suite level are only accessible within that specific suite. These variables can override definition-level variables with the same name.

```yaml
suites:
  - name: "Test Suite"
    variables:
      base_url:  # Overrides the definition-level base_url
        type: string
        value: "https://staging.example.com"
    cases:
      # All cases in this suite use the staging URL
```

### Exported Variables Scope
Variables exported from response bodies using the `export` section are scoped to the test suite where they were created.

```yaml
suites:
  - name: "Authentication"
    cases:
      - title: "Login"
        request:
          # Login request
          export:
            body:
              - path: "$.token"
                as: "access_token"
      
      - title: "Use Token"
        request:
          # Can use access_token here
          
  - name: "Another Suite"
    cases:
      # access_token is NOT available here
```

## Variable Override Rules

When a variable name exists at multiple levels, the following precedence rules apply:

1. Suite-level variables override definition-level variables with the same name
2. Exported variables can override both suite-level and definition-level variables
3. Environment variables always take precedence when referenced directly with `${env:VAR_NAME}`

## Variable Naming Best Practices

1. **Use namespacing in your variable names** to prevent conflicts:
   ```yaml
   variables:
     auth_token: "..."    # Auth-related variable
     user_id: "..."       # User-related variable
     product_sku: "..."   # Product-related variable
   ```

2. **Use consistent prefixes** for variables that serve similar purposes:
   ```yaml
   variables:
     api_url: "..."
     api_version: "..."
     api_key: "..."
   ```

3. **Use suite-level variables** for values that might change between environments or test runs:
   ```yaml
   suites:
     - name: "Production Tests"
       variables:
         environment: "production"
     
     - name: "Staging Tests"
       variables:
         environment: "staging"
   ```

## Example: Complete Variable Scoping

```yaml
name: "Variable Scoping Example"
variables:
  base_url:
    type: string
    value: "https://api.example.com"
  common_var:
    type: string
    value: "global-value"
  overridden_var:
    type: string
    value: "global-version"

suites:
  - name: "First Suite"
    variables:
      suite_var:
        type: string
        value: "first-suite-value"
      overridden_var:
        type: string
        value: "first-suite-version"  # Overrides the global version
    cases:
      - title: "First Test"
        request:
          method: GET
          url: "${base_url}/test?var=${overridden_var}"
          # overridden_var resolves to "first-suite-version" here
          export:
            body:
              - path: "$.id"
                as: "resource_id"
                
      - title: "Second Test"
        request:
          method: GET
          url: "${base_url}/resources/${resource_id}"
          # Can use resource_id exported from the first test
          
  - name: "Second Suite"
    variables:
      suite_var:
        type: string
        value: "second-suite-value"
      overridden_var:
        type: string
        value: "second-suite-version"  # Different override
    cases:
      - title: "Third Test"
        request:
          method: GET
          url: "${base_url}/test?var=${overridden_var}"
          # overridden_var resolves to "second-suite-version" here
          # resource_id from First Suite is not available here
```

## Common Pitfalls

1. **Assuming variables are shared between suites**: Variables exported in one suite are not available in other suites.
2. **Forgetting variable override rules**: Suite-level variables override definition-level variables with the same name.
3. **Not providing fallbacks**: If a variable might not exist, consider providing fallback values or checking for existence.