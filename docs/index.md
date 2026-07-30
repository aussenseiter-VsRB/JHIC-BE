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
    ├── auth/               — Auth domain documentation
    ├── user/               — User domain documentation
    └── berita/             — Berita (news) domain documentation
```

## Section-specific rules

### docs/core/

- All `type: Enforce`. Read these before any architectural or coding decisions.
- `systemDesign.md` takes precedence over all other docs for architecture questions.
- `design.md` records past decisions — do not contradict them without updating the file.

### docs/modules/

- `RULES.md` is `type: Enforce`. All domains must follow its rules.
- `docsRules.md` is `type: Enforce`. It is the boilerplate template for writing domain documentation. Use it whenever you document a new domain.
- Create new domain docs at the same level as `auth/`, not inside it.
