Plush Template Guide

This document provides the Plush usage guide and project conventions for writing and modifying Plush templates in this repository.

1) Purpose
- Central reference for Plush syntax, helpers, partials, and examples.
- Must be referenced by any agent or developer when creating or modifying templates.

2) Location
- File path: docs/plush/PLUSH_GUIDE.md
- All templates live under the project's templates directory (follow project conventions).

3) Key Rules (agents must follow)
- Always open and read this file before creating or modifying any template files.
- When creating a new template or editing an existing one, add (or update) a brief note in the top of this file under the "Change log" section indicating: file path, author, and a one-line description of the change.
- Partials: name partial files with a leading underscore: e.g. templates/shared/_header.plush.html and reference them using partial("shared/header").
- Template files must use the .plush.html extension.
- Use <%= %> to render expressions and <% %> for evaluations that produce no output.
- Use <%# comment %> for template comments.

4) Helper and Context Conventions
- Register helpers in Go via ctx.Set("name", fn) and document the helper's contract in this file.
- Block helpers should accept plush.HelperContext and return (template.HTML, error).

5) Security
- When returning raw HTML from a helper, ensure it's sanitized and returned as template.HTML.

6) Change log
- Any template change should add an entry here. Example:

- 2025-08-30 | opencode (agent) | templates/shared/_footer.plush.html - added site footer partial

7) Verification requirement for agents
- Agents updating templates must:
  - Read this file at start.
  - Append a changelog entry as described above.
  - Run any available template-related tests (if the repo includes them) and report results.

8) Examples (short)
- Output expression: <%= user["name"] %>
- Silent eval: <% let x = 1 %>
- For loop: <%= for (i, v) in items { %> ... <% } %>

---
Generated and enforced by repository agent guidance.
