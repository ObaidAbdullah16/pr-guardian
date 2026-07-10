# Contributing to PR Guardian

Thank you for your interest in contributing! 🎉

## Before you start

- Check the [open issues](https://github.com/obaid/pr-guardian/issues) to see if your idea is already being discussed
- Open an issue first for any significant change so we can discuss the approach
- Fork the repo, create a branch, make your changes, and open a PR

## Branch naming convention

Please follow this convention for branches:

| Type | Example |
|------|---------|
| New feature | `feat/add-file-size-check` |
| Bug fix | `fix/broken-link-timeout` |
| Docs | `docs/update-install-guide` |
| Chore | `chore/update-dependencies` |
| Refactor | `refactor/simplify-checker` |

## PR checklist

Your PR should:

- [ ] Reference an issue (`Fixes #123`)
- [ ] Include a meaningful description (at least 30 characters)
- [ ] Update `CHANGELOG.md` with your change
- [ ] Update `README.md` if the change affects usage
- [ ] Include or update tests for any new check logic
- [ ] Pass all existing tests (`go test ./...`)

## Running tests

```bash
go test ./internal/checker/...
```

## Code style

- Use `gofmt` to format your code before committing
- Keep functions small and focused
- Add a comment to any non-obvious logic

## Questions?

Open an issue and I'll respond as soon as possible.
