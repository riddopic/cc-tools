---
name: explanation-docs
description: Explanation documentation patterns for understanding-oriented content - conceptual guides that explain why things work the way they do
---

# Explanation Documentation

Explanations are understanding-oriented content for readers who want to know **why** things work the way they do. They are read away from the keyboard -- readers are building mental models, not accomplishing tasks.

**Explanations are NOT:** Tutorials (hands-on doing), How-To guides (specific goals), Reference docs (precise lookups).

## Explanation Document Template

```markdown
---
title: "[Concept/System Name] Explained"
description: "Understand how [concept] works and why it was designed this way"
---

# Understanding [Concept]

Brief intro (2-3 sentences): What this explains and why it matters.

## Overview

High-level summary. What is it? What problem does it solve?

## Background and Context

### The Problem
What situation or challenge led to this design?

### Historical Context
How did we get here? Why were alternatives rejected?

## How It Works

### Core Concepts
Fundamental ideas with analogies to connect to familiar concepts.
Use diagrams for complex relationships or flows.

### The Mechanism
Conceptual walkthrough of how the system operates (what happens, not what to do).

### Key Components
Major parts and their interactions. For each: role and relationships.

## Design Decisions and Trade-offs

### Why This Approach?
Reasoning behind key design choices and goals.

### Trade-offs Made
- What was prioritized
- What was sacrificed
- Under what conditions this design excels or struggles

### Constraints and Assumptions
What shaped and limits the design?

## Alternatives Considered
Brief description of alternatives, why not chosen, when they might be better.

## Implications and Consequences
Impact on performance, scalability, developer experience, extensibility.

## Related Concepts
Links to related topics and deeper references.
```

## Writing Principles

1. **Focus on understanding, not doing.** Answer "why?" and "how does it work?" -- not "how do I?"
2. **Use analogies and mental models.** Connect unfamiliar concepts to things readers already know before introducing technical details.
3. **Explain the "why" behind design decisions.** Don't just describe what exists -- explain why it exists that way, with rationale.
4. **Discuss trade-offs honestly.** Every design has costs. Be explicit about what was prioritized and sacrificed.
5. **Structure for reflection, not action.** Use flowing prose more than bullet points. Build concepts progressively. Allow for depth.
6. **Connect to the bigger picture.** Show how this concept relates to other parts of the system or broader patterns.

## Components

- **Diagrams/visuals**: Use mermaid or similar for architecture and flow diagrams
- **Comparison tables**: Compare approaches across consistent dimensions
- **Callouts**: `<Note>` for common confusion points, `<Warning>` for critical assumptions
- **Expandable sections**: For tangential but valuable historical context

## Checklist

- [ ] Title indicates this explains a concept (not a how-to)
- [ ] Introduction sets expectations for what reader will understand
- [ ] Background provides context and history
- [ ] Core concepts explained with analogies or mental models
- [ ] Design decisions include rationale, not just facts
- [ ] Trade-offs discussed honestly
- [ ] Alternatives mentioned and compared
- [ ] Related concepts linked
- [ ] Written for reading away from keyboard (no tasks to follow)
- [ ] Progressive structure builds understanding step by step

## When to Use Explanation vs Other Types

| Reader's Question | Doc Type |
|------------------|----------|
| "How do I do X?" | How-To Guide |
| "Teach me about X" | Tutorial |
| "What is the API for X?" | Reference |
| "Why does X work this way?" | **Explanation** |
| "What are the trade-offs of X?" | **Explanation** |

Write an explanation when users ask "why" questions, need to make architectural decisions, are evaluating fit for their use case, or need context before implementation.

## Related Skills

- **docs-style**: Core writing conventions and components
- **howto-docs**: How-To guide patterns for task-oriented content
- **reference-docs**: Reference documentation patterns for lookups
- **tutorial-docs**: Tutorial patterns for learning-oriented content
