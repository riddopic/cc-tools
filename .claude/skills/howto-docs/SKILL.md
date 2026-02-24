---
name: howto-docs
description: How-To guide patterns for documentation - task-oriented guides for users with specific goals
---

# How-To Documentation

How-To guides are task-oriented content for users who have a specific goal. They assume some familiarity with the product and focus on practical, actionable steps.

**How-To guides are NOT:** Tutorials (teach through exploration), Explanations (provide understanding), Reference docs (describe the system).

## How-To Guide Template

```markdown
---
title: "How to [achieve specific goal]"
description: "Learn how to [goal] using [product/feature]"
---

# How to [Goal]

Brief intro (1-2 sentences): what you'll accomplish and why it's useful.

## Prerequisites

- [What user needs before starting]
- [Required access, tools, or setup]
- [Any prior knowledge assumed]

## Steps

### 1. [Action verb] the [thing]

[Clear instruction with expected outcome]

### 2. [Next action]

[Continue with clear, single-action steps]

### 3. [Continue numbering]

[Each step should be one discrete action]

## Verify it worked

[How to confirm success - what should user see/experience?]

## Troubleshooting

- **[Common issue 1]**: [Solution or workaround]
- **[Common issue 2]**: [Solution or workaround]

## Next steps

- [Related how-to guide 1]
- [Related how-to guide 2]
- [Deeper dive reference doc]
```

## Writing Principles

### Titles
Always start with "How to" + active verb. Be specific: "How to add SSO authentication" not "How to set up auth".

### Step Structure
1. **One action per step** -- if you write "and", consider splitting
2. **Start with action verbs**: Click, Navigate, Enter, Select, Run, Create
3. **Show expected outcomes** after key steps

### Minimize Context
Don't explain why things work -- just show how. Link to explanations for deeper understanding.

### User Perspective
Write from the user's perspective, not the product's:

| Avoid (product-centric) | Prefer (user-centric) |
|------------------------|----------------------|
| "The API accepts..." | "Send a request to..." |
| "The system will..." | "You'll see..." |
| "This feature allows..." | "You can now..." |

### Prerequisites
Be explicit. List required access, tools, versions, and prior setup steps with links.

## Callouts

Use callouts sparingly for critical information:
- `<Warning>` for destructive or irreversible actions
- `<Note>` for timing or important context
- `<Tip>` for optional shortcuts

Use expandable sections for optional advanced details that shouldn't interrupt flow.

## Checklist

- [ ] Title starts with "How to" and describes a specific goal
- [ ] Prerequisites section lists all requirements
- [ ] Each step is a single, clear action
- [ ] Action verbs start each step
- [ ] Expected outcomes shown after key steps
- [ ] Verification section explains how to confirm success
- [ ] Troubleshooting covers common issues
- [ ] Next steps link to related content
- [ ] No unnecessary explanations -- links to concepts instead
- [ ] Written from user perspective, not product perspective

## When to Use How-To vs Other Types

| User's mindset | Doc type |
|---------------|----------|
| "I want to learn" | Tutorial |
| "I want to do X" | **How-To** |
| "I want to understand" | Explanation |
| "I need to look up Y" | Reference |

## Standard Markdown

Use standard Markdown for documentation. MDX components (`<Steps>`, `<Accordion>`, etc.) shown in templates are illustrative -- adapt to standard Markdown or your platform's shortcodes.

## Related Skills

- **docs-style**: Core writing conventions and components
- **tutorial-docs**: Tutorial patterns for learning-oriented content
- **reference-docs**: Reference documentation patterns
- **explanation-docs**: Conceptual documentation patterns
