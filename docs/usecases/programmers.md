# TUI Outliner for Programmers

TUI Outliner (tuo) is a powerful tool for organizing development work, from high-level planning to detailed documentation. This guide explores practical use cases tailored to the programmer's workflow.

## Project Planning & Architecture

Break down complex features into manageable tasks and subtasks, organize system design decisions, and track implementation steps.

**Use cases:**
- Outline features with implementation tasks and subtasks
- Design API endpoints with request/response structures
- Plan database schemas with table relationships
- Track refactoring steps with nested dependencies
- Organize sprint tasks with priorities and deadlines

**Example:**
```
Feature: User Authentication
  ├─ Database Design
  │  ├─ users table (id, email, password_hash, created_at)
  │  └─ sessions table (id, user_id, token, expires_at)
  ├─ API Endpoints
  │  ├─ POST /auth/register
  │  ├─ POST /auth/login
  │  └─ POST /auth/logout
  └─ Frontend Integration
     ├─ Login form component
     └─ Session token storage
```

**Useful features:**
- Set deadlines with `:set visattr date` to show due dates
- Use attributes for priority: `:attr add priority high`
- Search by status: `:/ @status=done` to find completed tasks

---

## Code Documentation

Document APIs, algorithms, and internal knowledge in a hierarchical format that's easy to navigate and export.

**Use cases:**
- Document function signatures with parameter descriptions
- Outline algorithm explanations with step-by-step logic
- Create internal documentation with nested concepts
- Build README files with nested sections
- Document design patterns and architectural decisions

**Example:**
```
Module: FileProcessor
  ├─ processFile(path, options) -> Promise<Result>
  │  ├─ Parameters:
  │  │  ├─ path: string - File path to process
  │  │  └─ options: ProcessOptions - Configuration object
  │  └─ Returns: Promise<Result>
  └─ Implementation Notes
     ├─ Opens file with streaming for large files
     └─ Applies regex transformations in order
```

**Useful features:**
- Export to markdown with `:export markdown README.md`
- Use external editor `e` for detailed explanations
- Multi-line items for code snippets and descriptions

---

## Debugging & Troubleshooting

Organize debugging sessions with checklists, stack trace analysis, and hypotheses tracking.

**Use cases:**
- Create debugging checklists to verify systematically
- Analyze stack traces with links to relevant code
- Track potential causes with nested hypotheses
- Document fix steps and verification tests
- Create post-mortem notes on issues

**Example:**
```
Bug: Memory leak in event listener cleanup
  ├─ Stack Trace Analysis
  │  ├─ EventEmitter.on() line 45
  │  └─ listener not removed in cleanup
  ├─ Hypotheses
  │  ├─ Missing removeListener call
  │  ├─ Scope issue preventing cleanup
  │  └─ Event listener bound to wrong context
  ├─ Verification Steps
  │  ├─ Check for all addEventListener calls
  │  ├─ Verify removeEventListener in cleanup
  │  └─ Run memory profiler to confirm fix
  └─ Solution
     └─ Added removeListener to destructor
```

**Useful features:**
- Use attributes for severity: `:attr add severity critical`
- Track status: `:attr add status resolved` when fixed
- Link to code with URLs: `:attr add url https://github.com/user/repo/blob/main/src/file.ts#L45`

---

## Learning & Research

Organize learning materials and research notes hierarchically for easy reference.

**Use cases:**
- Outline concepts being learned (data structures, frameworks, languages)
- Organize code snippets with explanations
- Track learning resources with nested references
- Create study guides with progressive complexity
- Document gotchas and common mistakes

**Example:**
```
React Hooks Learning Path
  ├─ useState Hook
  │  ├─ Concept: Manages functional component state
  │  ├─ Syntax: const [value, setValue] = useState(initialValue)
  │  └─ Common Mistake: Don't call hooks conditionally
  ├─ useEffect Hook
  │  ├─ Concept: Side effects in functional components
  │  ├─ Dependency Array: Controls when effect runs
  │  └─ Cleanup Function: Runs on unmount
  └─ Resources
     ├─ Official Docs: https://react.dev/reference/react/hooks
     └─ Tutorial: https://example.com/react-hooks-guide
```

**Useful features:**
- Use attributes to track learning progress: `:attr add status learned`
- Set dates for when concepts were learned: `:attr add learned-date 2025-11-04`
- Link to resources: `:attr add url https://react.dev/...`
- Use external editor for detailed notes: `e`

---

## Code Review Notes

Organize and track feedback from code reviews systematically.

**Use cases:**
- Outline review feedback organized by file or function
- Track issues found with severity and type classifications
- Link to specific code locations
- Create reusable review checklists
- Document review discussions and decisions

**Example:**
```
Code Review: Pull Request #456
  ├─ Checklist
  │  ├─ Code style consistency
  │  ├─ Test coverage
  │  ├─ Documentation
  │  └─ Performance implications
  ├─ src/api/users.ts
  │  ├─ Issue: Missing input validation on email field
  │  │  ├─ Severity: High
  │  │  └─ URL: https://github.com/repo/pull/456#discussion_r123456
  │  └─ Issue: No error handling for database timeout
  │     └─ Severity: Medium
  └─ Overall Assessment
     ├─ Approval Status: Approved with changes
     └─ Next Steps: Address validation, then we can merge
```

**Useful features:**
- Use attributes for severity: `:attr add severity high`
- Track status: `:attr add status addressed`
- Link to PR discussions: `:attr add url https://github.com/...`
- Search by severity: `@severity=high` to prioritize issues

---

## Technical Decision Log

Document major decisions, trade-offs, and their context for future reference.

**Use cases:**
- Record architectural decisions with reasoning
- Outline alternatives considered and why they were rejected
- Track decision dates and ownership
- Link to related issues or discussions
- Create an audit trail of technical choices

**Example:**
```
Technical Decisions 2025
  ├─ Decision: Use PostgreSQL instead of MongoDB
  │  ├─ Date: 2025-11-01
  │  ├─ Decidedby: architecture-team
  │  ├─ Reasoning: Strong schema validation requirements
  │  ├─ Alternatives Considered
  │  │  ├─ MongoDB: Flexibility but lacks ACID transactions
  │  │  └─ SQLite: Simpler but doesn't scale for concurrent users
  │  ├─ Trade-offs
  │  │  ├─ Pro: Strong consistency and query flexibility
  │  │  └─ Con: More operational complexity
  │  └─ Reference: https://github.com/repo/issues/789
  └─ Decision: Use React for frontend
     ├─ Date: 2025-10-15
     └─ Status: Active
```

**Useful features:**
- Use date attributes for decision tracking: `:attr add decided-date 2025-11-01`
- Add decision maker: `:attr add decided-by architecture-team`
- Link to discussions: `:attr add url https://github.com/repo/issues/...`
- Search decisions by date: `@decided-date>-30d` (last 30 days)

---

## Meeting Notes & Stand-ups

Quick capture and organization of meeting information with temporal tracking.

**Use cases:**
- Document daily stand-up notes with date tracking
- Outline sprint planning decisions
- Track action items and owners
- Record meeting discussions hierarchically
- Create recurring note templates

**Example:**
```
Daily Stand-ups (November 2025)
  ├─ 2025-11-04
  │  ├─ Completed: Fixed login bug, merged PR #450
  │  ├─ In Progress: Implementing search functionality
  │  ├─ Blocked: Waiting on API schema from backend team
  │  └─ Owner: alice
  ├─ 2025-11-03
  │  ├─ Completed: Code review, deployed hotfix
  │  ├─ In Progress: Database migration planning
  │  └─ Owner: bob
  └─ 2025-11-02
     ├─ Completed: Design review meeting
     ├─ In Progress: Component refactoring
     └─ Owner: charlie
```

**Useful features:**
- Create daily notes automatically: `:dailynote` adds date attribute
- Show dates in tree: `:set visattr date` displays dates inline
- Search by date range: `@date>-7d` shows last week's notes
- Attribute for status: `:attr add status done`

---

## Configuration & Environment Management

Document environment setup, deployment procedures, and configuration options.

**Use cases:**
- Document environment variables and their purposes
- Outline deployment steps for different environments
- Track configuration options and their effects
- Create runbooks for common operations
- Document system requirements and dependencies

**Example:**
```
Deployment Guide
  ├─ Development Environment
  │  ├─ Setup Steps
  │  │  ├─ Install Node.js v18+
  │  │  ├─ npm install
  │  │  └─ npm run dev
  │  └─ Environment Variables
  │     ├─ API_URL=http://localhost:3000
  │     └─ DEBUG=app:*
  ├─ Production Environment
  │  ├─ Prerequisites
  │  │  ├─ Kubernetes cluster configured
  │  │  └─ Docker images built
  │  ├─ Deployment Steps
  │  │  ├─ Run database migrations
  │  │  ├─ Deploy container images
  │  │  └─ Run smoke tests
  │  └─ Environment Variables
  │     ├─ API_URL=https://api.example.com
  │     └─ NODE_ENV=production
  └─ Troubleshooting
     ├─ Pod crashes on startup: Check migrations
     └─ High memory usage: Review resource limits
```

**Useful features:**
- Use external editor `e` for detailed configuration explanations
- Link to documentation: `:attr add url https://...`
- Mark critical items with status: `:attr add critical true`

---

## Bug Tracking & Issue Management

Track bugs and issues with hierarchical organization and metadata.

**Use cases:**
- Create bug reports with reproduction steps
- Track related issues hierarchically
- Organize by severity, status, or component
- Link to GitHub issues and PRs
- Create issue templates with standard fields

**Example:**
```
Active Issues Q4 2025
  ├─ Bug: Login fails with special characters in password
  │  ├─ Severity: High
  │  ├─ Status: In Progress
  │  ├─ Component: Authentication
  │  ├─ Reproduction Steps
  │  │  ├─ Create account with @ in password
  │  │  ├─ Attempt to login
  │  │  └─ Observe: "Invalid credentials" error
  │  ├─ Root Cause: SQL injection vulnerability in password check
  │  └─ GitHub Issue: https://github.com/repo/issues/1234
  └─ Bug: Memory leak on page refresh
     ├─ Severity: Medium
     ├─ Status: Resolved
     ├─ Fixed In: v2.1.0
     └─ Reference: https://github.com/repo/pull/1250
```

**Useful features:**
- Use attributes for tracking: `:attr add severity high`, `:attr add status in-progress`
- Show severity in tree: `:set visattr severity,status`
- Search by severity: `@severity=high @status=in-progress`
- Link to issues: `:attr add github-issue https://github.com/repo/issues/1234`

---

## Personal Knowledge Base

Build a personal reference library of code snippets, patterns, and solutions.

**Use cases:**
- Organize code snippets by language and topic
- Document workarounds and solutions to common problems
- Create pattern library (design patterns, anti-patterns)
- Track best practices and lessons learned
- Build a personal cheat sheet

**Example:**
```
Code Snippets & Patterns
  ├─ JavaScript
  │  ├─ Array Methods
  │  │  ├─ map() for transformations
  │  │  ├─ filter() for conditional selection
  │  │  └─ reduce() for aggregation
  │  └─ Common Gotchas
  │     ├─ This binding in arrow functions vs regular functions
  │     └─ Async/await error handling with try/catch
  ├─ SQL
  │  ├─ Query Optimization
  │  │  ├─ Index common WHERE clauses
  │  │  └─ Use EXPLAIN to analyze queries
  │  └─ Transactions
  │     ├─ BEGIN TRANSACTION
  │     └─ ROLLBACK on error
  └─ DevOps
     ├─ Docker best practices
     └─ Kubernetes deployment patterns
```

**Useful features:**
- Use multi-line items for code: Write snippets directly in outline
- External editor `e` for detailed explanations
- Search by keyword: `javascript async` to find relevant notes
- Export to markdown: Share snippets with team

---

## Key Features for Programmers

### Attributes System
Track metadata on every item without cluttering the view:
- `:attr add status done` - Track task completion
- `:attr add priority high` - Mark important items
- `:attr add date 2025-11-15` - Set deadlines
- `:attr add url https://...` - Link to external resources
- `:set visattr date,status` - Display attributes inline

### Search & Filtering
Find items quickly across complex outlines:
- `/task @status=done` - Find completed tasks
- `/@priority=high` - Find high-priority items
- `/@date>-7d` - Items with dates in the last week
- `/d:>2` - Find items at depth greater than 2
- `/@url` - Find items with URL attributes

### Export to Markdown
Share documentation with your team:
- `:export markdown README.md` - Export entire outline
- Preserves hierarchy with nested bullet points
- Includes title as top-level header

### External Editor Integration
Write detailed explanations in your preferred editor:
- `e` - Open current item in $EDITOR
- Edit text, tags, and attributes with TOML frontmatter
- Changes automatically applied when saved

### Multi-line Items & Text Wrapping
Write more than a brief title:
- Use `Shift+Enter` to add newlines within items
- Long text automatically wraps to fit terminal width
- Ideal for detailed descriptions and code snippets

### Backup System
Never lose your work:
- Automatic backups before every save
- `[b` / `]b` - Navigate backup history
- `gd` or `:diff` - Compare backup versions
- Backups stored in `~/.local/share/tui-outliner/backups/`

### Daily Notes Mode
Quick capture with automatic date tracking:
- `:dailynote` - Create item with today's date
- Items automatically tagged with date attribute
- Search by date: `/@date>-7d` for this week's notes

### Calendar Widget
Navigate and select dates interactively:
- `gc` or `:calendar` - Open calendar picker
- Click or navigate with vim-style keys (hjkl)
- Create or select items by date visually

---

## Tips for Effective Use

1. **Start with a template** - Create reusable templates for common documents (code reviews, bug reports, meeting notes)

2. **Use consistent attribute names** - Define naming conventions for your team (status, priority, severity, owner)

3. **Regular exports** - Export documentation to markdown for sharing and version control

4. **Date tracking** - Use date attributes liberally to track when decisions were made or work was completed

5. **Search before creating** - Use the search with filters to check if something already exists

6. **Backup reviews** - Periodically review backups to recover from accidental deletions

7. **External editor for deep work** - Use `e` when writing detailed explanations or code snippets

8. **Multi-line items for context** - Break down complex items with nested children rather than trying to fit everything in one line

---

## Integration with Development Tools

### With Version Control
- Link commits and PRs in URL attributes
- Track decision history alongside code history
- Reference issues in commit messages

### With Issue Trackers
- Link to GitHub issues, GitLab issues, or Jira
- Sync priorities across tools via attributes
- Use tuo for deeper personal notes

### With Project Management
- Export sections as markdown for wiki/documentation
- Use daily notes for stand-up input
- Track technical debt alongside features

### With Code Review
- Organize feedback hierarchically
- Link to specific code locations
- Create reusable review checklists

---

Start using tuo for your programming workflow and adapt these patterns to your specific needs. The outline format naturally mirrors how programmers think about problems: breaking them down into manageable pieces, tracking progress, and maintaining context.
