# Comprehensive Plush Template Guide

This guide provides a comprehensive overview of the Plush templating language, including project-specific conventions and a detailed syntax reference.

## 🚨 CRITICAL: Buffalo Partial Naming Convention

**⚠️ COMMON GOTCHA - This causes recurring 500 errors:**

Buffalo automatically adds an underscore prefix to partial names. When calling partials:

**❌ WRONG:**
```html
<%= partial("pages/_donate_content.plush.html") %>
```

**✅ CORRECT:**
```html
<%= partial("pages/donate_content.plush.html") %>
```

**Why this happens:**
- Buffalo looks for `_donate_content.plush.html` when you call `partial("donate_content.plush.html")`
- If you add the underscore yourself, Buffalo looks for `__donate_content.plush.html` (double underscore)
- This results in "could not find template" errors

**Rule: Never include the underscore in partial() calls - Buffalo adds it automatically**

## Common Project Patterns

### Buffalo Context Variables
- `current_user` - contains the authenticated user object

### String Manipulation
- Use Plush's built-in helpers like `capitalize(string)` and `len(string)`.
- Avoid Go-style syntax like `string[0:1]` as Plush doesn't support this directly. For more complex manipulations, create a custom helper.

---

## Plush Syntax Reference

This section is based on the official Plush documentation.

### Usage

Plush allows for the embedding of dynamic code inside of your templates.

```erb
<!-- input -->
<p><%= "plush is great" %></p>

<!-- output -->
<p>plush is great</p>
```

#### Controlling Output

By using the `<%= %>` tags we tell Plush to dynamically render the inner content. If we were to change the example to use `<% %>` tags instead the inner content will be evaluated and executed, but not injected into the template:

```erb
<!-- input -->
<p><% "plush is great" %></p>

<!-- output -->
<p></p>
```

By using the `<% %>` tags we can create variables (and functions!) inside of templates to use later:

```erb
<!-- does not print output -->
<%
let h = {name: "mark"}
let greet = fn(n) {
  return "hi " + n
}
%>
<!-- prints output -->
<h1><%= greet(h["name"]) %></h1>
```

### Comments

You can add comments like this:

```erb
<%# This is a comment %>
```

You can also add line comments within a code section

```erb
<%
# this is a comment
not_a_comment()
%>
```

### If/Else Statements

The basic syntax of `if/else if/else` statements is as follows:

```erb
<%
if (true) {
  # do something
} else if (false) {
  # do something
} else {
  # do something else
}
%>
```

When using `if/else` statements to control output, remember to use the `<%= %>` tag to output the result of the statement:

```erb
<%= if (true) { %>
  <!-- some html here -->
<% } else { %>
  <!-- some other html here -->
<% } %>
```

#### Operators

Complex `if` statements can be built in Plush using "common" operators:

* `==` - checks equality of two expressions
* `!=` - checks that the two expressions are not equal
* `~=` - checks a string against a regular expression (`foo ~= "^fo"`)
* `<` - checks the left expression is less than the right expression
* `<=` - checks the left expression is less than or equal to the right expression
* `>` - checks the left expression is greater than the right expression
* `>=` - checks the left expression is greater than or equal to the right expression
* `&&` - requires both the left **and** right expression to be true
* `||` - requires either the left **or** right expression to be true

### For Loops

There are three different types that can be looped over: maps, arrays/slices, and iterators. The format for them all looks the same:

```erb
<%= for (key, value) in expression { %>
  <%= key %> <%= value %>
<% } %>
```

You can also `continue` to the next iteration of the loop:
```erb
for (i,v) in [1, 2, 3,4,5,6,7,8,9,10] {
  if (i > 0) {
    continue
  }
  return v
}
```

You can terminate the for loop with `break`:
```erb
for (i,v) in [1, 2, 3,4,5,6,7,8,9,10] {
  if (i > 5) {
    break
  }
  return v
}
```

### Maps

Maps in Plush will get translated to the Go type `map[string]interface{}` when used.

```erb
<% let h = {key: "value", "a number": 1, bool: true} %>
```

Accessing maps is just like access a JSON object:

```erb
<%= h["key"] %>
```

### Arrays

Arrays in Plush will get translated to the Go type `[]interface{}` when used.

```erb
<% let a = [1, 2, "three", "four", h] %>
```

### Custom Helpers

You can add custom helper functions to use in your templates.

```go
ctx := NewContext()

// one() #=> 1
ctx.Set("one", func() int {
  return 1
})

// greet("mark") #=> "Hi mark"
ctx.Set("greet", func(s string) string {
  return fmt.Sprintf("Hi %s", s)
})

// can("update") #=> returns the block associated with it
ctx.Set("can", func(s string, help HelperContext) (template.HTML, error) {
  if s == "update" {
    h, err := help.Block()
    return template.HTML(h), err
  }
  return "", nil
})
```

```erb
<p><%= one() %></p>
<p><%= greet("mark")%></p>
<%= can("update") { %>
<p>i can update</p>
<% } %>
```
