# Vingo Templates

Vingo uses template files to render views in your project.  
By default, template files can be written in two different extensions:

- **`.html`** → Plain HTML templates. These files are treated just like normal HTML, but can also include Vingo directives and placeholders.
- **`.vgo`** → Vingo’s custom template extension. These are designed for projects that want to clearly separate framework templates from regular static HTML files.

👉 Both `.html` and `.vgo` templates are compiled and processed by Vingo in the same way. The choice of extension depends on whether you prefer **compatibility** (`.html`) or **clarity** (`.vgo`).

## VSCode Configuration

To make `.vgo` and `.vingo` files open as HTML inside VSCode, add the following configuration:

Create (or edit) a file at:

```
.vscode/settings.json
```

and add:

```json
{
  "files.associations": {
    "*.vgo": "html",
    "*.vingo": "html"
  }
}
```

Now VSCode will:  
- Show syntax highlighting for `.vgo` and `.vingo` files as if they are HTML.  
- Use the same icons and formatting rules as `.html` files.

## Example Template Folder Structure

```
/templates
  ├─ home.vgo
  ├─ about.html
  └─ layout.vgo
```

- `.vgo` files: Vingo custom templates  
- `.html` files: Standard HTML templates

You can now edit your templates in VSCode with full HTML syntax support, even for `.vgo` files.
