# Architecture Decision Records (ADRs)

This directory contains Architecture Decision Records (ADRs) for the Cadangkan project.

## What is an ADR?

An Architecture Decision Record (ADR) is a document that captures an important architectural decision made along with its context and consequences.

ADRs help us:
- **Document the "why"** behind significant decisions
- **Provide context** for future contributors
- **Track evolution** of the system architecture
- **Avoid repeating** past discussions
- **Share knowledge** across the team

## When to Write an ADR

Create an ADR when making decisions about:

- Technology choices (languages, frameworks, libraries)
- System architecture and design patterns
- Data storage and modeling approaches
- API design and communication protocols
- Security and authentication strategies
- Testing strategies and tooling
- Deployment and infrastructure choices
- Development workflows and processes

**Rule of thumb:** If you're making a decision that will be hard to change later or affects multiple parts of the system, write an ADR.

## How to Create a New ADR

1. **Copy the template:**
   ```bash
   cp docs/adr/template.md docs/adr/XXXX-short-title.md
   ```
   Replace `XXXX` with the next number (e.g., `0007`).

2. **Fill in all sections:**
   - Set status to "Proposed" initially
   - Add today's date
   - Complete Context, Decision, Consequences, and Alternatives sections

3. **Discuss if needed:**
   - Share with the team
   - Gather feedback
   - Revise as necessary

4. **Finalize:**
   - Set status to "Accepted" when decision is made
   - Update this README with a link to the new ADR

5. **Commit:**
   ```bash
   git add docs/adr/XXXX-short-title.md
   git commit -m "docs: add ADR-XXXX for [decision]"
   ```

## ADR Status Definitions

- **Proposed:** Decision is under discussion
- **Accepted:** Decision has been agreed upon and is being implemented
- **Deprecated:** Decision is no longer relevant but kept for historical context
- **Superseded:** Decision has been replaced by a newer ADR (reference the new ADR)

## Architecture Decision Records

### Active Decisions

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [0009](0009-configuration-and-credential-storage.md) | Configuration Management and Credential Storage | Accepted | 2025-01-04 |
| [0008](0008-cli-architecture.md) | CLI Architecture and User Interface Design | Accepted | 2025-01-03 |
| [0007](0007-mysqldump-backup-strategy.md) | mysqldump Backup Strategy | Accepted | 2025-01-02 |
| [0006](0006-connection-pool-configuration.md) | Connection Pool Configuration | Accepted | 2025-01-02 |
| [0005](0005-interface-based-mocking.md) | Interface-Based Client Design | Accepted | 2025-01-02 |
| [0004](0004-custom-error-types.md) | Custom Error Types | Accepted | 2025-01-02 |
| [0003](0003-use-sqlmock-for-testing.md) | Use go-sqlmock for Unit Testing | Accepted | 2025-01-02 |
| [0002](0002-mysql-client-architecture.md) | MySQL Client Architecture | Accepted | 2025-01-02 |
| [0001](0001-use-go-for-implementation.md) | Use Go for Implementation | Accepted | 2025-01-02 |

### Deprecated Decisions

None yet.

## Best Practices

### Writing Good ADRs

1. **Be specific:** Vague decisions are hard to evaluate
2. **Include context:** Explain why this decision matters
3. **List alternatives:** Show what else was considered
4. **Be honest about trade-offs:** Every decision has pros and cons
5. **Keep it concise:** Focus on the decision, not implementation details
6. **Use clear language:** Write for future contributors who lack context

### Example of a Good Context Section

```markdown
## Context

We need to backup MySQL databases for the Cadangkan tool. The backup
process must:
- Support databases from 100MB to 100GB
- Handle concurrent backups
- Maintain connection stability during long operations
- Support both local and remote databases

Without proper connection pooling, we risk:
- Connection exhaustion under load
- Memory leaks from unclosed connections
- Poor performance with too few connections
- Network resource waste with too many idle connections
```

### Example of a Poor Context Section

```markdown
## Context

We need connection pooling because it's a best practice.
```

## References

- [ADR GitHub Organization](https://adr.github.io/)
- [Documenting Architecture Decisions by Michael Nygard](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
- [ADR Tools](https://github.com/npryce/adr-tools)

## Questions?

If you have questions about ADRs or need help creating one, feel free to ask in the project discussions or open an issue.
