# Markdown Test Document

This document contains examples of all standard Markdown syntax to verify proper rendering.

## Headings

# Heading 1
## Heading 2
### Heading 3
#### Heading 4
##### Heading 5
###### Heading 6

## Text Formatting

**Bold text** or __bold text__

*Italic text* or _italic text_

***Bold and italic*** or ___bold and italic___

~~Strikethrough text~~

## Lists

### Unordered Lists

- Item 1
- Item 2
  - Nested item 2.1
  - Nested item 2.2
    - Deeply nested item 2.2.1
- Item 3

* Alternative bullet style
* Another item
  * Nested with asterisk

### Ordered Lists

1. First item
2. Second item
   1. Nested item 2.1
   2. Nested item 2.2
3. Third item

### Task Lists

- [x] Completed task
- [ ] Incomplete task
- [ ] Another incomplete task

## Links

[Link text](https://example.com)

[Link with title](https://example.com "Link Title")

<https://autolink.com>

## Images

![Alt text](https://via.placeholder.com/150)

![Alt text with title](https://via.placeholder.com/150 "Image Title")

## Blockquotes

> This is a blockquote.
> It can span multiple lines.
>
> It can also contain multiple paragraphs.

> Nested blockquotes:
>> This is nested level 2
>>> This is nested level 3

## Code

### Inline Code

This is `inline code` within a sentence.

Use the `printf()` function to print text.

### Code Blocks

```
Plain code block without syntax highlighting
const x = 10;
```

```javascript
// JavaScript code block
function greet(name) {
  console.log(`Hello, ${name}!`);
}
greet('World');
```

```python
# Python code block
def factorial(n):
    if n <= 1:
        return 1
    return n * factorial(n - 1)

print(factorial(5))
```

```go
// Go code block
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```

```html
<!-- HTML code block -->
<!DOCTYPE html>
<html>
<head>
  <title>Page Title</title>
</head>
<body>
  <h1>Hello World</h1>
</body>
</html>
```

```css
/* CSS code block */
body {
  font-family: Arial, sans-serif;
  color: #333;
  background-color: #f5f5f5;
}
```

## Horizontal Rules

---

***

___

## Tables

| Header 1 | Header 2 | Header 3 |
|----------|----------|----------|
| Row 1 Col 1 | Row 1 Col 2 | Row 1 Col 3 |
| Row 2 Col 1 | Row 2 Col 2 | Row 2 Col 3 |
| Row 3 Col 1 | Row 3 Col 2 | Row 3 Col 3 |

### Table with Alignment

| Left Aligned | Center Aligned | Right Aligned |
|:-------------|:--------------:|--------------:|
| Left         | Center         | Right         |
| Text         | Text           | Text          |

## HTML Elements

You can use <strong>HTML tags</strong> in Markdown.

<div style="color: blue;">This is a blue div.</div>

<details>
<summary>Click to expand</summary>

This is hidden content that appears when you click the summary.

</details>

## Special Characters & Escaping

\* Escaped asterisk

\_ Escaped underscore

\# Escaped hash

\\ Backslash

## Footnotes

Here's a sentence with a footnote[^1].

Here's another with a longer note[^longnote].

[^1]: This is the first footnote.

[^longnote]: This is a longer footnote with multiple paragraphs.

    You can have multiple paragraphs in a footnote.

## Definitions

Term 1
: Definition 1

Term 2
: Definition 2a
: Definition 2b

## Emphasis in Different Contexts

This is **bold in a sentence** and this is *italic*.

**Bold at start** of line.

*Italic at start* of line.

This has **bold** and *italic* and `code` all together.

## Line Breaks

This is line one.  
This is line two (two spaces at end of previous line).

This is line one.

This is line two (blank line between).

## URLs and Email

Contact us at <admin@avrnpo.org>

Visit our website at <https://avrnpo.org>

## Emoji (if supported)

:smile: :heart: :thumbsup: :rocket: :fire:

😀 ❤️ 👍 🚀 🔥

## Mixed Content Example

### Real-World Blog Post Structure

**Introduction:** This is an example of how a real blog post might look with *various* formatting options.

Here's what we'll cover:

1. Main concepts
2. Implementation details
3. Code examples
4. Conclusion

#### Code Example with Explanation

The following code demonstrates a simple function:

```javascript
function calculateTotal(items) {
  return items.reduce((sum, item) => sum + item.price, 0);
}
```

> **Note:** This function uses the `reduce` method to sum prices.

#### Results Table

| Test Case | Input | Expected | Actual | Pass |
|-----------|-------|----------|--------|------|
| Test 1    | [1,2,3] | 6      | 6      | ✓    |
| Test 2    | []      | 0      | 0      | ✓    |
| Test 3    | [10]    | 10     | 10     | ✓    |

---

**Conclusion:** This document demonstrates proper Markdown rendering.

For more information, visit [Markdown Guide](https://www.markdownguide.org/).
