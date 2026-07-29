---
name: Documentation Index
relation: README.md
description: Master table of contents with section-specific rules and folder structure overview
type: Enforce
---

# Documentation Index

## Structure

```
docs/
├── core/                   # System-wide ground truth (all Enforce)
│   ├── systemDesign.md     — Architecture and system design
│   ├── design.md           — Design decisions and rationale
│   ├── codingPrinciple.md  — Code standards and conventions
│   └── SKILLS.md           — Agent skills and subagent reference
└── modules/                # Domain patterns (Enforce rules + editable docs)
    ├── RULES.md            — Domain creation rules (Enforce)
    ├── docsRules.md        — Domain documentation boilerplate (Enforce)
    ├── examples/           — Example domains (editable, NOT real docs)
    │   └── user/           — Example domain showing the full pattern
    ├── user/               — User domain documentation
    ├── workspace/          — Workspace domain documentation
    └── pipeline/           — Pipeline domain documentation
```

## Section-specific rules

### docs/core/

- All `type: Enforce`. Read these before any architectural or coding decisions.
- `systemDesign.md` takes precedence over all other docs for architecture questions.
- `design.md` records past decisions — do not contradict them without updating the file.

### docs/modules/

- `RULES.md` is `type: Enforce`. All domains must follow its rules.
- `docsRules.md` is `type: Enforce`. It is the boilerplate template for writing domain documentation. Use it whenever you document a new domain.
- `examples/` contains example domains (`user/`). These are `type: editable` pattern references, NOT real documentation. Do not treat them as project docs.
- Create new domain docs at the same level as `examples/`, not inside it.
