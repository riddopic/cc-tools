# Core TDD Principles for Go Development

**TEST-DRIVEN DEVELOPMENT IS NON-NEGOTIABLE.** Every single line of production code must be written in response to a failing test. No exceptions. This is not a suggestion or a preference - it is the fundamental practice that enables all other principles in this codebase.

> "The best code is no code. The second best code is code that already exists and works."

## TDD is About Design, Not Testing

Test-Driven Development is fundamentally a **design methodology**, not a testing strategy. The tests are a beneficial side effect of the design process:

1. **Design First**: Writing a test first forces you to design the API before implementation
2. **Small Steps**: The Red-Green-Refactor cycle ensures incremental, manageable changes
3. **Emergent Architecture**: Better designs emerge from the pressure of testability
4. **Immediate Feedback**: Design flaws become apparent immediately when writing tests
5. **YAGNI Enforcement**: You only write code that's actually needed (tested)

## The Sacred TDD Cycle

### Red-Green-Refactor

```text
┌─────────────┐
│     RED     │ Write a failing test that describes desired behavior
└──────┬──────┘
       │
       ▼
┌─────────────┐
│    GREEN    │ Write MINIMUM code to make the test pass
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  REFACTOR   │ Assess if code can be improved (not always needed)
└─────────────┘
```

### Critical Rules

**1. No Production Code Without a Failing Test**

```go
// ❌ NEVER write this without a test demanding it
func CalculateDiscount(price float64, tier string) float64 {
    if tier == "premium" {
        return price * 0.8
    }
    if tier == "gold" {
        return price * 0.7
    }
    return price
}

// ✅ ALWAYS start with a test
func TestCalculateDiscount(t *testing.T) {
    t.Run("should apply 20% discount for premium tier", func(t *testing.T) {
        result := CalculateDiscount(100.0, "premium")
        assert.Equal(t, 80.0, result)
    })
}
// NOW write the minimum code to pass
```

**2. Write the Minimum Code to Pass**

```go
// Test demands: "should return user by id"
func TestFindUserByID(t *testing.T) {
    t.Run("should return user by id", func(t *testing.T) {
        user, err := FindUserByID("123")
        assert.NoError(t, err)
        assert.Equal(t, "123", user.ID)
    })
}

// ❌ Don't add features the test doesn't demand
func FindUserByID(id string) (*User, error) {
    user, err := db.Query("SELECT * FROM users WHERE id = ?", id)
    if err != nil {
        return nil, err
    }

    // DON'T add these without tests demanding them!
    if user == nil {
        return nil, ErrUserNotFound
    }
    if !user.IsActive {
        return nil, ErrInactiveUser
    }

    return user, nil
}

// ✅ Minimum code to pass
func FindUserByID(id string) (*User, error) {
    return &User{ID: id}, nil // If this passes the test, it's enough for now!
}
```

**3. Refactor Only When It Adds Value**

```go
// After getting tests green
func ProcessPayment(amount float64) (*ProcessedPayment, error) {
    if amount <= 0 {
        return nil, errors.New("invalid amount")
    }
    return &ProcessedPayment{
        ID:     generateID(),
        Amount: amount,
        Status: "processed",
    }, nil
}

// Assess: Is this code...
// - Clear? ✓
// - Simple? ✓
// - Maintainable? ✓
// Then DON'T refactor just to refactor!
```

## Why TDD Matters

### Forces Good Design

Testable code is inherently better designed:

- Loosely coupled (easy to mock dependencies)
- Highly cohesive (single responsibility)
- Clear interfaces (tests define the API)

### Documents Intent

```go
func TestOrderProcessing(t *testing.T) {
    t.Run("should apply free shipping for orders over $50", func(t *testing.T) {
        order := &Order{Items: []Item{{Price: 60.0}}}
        processed := ProcessOrder(order)
        assert.Equal(t, 0.0, processed.Shipping)
    })

    t.Run("should charge $5.99 shipping for orders under $50", func(t *testing.T) {
        order := &Order{Items: []Item{{Price: 30.0}}}
        processed := ProcessOrder(order)
        assert.Equal(t, 5.99, processed.Shipping)
    })
}
// The tests ARE the documentation!
```

### Enables Refactoring

With comprehensive tests, you can refactor with confidence:

```go
// Original implementation
func CalculateTotal(items []Item) float64 {
    var total float64
    for _, item := range items {
        total += item.Price * float64(item.Quantity)
    }
    return total
}

// Refactored with confidence (tests still pass!)
func CalculateTotal(items []Item) float64 {
    return lo.SumBy(items, func(item Item) float64 {
        return item.Price * float64(item.Quantity)
    })
}
```

### Reduces Debugging

Bugs are caught at the moment of creation:

```go
// Writing test first catches the bug immediately
func TestCalculateAverage(t *testing.T) {
    t.Run("should handle empty slice", func(t *testing.T) {
        result := CalculateAverage([]float64{})
        assert.Equal(t, 0.0, result) // Forces you to handle edge case
    })
}

// Without TDD, this bug ships to production
func CalculateAverage(numbers []float64) float64 {
    if len(numbers) == 0 {
        return 0 // Test forced us to handle this case!
    }
    return lo.Sum(numbers) / float64(len(numbers))
}
```

### Improves Focus

You only solve the problem at hand:

```go
// Test defines exact requirement
func TestFormatCurrency(t *testing.T) {
    t.Run("should format currency with 2 decimal places", func(t *testing.T) {
        assert.Equal(t, "$10.00", FormatCurrency(10.0))
        assert.Equal(t, "$10.50", FormatCurrency(10.5))
    })
}

// You won't accidentally add:
// - Locale support (not tested)
// - Currency symbols (not tested)
// - Thousand separators (not tested)
// Just what's needed!
```

## Common TDD Violations to Avoid

### 1. Writing Code "While You're There"

```go
// ❌ VIOLATION: Adding untested functionality
func CreateUser(data CreateUserDto) (*User, error) {
    user := &User{
        ID:   generateID(),
        Name: data.Name,
        Email: data.Email,
    }

    // "While I'm here, let me add this useful feature..."
    if strings.Contains(user.Email, "admin") {
        user.Role = "admin" // NO TEST DEMANDED THIS!
    }

    return user, nil
}
```

### 2. Writing Multiple Tests Before Going Green

```go
// ❌ VIOLATION: Writing test batch
func TestUserService(t *testing.T) {
    t.Run("should create user", func(t *testing.T) { /* pending */ })
    t.Run("should validate email", func(t *testing.T) { /* pending */ })
    t.Run("should hash password", func(t *testing.T) { /* pending */ })
    t.Run("should send welcome email", func(t *testing.T) { /* pending */ })
    // STOP! Make the first one pass before writing more!
}
```

### 3. Skipping Refactor Assessment

```go
// Got to green with messy code
func ProcessData(data interface{}) interface{} {
    // 50 lines of nested ifs and complex logic
}

// ❌ VIOLATION: Moving to next test without assessing refactoring
// ✅ CORRECT: Stop, assess, refactor if it adds value, commit, then continue
```

## The TDD Mindset

### Think Like a User

```go
// ❌ Implementation-focused thinking
t.Run("should call database.Save() method", func(t *testing.T) { /* ... */ })
t.Run("should instantiate EmailService", func(t *testing.T) { /* ... */ })

// ✅ Behavior-focused thinking
t.Run("should persist user data", func(t *testing.T) { /* ... */ })
t.Run("should send welcome email to new users", func(t *testing.T) { /* ... */ })
```

### Embrace Failure

```go
// Red phase is SUCCESS, not failure!
// Seeing this fail proves your test works:
func TestCalculateTax(t *testing.T) {
    result := CalculateTax(100.0)
    assert.Equal(t, 8.0, result)
}
// Error: CalculateTax undefined ← THIS IS GOOD!
```

### Small Steps Win

```go
// ❌ Trying to do too much at once
t.Run("should process order with discounts, tax, shipping, and inventory update", func(t *testing.T) { /* ... */ })

// ✅ One small step at a time
t.Run("should calculate subtotal", func(t *testing.T) { /* ... */ })
// Make it pass, commit

t.Run("should apply percentage discount", func(t *testing.T) { /* ... */ })
// Make it pass, commit

t.Run("should calculate tax on discounted total", func(t *testing.T) { /* ... */ })
// Make it pass, commit
```

## TDD in Practice

### Starting a New Feature

```go
// 1. Start with the simplest test
func TestPasswordStrengthChecker(t *testing.T) {
    t.Run("should reject passwords under 8 characters", func(t *testing.T) {
        result := IsStrongPassword("short")
        assert.False(t, result)
    })
}

// 2. Write minimum code
func IsStrongPassword(password string) bool {
    return len(password) >= 8
}

// 3. Add next test
t.Run("should accept passwords with 8+ characters", func(t *testing.T) {
    result := IsStrongPassword("12345678")
    assert.True(t, result)
})

// 4. Code already passes! Add more specific test
t.Run("should require at least one uppercase letter", func(t *testing.T) {
    assert.False(t, IsStrongPassword("lowercase1"))
    assert.True(t, IsStrongPassword("Uppercase1"))
})

// 5. Enhance implementation
func IsStrongPassword(password string) bool {
    if len(password) < 8 {
        return false
    }
    hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
    return hasUpper
}
```

### The Power of Incremental Design

```go
// Test 1: Basic functionality
func TestCart(t *testing.T) {
    t.Run("should add item to cart", func(t *testing.T) {
        cart := NewCart()
        cart.Add(Item{ID: "1", Price: 10.0})
        assert.Equal(t, 1, cart.ItemCount())
    })
}

// Minimal implementation
type Cart struct {
    items []Item
}

func NewCart() *Cart {
    return &Cart{}
}

func (c *Cart) Add(item Item) {
    c.items = append(c.items, item)
}

func (c *Cart) ItemCount() int {
    return len(c.items)
}

// Test 2: New requirement emerges
t.Run("should calculate total price", func(t *testing.T) {
    cart := NewCart()
    cart.Add(Item{ID: "1", Price: 10.0})
    cart.Add(Item{ID: "2", Price: 20.0})
    assert.Equal(t, 30.0, cart.TotalPrice())
})

// Add just what's needed
func (c *Cart) TotalPrice() float64 {
    var total float64
    for _, item := range c.items {
        total += item.Price
    }
    return total
}

// Design emerges through tests!
```

## StatusLine-Specific TDD Examples

### Terminal Display Component

```go
func TestTerminalDisplay(t *testing.T) {
    t.Run("should render status line to terminal", func(t *testing.T) {
        display := NewTerminalDisplay()
        status := &Status{
            SessionID: "abc123",
            Active:    true,
            Theme:     "powerline",
        }

        output := display.Render(status)
        assert.Contains(t, output, "abc123")
        assert.Contains(t, output, "●") // Active indicator
    })
}
```

### Configuration Loading

```go
func TestConfigLoader(t *testing.T) {
    t.Run("should load default configuration", func(t *testing.T) {
        loader := NewConfigLoader()
        config, err := loader.Load()

        assert.NoError(t, err)
        assert.Equal(t, "powerline", config.Theme)
        assert.Equal(t, time.Second, config.RefreshRate)
    })
}
```

## Remember

**If you find yourself writing production code without a failing test, STOP immediately and write the test first.**

The discipline of TDD is what separates professional code from amateur code. It's not about the tests - it's about the design process that tests force upon us. Every keystroke of production code should be justified by a failing test.

This is the way.
