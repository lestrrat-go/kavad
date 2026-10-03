# Security Policy

## Scope

kavad plays a show that a Go program defines, and writes SVG frames, an HTML
page and WebAssembly files to a directory the caller names. Bugs in scope are a
show or output path that reaches the public API and causes a denial of service
(unbounded CPU or memory, a hang), a reachable panic, or a file written outside
the directory the caller named. A bug in the page files (`kavad.js`,
`index.html`) that lets the embedding page run script it did not load is also
in scope.

Issues that only affect test code, examples, or `capture/capture.cjs` are
welcome as ordinary bug reports rather than security reports.

## Supported Versions

No release is tagged yet. `main` is the only line that receives fixes.

| Version  | Supported                    |
| -------- | ---------------------------- |
| v0.x.x   | :white_check_mark: (pre-1.0) |
| < v0.1.0 | :x: (unreleased)             |

## Reporting a Vulnerability

If you think you found a vulnerability, please report it via
[GitHub Security Advisory](https://github.com/lestrrat-go/kavad/security/advisories/new).
Please include explicit steps to reproduce the security issue. A minimal
reproducer or failing test, and the commit SHA you tested against, are ideal.

We will do our best to respond in a timely manner, but please also be aware that
this project is maintained by a very limited number of people. Please help us
with test code and such.
