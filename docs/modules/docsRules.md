---
name: Domain Documentation Rules
relation: index.md → modules/
description: Boilerplate template and rules for agents writing documentation for a new domain
type: Enforce
---

# Domain Documentation Rules

Use this file as a boilerplate whenever you write documentation for a new domain. Follow the structure below exactly.

## 1. Required frontmatter

Every domain documentation file must start with this frontmatter block:

```yaml
---
name: {Domain Name} Domain Documentation
relation: RULES.md → modules/{domainName}/
description: One-sentence summary of what this doc covers
type: editable
---
```

- `name` — Human-readable title. Use `{Domain Name} Domain Documentation` format.
- `relation` — The documentation chain that leads here. Always start from the nearest `Enforce` ancestor. Example: `RULES.md → modules/user/`.
- `description` — Concise, one sentence. Describes the doc's content, not the domain's purpose.
- `type` — Always `editable` for domain docs.

## 2. Required sections

Every domain documentation must contain these sections in order:

1. **# Title** — H1 matching the `name` field.
2. **Overview** — What the domain is for (1-3 sentences).
3. **Entity** — The domain struct fields and their meaning.
4. **Endpoints** — Table of HTTP endpoints with method, path, and description.
5. **Data flow** — How the request flows through the layers (handler → service → repository).
6. **Rules** (if any) — Domain-specific constraints not covered by `RULES.md`.
7. **Examples** — Concrete usage examples (cURL for API calls, code snippets for service calls).

## 3. Domain doc placement

- Place each domain's documentation inside `docs/modules/{domainName}/`.
- Use a single `{domainName}.md` file per domain. Split into sub-files only if the domain is large enough to warrant separate concerns.
- Do not nest domain docs inside other domains. Each domain is a sibling directory.

## 4. Naming convention

- Directory: `docs/modules/{domainName}/` (lowercase, matching the Go package name).
- File: `{domainName}.md` — lowercase.

## 5. Relation chain rules

- The first segment of `relation` must be a file you actually read before writing.
- Chain order: README.md → index.md → RULES.md → specific domain.
- Example full chain: `README.md → index.md → RULES.md → modules/user/`.
- Never reference a file you have not read.
