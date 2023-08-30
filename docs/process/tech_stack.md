# Tech stack

Currently, our development team has fewer people than components/repositories. It is important for us to use shared libraries and tools to make our life easier. It's also fine to bring in new ones if there is a reason, but that reason should be more appealing than just "let's try this new cool package" or "that's an overengineering". Also, if we decide to make a change in this list, it's better to change it in all components within a reasonable timeframe.

- Read more
  - [Best practices](./best_practices.md)
  - [Code style](./best_practices.md#code-style)

## Our technology stack

- [testify](https://github.com/stretchr/testify) or stdlib `testing` package should be used for writing tests. Testify should be used only for `assert` and `require` packages – suites here have some problems with logging and parallel tests. Common setups and teardowns should be implemented with `testing` [subtests](https://golang.org/pkg/testing/#hdr-Subtests_and_Sub_benchmarks).
- [golangci-lint](https://github.com/golangci/golangci-lint) is used for static code checks.
- [Docker Compose](https://docs.docker.com/compose/) is used for a local development environment and in CI.
- [go modules](https://go.dev/ref/mod#introduction) for vendoring.

