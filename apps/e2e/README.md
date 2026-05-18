# End-to-End Tests

Playwright tests live in this workspace and are run manually:

```sh
npm run e2e:test
```

The Playwright config starts the Go API and Angular dev server automatically when they are not already running. These tests are intentionally not part of pre-commit.
