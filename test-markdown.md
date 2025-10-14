# Markdown Test Document

## Headers

# H1 Header
## H2 Header
### H3 Header
#### H4 Header
##### H5 Header
###### H6 Header

## Emphasis

*Italic text with asterisks*
_Italic text with underscores_

**Bold text with asterisks**
__Bold text with underscores__

***Bold and italic***
~~Strikethrough text~~

## Lists

### Unordered Lists

* Item 1
* Item 2
  * Nested item 2.1
  * Nested item 2.2
* Item 3

- Alternative syntax
- Works the same way
  - With nesting

### Ordered Lists

1. First item
2. Second item
   1. Nested item 2.1
   2. Nested item 2.2
3. Third item

### Task Lists

- [x] Completed task
- [ ] Incomplete task
- [ ] Another task

## Links

[Inline link](https://example.com)
[Link with title](https://example.com "Title text")
[Reference link][1]

[1]: https://example.com

## Images

![Alt text](https://via.placeholder.com/150)
![Alt text with title](https://via.placeholder.com/150 "Image title")

## Code

### Inline Code

Use `inline code` with backticks.

### Code Blocks

```
Plain code block
No syntax highlighting
```

```javascript
// JavaScript code block
function hello() {
  console.log("Hello, world!");
}
```

```go
// Go code block
func main() {
    fmt.Println("Hello, world!")
}
```

```python
# Python code block
def hello():
    print("Hello, world!")
```

## Blockquotes

> Single line blockquote

> Multi-line blockquote
> with multiple lines
> of text

> Nested blockquotes
>> Second level
>>> Third level

## Horizontal Rules

---

***

___

## Tables

| Header 1 | Header 2 | Header 3 |
|----------|----------|----------|
| Row 1 Col 1 | Row 1 Col 2 | Row 1 Col 3 |
| Row 2 Col 1 | Row 2 Col 2 | Row 2 Col 3 |

### Alignment

| Left | Center | Right |
|:-----|:------:|------:|
| L1   | C1     | R1    |
| L2   | C2     | R2    |

## Line Breaks

Line 1  
Line 2 (with two spaces before)

Line 3

Line 4 (with blank line between)

## Escape Characters

\* Not italic \*
\# Not a header
\[Not a link\](url)

## HTML (if supported)

<div>
  <p>HTML paragraph</p>
  <strong>HTML bold</strong>
  <em>HTML italic</em>
</div>

## Footnotes

Here's a sentence with a footnote[^1].

[^1]: This is the footnote content.

## Definition Lists

Term 1
: Definition 1

Term 2
: Definition 2a
: Definition 2b

## Abbreviations

The HTML specification is maintained by the W3C.

*[HTML]: Hyper Text Markup Language
*[W3C]: World Wide Web Consortium

## Math (if supported)

Inline math: $E = mc^2$

Block math:

$$
\sum_{i=1}^{n} x_i = x_1 + x_2 + \cdots + x_n
$$

## Emoji (GitHub-flavored)

:smile: :heart: :thumbsup: :rocket: :tada:

## Special Characters

Copyright &copy; 2024
Trademark &trade;
Registered &reg;
Less than &lt; Greater than &gt;
Ampersand &amp;
