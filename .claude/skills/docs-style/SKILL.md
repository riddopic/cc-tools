---
name: docs-style
description: Core technical documentation writing principles for voice, tone, structure, and LLM-friendly patterns. Use when writing or reviewing any documentation.
---

# Documentation Style Guide

Apply these principles when writing or reviewing documentation for clarity, consistency, and accessibility.

## Voice and Tone

1. **Use second person.** Address the reader as "you" -- not "the user" or "developers."
2. **Prefer active voice.** The subject performs the action: "Create a config file" not "A config file should be created."
3. **Be concise.** Every word earns its place. Cut filler.

| Instead of | Use |
|------------|-----|
| in order to | to |
| for the purpose of | to, for |
| in the event that | if |
| at this point in time | now |
| due to the fact that | because |
| it is necessary to | you must |
| is able to | can |
| make use of | use |

## Document Structure

- **Descriptive headings.** Tell readers exactly what the section contains. Avoid clever or vague titles.
- **Self-contained pages.** Assume readers land from search. State what the feature is, list prerequisites, provide full context.
- **Proper heading hierarchy.** h1 > h2 > h3, never skip levels.
- **Skimmable content.** Paragraphs of 3-4 sentences max. Bullet points for lists. Key information first (inverted pyramid).

### Semantic Markup

- **UI elements**: Bold (Click **Save**)
- **Code/commands/file paths**: Backticks (`npm install`, `/etc/config.yaml`)
- **Key terms on first use**: Bold or italics
- **Placeholders**: SCREAMING_CASE or angle brackets (`YOUR_API_KEY` or `<api-key>`)
- **Structured data**: Tables with consistent columns
- **Multiple related items**: Lists

## Consistency

- **One term per concept.** Pick a term and use it everywhere. Document terminology choices.
- **Consistent formatting.** Same formatting rules for similar content types across all pages.

| Concept | Use | Don't use |
|---------|-----|-----------|
| Authentication credential | API key | API token, secret key, access key |
| Configuration file | config file | settings file, preferences file |
| Command line | CLI | terminal, command prompt, shell |

## LLM-Friendly Patterns

- **State prerequisites explicitly.** List what users need before starting.
- **Define acronyms on first use.** Spell out the first time on each page.
- **Provide complete, runnable code examples.** Include all imports, realistic placeholder values, and expected output.
- **Write descriptive titles and meta descriptions.** Help search and LLM understanding.

## Pitfalls to Avoid

- **Product-centric language.** Orient around user goals, not product features. "Send emails" not "Email Service."
- **Obvious instructions.** Don't document self-explanatory UI actions. "Enter your webhook URL (must use HTTPS)" not "Click in the text field. Type your URL. Click Save."
- **Colloquialisms.** Hurt clarity and localization. "Significantly improves performance" not "game-changer."

## Quick Reference Checklist

- [ ] Using "you" instead of "the user"
- [ ] Active voice throughout
- [ ] No unnecessary words
- [ ] Headings are descriptive
- [ ] Page is self-contained
- [ ] Proper heading hierarchy
- [ ] One term per concept
- [ ] Prerequisites listed
- [ ] Acronyms defined
- [ ] Code examples are complete
- [ ] No product-centric language
- [ ] No colloquialisms

## Godoc Conventions

For Go package documentation:
- Package comments describe purpose and functionality
- Function docs start with the function name
- Use complete sentences with proper grammar
- Document parameters, return values, and errors

## Related Skills

- **howto-docs**: How-To guide patterns for task-oriented content
- **tutorial-docs**: Tutorial patterns for learning-oriented content
- **reference-docs**: Reference documentation patterns
- **explanation-docs**: Conceptual documentation patterns
