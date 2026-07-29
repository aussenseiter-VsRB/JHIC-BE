# JHIC-BE

Go backend with domain-based structure, raw SQL via pgx, and stdlib `net/http`.

## Base Documentation Rules

This project uses a frontmatter-based documentation system. Every `.md` file (except this one) has YAML frontmatter with these fields:

- `name` — Document name
- `relation` — Position in the documentation hierarchy
- `description` — What the document covers
- `type` — Either `Enforce` (read-only ground truth) or `editable` (agents may modify)

### Rules for agents

1. **Frontmatter is sacred.** Never add, remove, or modify frontmatter fields on `type: Enforce` files. Only modify `type: editable` files.
2. **Respect the relation chain.** Before writing code in any area, read all docs in its relation chain. For example, to add a new domain, read: `README.md → index.md → docs/modules/RULES.md`.
3. **No duplicate rules.** If a rule exists in an `Enforce` doc, do not repeat it elsewhere. Reference it by document name instead.
4. **Read Enforce files before writing.** Always read all `Enforce` files relevant to your task before creating or modifying any code.
5. **Domain structure.** Each business domain is a top-level package under `internal/domain/`. Every domain has: `entity.go`, `repository.go`, `repository_pg.go`, `service.go`, `handler.go`.
6. **No framework.** The project uses Go stdlib only (`net/http`, `database/sql`, `context`). No Gin, Echo, or ORM.
7. **Layered architecture.** Data flows one direction: `handler → service → repository (interface)`. Never the reverse.

### Agent workflow

1. Read `README.md` (this file) for base rules.
2. Read `index.md` for the documentation TOC and section-specific rules.
3. Read the relevant section docs (e.g., `docs/modules/RULES.md` for domain work).
4. Read `type: Enforce` docs before writing any code.
5. For domain creation, read the examples in `docs/modules/examples/` for reference.
6. Write your code following the documented patterns.
7. If modifying existing docs, only touch `type: editable` files.
