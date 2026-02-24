---
name: tutorial-docs
description: Tutorial patterns for documentation - learning-oriented guides that teach through guided doing
---

# Tutorial Documentation Skill

Patterns for writing effective tutorials following the Diataxis framework. Tutorials are learning-oriented content where the reader learns by doing under the guidance of a teacher.

## Core Principles

1. **Learn by Doing** — Tutorials teach through action, not explanation. The reader should be doing something at every moment.
2. **Visible Results at Every Step** — After each action, tell readers exactly what they should see. This confirms success and builds confidence.
3. **One Clear Path** — No alternatives. Pick one way and guide the reader through it completely.
4. **Teacher Takes Responsibility** — If the reader fails, the tutorial failed. Anticipate problems and prevent them.
5. **Permit Repetition** — Repeating similar actions in different contexts cements learning. Don't optimize for brevity.

## Tutorial Template

```markdown
---
title: "Build your first [thing]"
description: "Learn the basics of [product] by building a working [thing]"
---

# Build Your First [Thing]

In this tutorial, you'll build a [concrete deliverable]. By the end, you'll have a working [thing] that [does something visible].

## What you'll build

[Screenshot or diagram of the end result]

## Prerequisites

- [Minimal requirement 1 - link to install guide if needed]
- [Minimal requirement 2]

## Step 1: [Set up your project]

[First action - always start with something that produces visible output]

```bash
[command]
```

You should see:

```
[expected output]
```

## Step 2: [Create your first thing]

[Next action with clear instruction]

```code
[code to add or modify]
```

Save the file. You should see [visible change].

## Step 3-N: [Continue building]

[Each step produces visible output]

## What you've learned

- [Concrete skill 1]
- [Concrete skill 2]

## Next steps

- **[Tutorial 2]** - Continue learning by [next goal]
- **[How-to guide]** - Learn how to [specific task]
- **[Concepts page]** - Understand [concept] in depth
```

## Writing Principles

- **Titles**: Start with action outcomes — "Build your first...", "Create a...", "Deploy your..."
- **Steps**: Lead with the action, show exactly what to type, confirm success after every step, one visible change per step
- **Prerequisites**: Minimize them — tutorials are for beginners
- **Errors**: Anticipate failures and guide readers back on track with `<Warning>` callouts

## Checklist

- [ ] Title describes what they'll build, not what they'll learn
- [ ] Introduction shows the concrete end result
- [ ] Prerequisites are minimal
- [ ] Every step produces visible output
- [ ] "You should see" appears after each significant action
- [ ] No choices offered — one clear path only
- [ ] No explanations of why things work (save for explanation docs)
- [ ] Potential failures anticipated with recovery guidance
- [ ] "What you've learned" summarizes concrete skills gained
- [ ] Next steps guide to continued learning

## When to Use

| User's mindset | Doc type | Example |
|---|---|---|
| "I want to learn" | **Tutorial** | "Build your first chatbot" |
| "I want to do X" | How-To | "How to configure SSO" |
| "I want to understand" | Explanation | "How our caching works" |
| "I need to look up Y" | Reference | "API endpoint reference" |
