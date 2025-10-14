package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {
		postsCollection, err := app.FindCollectionByNameOrId("posts")
		if err != nil {
			return err
		}

		usersCollection, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		adminUser, err := app.FindFirstRecordByFilter(usersCollection, "role = 'admin'")
		if err != nil {
			return nil // Admin doesn't exist yet, skip seeding (will run again on next startup)
		}

		// Check if posts already exist to avoid duplicates
		existingPost, _ := app.FindFirstRecordByFilter(postsCollection, "slug = 'markdown-guide'")
		if existingPost != nil {
			return nil // Posts already seeded
		}

		now := types.NowDateTime()

		markdownGuidePost := core.NewRecord(postsCollection)
		markdownGuidePost.Set("title", "Writing Blog Posts with Markdown")
		markdownGuidePost.Set("slug", "markdown-guide")
		markdownGuidePost.Set("excerpt", "A comprehensive guide to using Markdown syntax for creating rich, formatted blog posts on the AVR NPO website.")
		markdownGuidePost.Set("content", `# Writing Blog Posts with Markdown

This guide will help you create beautifully formatted blog posts using Markdown syntax.

## What is Markdown?

Markdown is a lightweight markup language that lets you format text using simple, readable syntax. It's perfect for writing blog posts because it's easy to learn and produces clean, professional results.

## Text Formatting

### Bold and Italic

Make text **bold** by wrapping it in double asterisks or __double underscores__.

Make text *italic* by wrapping it in single asterisks or _single underscores_.

Combine them for ***bold and italic*** text.

### Strikethrough

Use strikethrough for ~~text you want to cross out~~.

## Headings

Create headings by starting a line with one or more # symbols:

# Heading 1
## Heading 2
### Heading 3
#### Heading 4
##### Heading 5
###### Heading 6

## Lists

### Unordered Lists

Create bullet lists with hyphens, asterisks, or plus signs:

- First item
- Second item
  - Nested item
  - Another nested item
- Third item

### Ordered Lists

Create numbered lists:

1. First step
2. Second step
   1. Sub-step A
   2. Sub-step B
3. Third step

### Task Lists

Track tasks with checkboxes:

- [x] Completed task
- [ ] Pending task
- [ ] Another pending task

## Links

Create links using this syntax: [Link text](https://example.com)

Example: Visit the [AVR NPO website](https://avrnpo.org) for more information.

You can also use direct links: <https://avrnpo.org>

## Images

Add images using this syntax: ![Alt text](image-url.jpg)

Example: ![AVR Logo](/assets/images/logo.avif)

## Blockquotes

Create blockquotes for emphasis:

> This is a blockquote.
> It can span multiple lines.
>
> And include multiple paragraphs.

> Nested quotes:
>> Are also supported
>>> At multiple levels

## Code

### Inline Code

Use backticks for inline code within sentences.

### Code Blocks

Use triple backticks for code blocks:

`+"```javascript\n"+`function greet(name) {
  console.log("Hello, " + name + "!");
}
`+"\n```\n\n"+"```python\n"+`def calculate_total(items):
    return sum(item.price for item in items)
`+"\n```\n\n"+"```go\n"+`package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
`+"\n```"+`

## Tables

Create tables using pipes and hyphens:

| Header 1 | Header 2 | Header 3 |
|----------|----------|----------|
| Row 1    | Data     | Data     |
| Row 2    | Data     | Data     |
| Row 3    | Data     | Data     |

### Aligned Tables

| Left | Center | Right |
|:-----|:------:|------:|
| Left | Center | Right |
| Aligned | Columns | Here |

## Horizontal Rules

Create dividers with three or more hyphens, asterisks, or underscores:

---

***

___

## Special Characters

Escape special characters with backslash:

\* Not a bullet point
\# Not a heading
\[Not a link\]

## Best Practices

1. **Use headings hierarchically** - Start with H2 (##) for main sections
2. **Add blank lines** - Separate paragraphs and sections for readability
3. **Preview before publishing** - Check formatting in the editor preview
4. **Keep it simple** - Use formatting to enhance, not distract
5. **Be consistent** - Stick to one style for bullets, emphasis, etc.

## Tips for Great Blog Posts

- **Start with a compelling introduction** to hook readers
- **Use headings** to break up long content
- **Add images** to illustrate points and maintain interest
- **Include code examples** when relevant
- **Link to related content** for deeper exploration
- **End with a clear conclusion** or call-to-action

---

**Ready to write?** Head to the admin panel to create your first post!

For more information, visit the [Markdown Guide](https://www.markdownguide.org/).`)
		markdownGuidePost.Set("published", false)
		markdownGuidePost.Set("author", adminUser.Id)
		markdownGuidePost.Set("created", now)
		markdownGuidePost.Set("updated", now)

		if err := app.Save(markdownGuidePost); err != nil {
			return err
		}

		announcementPost := core.NewRecord(postsCollection)
		announcementPost.Set("title", "Introducing Our New Blog and Recurring Donation Features")
		announcementPost.Set("slug", "new-features-announcement")
		announcementPost.Set("excerpt", "We're excited to announce two major improvements to our website: a new blog system for sharing updates and stories, and the ability to set up recurring monthly donations to support our mission.")
		announcementPost.Set("content", `# Exciting Updates to AVR NPO

We're thrilled to announce two major enhancements to our website that will help us better serve our community and mission.

## New Blog System

We've launched a brand new blog system that allows us to share:

- **Mission Updates** - Stay informed about our latest initiatives and accomplishments
- **Community Stories** - Read inspiring stories from veterans and their families
- **Event Announcements** - Learn about upcoming events and opportunities to get involved
- **Resource Guides** - Access helpful information and resources for veterans

Our blog posts support rich formatting with Markdown, allowing us to create engaging, well-structured content complete with images, links, code examples, and more.

### Why This Matters

Transparent communication is essential to our mission. This blog enables us to:

1. Keep our community informed in real-time
2. Share success stories and impact metrics
3. Provide educational resources
4. Build stronger connections with supporters

---

## Recurring Donation Support

We're equally excited to introduce **recurring monthly donations** through our secure Helcim payment integration.

### How It Works

Supporting AVR NPO is now easier than ever:

- **Choose Your Impact** - Select a monthly amount that works for your budget
- **Set It and Forget It** - Automated monthly contributions with no repeated effort
- **Cancel Anytime** - Full control over your subscription with easy management
- **Secure Processing** - Bank-level encryption and PCI compliance

### Why Recurring Donations Matter

Monthly donations provide:

- **Predictable funding** - Helps us plan long-term programs and initiatives
- **Sustained impact** - Continuous support compounds over time
- **Lower overhead** - Reduced transaction costs mean more goes to our mission
- **Community building** - Join a dedicated group of regular supporters

> "Even a small monthly contribution can make a tremendous difference when combined with others who share our commitment to serving veterans."

### Getting Started

Ready to make a difference? Visit our [donation page](/donate) to:

1. Choose between one-time or recurring monthly donations
2. Select your contribution amount
3. Complete the secure checkout process
4. Receive instant email confirmation and tax receipt

---

## Looking Forward

These improvements represent our commitment to transparency, accessibility, and sustainable growth. We're grateful for your continued support and excited to share our journey with you through this new platform.

**Questions?** Feel free to [contact us](/contact) - we'd love to hear from you!

### Stay Connected

- Subscribe to our blog for updates
- Follow us on social media
- Join our Discord community
- Set up a recurring donation to sustain our mission

Thank you for being part of the AVR NPO family. Together, we're making a lasting impact in the lives of veterans and their families.

---

*Published by the AVR NPO Team*`)
		announcementPost.Set("published", false)
		announcementPost.Set("author", adminUser.Id)
		announcementPost.Set("created", now)
		announcementPost.Set("updated", now)

		if err := app.Save(announcementPost); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		postsCollection, err := app.FindCollectionByNameOrId("posts")
		if err != nil {
			return err
		}

		slugs := []string{"markdown-guide", "new-features-announcement"}
		for _, slug := range slugs {
			record, err := app.FindFirstRecordByFilter(postsCollection, "slug = '"+slug+"'")
			if err != nil {
				continue
			}
			if err := app.Delete(record); err != nil {
				return err
			}
		}

		return nil
	})
}
