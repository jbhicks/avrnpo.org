This is a project using Go as an HTTP server, and Tailwind CSS along with Daisy UI components for UI. HTMX is used to provide interactivity, and will use Javascript only when absolutely not possible to accomplish something with HTMX. When making any new UI or changing existing UI, be sure to conform to the existing colors and styling and always prefer Daisy UI components over rolling your own or using straight Tailwind. Always ensure code is clean and review for potential security issues and non-standard styling.

Things to not do that you usually do when writing code:
1. Do not add too many comments that explain what you yourself is doing in the code. At best, sparsely populate comments throughout explaining the code only, not what you are doing.

Here are examples of things you've gotten wrong before that you should ensure never happen again:
1. You tried to correct a style issue by overwriting a tailwind class in a .css file instead of just removing that class from the HTML. This was extremely stupid. Never do that.