# Blockbusterr Web Templates

This directory contains all the HTML templates for the Blockbusterr web UI.

## File Structure

```
web/templates/
├── base.html                 # Base layout with navigation and scripts
├── index.html                # Configuration page
├── jobs.html                 # Jobs management page  
├── activity.html             # Activity log page
├── filters.html              # Content filters configuration
├── activity_table.html       # Activity table partial
├── STYLE_GUIDE.md           # Component patterns and styling standards
├── COMPONENT_SUMMARY.md     # Summary of improvements made
└── components/               # Reusable component templates
    ├── card.html            # Card components
    ├── button.html          # Button variants
    ├── input.html           # Form input components
    ├── badge.html           # Status badges
    ├── alert.html           # Alert/notification components
    └── icons.html           # Common SVG icons
```

## Technology Stack

- **Templating**: Go templates with Fiber rendering
- **HTMX**: For dynamic interactions without JavaScript frameworks
- **TailwindCSS**: Utility-first CSS via CDN
- **Icons**: Heroicons (via inline SVG)

## Design System

### Color Palette
- **Primary**: Blue (#3b82f6)
- **Background**: Slate-900 (#0f172a)
- **Cards**: Slate-800 (#1e293b)
- **Borders**: Slate-700 (#334155)
- **Text**: Slate-100 (#f1f5f9)

### Component Library

See [STYLE_GUIDE.md](./STYLE_GUIDE.md) for detailed component patterns and usage examples.

### Key Components

1. **Cards**: Consistent containers with rounded borders
2. **Buttons**: Primary, secondary, success, and danger variants
3. **Forms**: Styled inputs, selects, and checkboxes
4. **Badges**: Status indicators with color coding
5. **Alerts**: Info, success, warning, and error notifications

## Page Structure

Each page template follows this structure:

```go
{{define "pagename"}}
<div class="space-y-6">
  <!-- Page content with consistent spacing -->
</div>
{{end}}
```

Pages are rendered with the `base` layout:

```go
c.Render("pagename", fiber.Map{
  "Title": "Page Title",
  "Config": config,
}, "base")
```

## Styling Guidelines

### Spacing
- Page sections: `space-y-6` or `mb-6`
- Form fields: `space-y-4`
- Card padding: `p-6`
- Component gaps: `gap-2`, `gap-4`, or `gap-6`

### Responsiveness
- Mobile-first approach
- Use `md:` prefix for tablet (768px)
- Use `lg:` prefix for desktop (1024px)
- Grid layouts: `grid-cols-1 lg:grid-cols-2`

### Transitions
Apply smooth transitions for better UX:
- Color changes: `transition-colors duration-200`
- Transforms: `transition-transform duration-200`

### Focus States
Always include proper focus indicators:
```html
focus:outline-none focus:ring-2 focus:ring-blue-500
```

## HTMX Patterns

### Form Submission
```html
<form hx-post="/endpoint" hx-target="#message" hx-swap="innerHTML">
  <!-- Form fields -->
  <div id="message"></div>
</form>
```

### Dynamic Loading
```html
<div hx-get="/api/data" hx-trigger="load" hx-swap="innerHTML">
  Loading...
</div>
```

## JavaScript Functions

Global functions defined in `base.html`:

- `togglePassword(inputId, button)` - Toggle password visibility
- `showNotification(message, type)` - Display toast notifications
- `toggleAdvanced(jobName, button)` - Expand/collapse advanced settings
- `testConnection(service, button)` - Test API connections

## Best Practices

1. **Consistency**: Use the component patterns from STYLE_GUIDE.md
2. **Accessibility**: Include proper labels, ARIA attributes, and focus states
3. **Performance**: Minimize inline scripts, use HTMX for dynamic content
4. **Maintainability**: Follow DRY principles, extract common patterns
5. **Responsive**: Test on mobile, tablet, and desktop viewports

## Adding New Pages

1. Create template file: `web/templates/newpage.html`
2. Define template: `{{define "newpage"}} ... {{end}}`
3. Add route in `internal/rest/v1/routes/ui.go`
4. Follow existing styling patterns from STYLE_GUIDE.md
5. Test responsiveness and HTMX interactions

## Updating Styles

When modifying styles:
1. Update STYLE_GUIDE.md if introducing new patterns
2. Apply changes consistently across all pages
3. Test dark mode appearance (current default)
4. Verify form validations and error states

## Component Development

The `components/` directory contains reusable template definitions. These are included in `base.html` and can be used throughout other templates.

### Using Components

Components use Go template syntax and can accept parameters:

```go
{{template "button-primary" map 
  "Text" "Save Changes"
  "Type" "submit"
  "Class" "w-full"
}}
```

Note: Component usage may vary based on Go template limitations. For complex components, prefer direct HTML with consistent classes from STYLE_GUIDE.md.

## Development Workflow

1. Make changes to template files
2. Restart the application (templates are compiled at startup)
3. Test in browser
4. Verify HTMX interactions work correctly
5. Check browser console for errors

## Resources

- [Fiber Template Guide](https://docs.gofiber.io/guide/templates/)
- [HTMX Documentation](https://htmx.org/docs/)
- [Tailwind CSS](https://tailwindcss.com/docs)
- [Heroicons](https://heroicons.com/)
