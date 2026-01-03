<!-- omit in toc -->

# Contributing to Idensyra

First off, thanks for taking the time to contribute!

All types of contributions are encouraged and valued. See the [Table of Contents](#table-of-contents) for different ways to help and details about how this project handles them. Please make sure to read the relevant section before making your contribution. It will make it a lot easier for us maintainers and smooth out the experience for all involved. The community looks forward to your contributions.

> And if you like the project, but just don't have time to contribute, that's fine. There are other easy ways to support the project and show your appreciation:
>
> - Star the project
> - Share it with others
> - Mention the project in your project's readme
> - Talk about it at local meetups

<!-- omit in toc -->

## Table of Contents

- [I Have a Question](#i-have-a-question)
- [I Want To Contribute](#i-want-to-contribute)
- [Reporting Bugs](#reporting-bugs)
- [Suggesting Enhancements](#suggesting-enhancements)
- [Your First Code Contribution](#your-first-code-contribution)
- [Improving The Documentation](#improving-the-documentation)
- [Styleguides](#styleguides)
- [Commit Messages](#commit-messages)
- [Join The Project Team](#join-the-project-team)

## I Have a Question

> If you want to ask a question, we assume that you have read the available [Documentation](https://github.com/HazelnutParadise/idensyra/blob/main/README.md).

Before you ask a question, it is best to search for existing [Issues](https://github.com/HazelnutParadise/idensyra/issues) that might help you. In case you have found a suitable issue and still need clarification, you can write your question in this issue. It is also advisable to search the internet for answers first.

If you then still feel the need to ask a question and need clarification, we recommend the following:

- Open an [Issue](https://github.com/HazelnutParadise/idensyra/issues/new).
- Provide as much context as you can about what you're running into.
- Provide project and platform versions (Go, Node.js, Wails, OS).

We will then take care of the issue as soon as possible.

## I Want To Contribute

> ### Legal Notice <!-- omit in toc -->
>
> When contributing to this project, you must agree that you have authored 100% of the content, that you have the necessary rights to the content and that the content you contribute may be provided under the project license.

### Reporting Bugs

<!-- omit in toc -->

#### Before Submitting a Bug Report

A good bug report shouldn't leave others needing to chase you up for more information. Therefore, we ask you to investigate carefully, collect information and describe the issue in detail in your report. Please complete the following steps in advance to help us fix any potential bug as fast as possible.

- Make sure that you are using the latest version.
- Determine if your bug is really a bug and not an error on your side (e.g. using incompatible environment components/versions).
- Check if there is already a bug report for your issue in the [bug tracker](https://github.com/HazelnutParadise/idensyra/issues?q=label%3Abug).
- Collect information about the bug:
  - Stack trace (if any)
  - OS, platform and version (Windows, Linux, macOS, x86, ARM)
  - Version of Go/Node/Wails and any relevant dependencies
  - Steps to reproduce and expected vs actual behavior

<!-- omit in toc -->

#### How Do I Submit a Good Bug Report?

> Please do not report security-related issues, vulnerabilities, or sensitive data in public issues. Contact the maintainers directly or use GitHub Security Advisories if available.

We use GitHub issues to track bugs and errors. If you run into an issue with the project:

- Open an [Issue](https://github.com/HazelnutParadise/idensyra/issues/new).
- Explain the behavior you would expect and the actual behavior.
- Provide reproduction steps and a minimal example if possible.
- Provide the information you collected in the previous section.

Once it's filed:

- The project team will label the issue accordingly.
- A team member will try to reproduce the issue with your provided steps.
- If the team is able to reproduce the issue, it will be marked `needs-fix`, as well as possibly other tags.

### Suggesting Enhancements

This section guides you through submitting an enhancement suggestion for Idensyra, **including completely new features and minor improvements to existing functionality**. Following these guidelines will help maintainers and the community to understand your suggestion and find related suggestions.

<!-- omit in toc -->

#### Before Submitting an Enhancement

- Make sure that you are using the latest version.
- Read the [documentation](https://github.com/HazelnutParadise/idensyra/blob/main/README.md) carefully and check if the functionality is already covered.
- Perform a [search](https://github.com/HazelnutParadise/idensyra/issues) to see if the enhancement has already been suggested.
- Ensure the idea fits the scope and aims of the project.

<!-- omit in toc -->

#### How Do I Submit a Good Enhancement Suggestion?

Enhancement suggestions are tracked as [GitHub issues](https://github.com/HazelnutParadise/idensyra/issues).

- Use a clear and descriptive title for the issue.
- Provide a step-by-step description of the suggested enhancement.
- Describe the current behavior and the behavior you expect instead.
- Include screenshots or GIFs if relevant.
- Explain why this enhancement would be useful to most users.

### Your First Code Contribution

1. Install prerequisites:
   - Go 1.25+
   - Node.js 16+ (or Bun)
   - Wails CLI v2.11.0+
2. Install dependencies:
   - `go mod download`
   - `cd frontend && npm install` (or `bun install`)
3. Run the dev build:
   - `wails dev`
4. Build a production package:
   - `wails build`
5. If you update Insyra packages, regenerate symbols:
   - `cd internal && go generate`

#### Key Code Areas

- **Backend**
  - `app.go` - Main Wails bindings and Go code execution
  - `workspace.go` - Workspace and file management
  - `igonb_exec.go` - igonb notebook execution bindings
  - `python_exec.go` - Python file execution
  - `python_packages.go` - Python package management (pip)
  - `igonb/` - Core igonb module (parsing, execution, Go-Python bridge)

- **Frontend**
  - `frontend/src/main.js` - Main UI logic, Monaco editor, igonb notebook UI
  - `frontend/src/style.css` - Styling

### Improving The Documentation

- Update `README.md`, `FEATURES.md`, `QUICK_REFERENCE.md`, and `CHANGELOG.md` when behavior changes.
- Keep examples aligned with the default template in the app.
- Prefer concise, task-focused instructions.
- Document igonb notebook features and Go-Python interoperability when relevant.

## Styleguides

### Commit Messages

- Use the imperative mood (e.g., "Add workspace preview").
- Keep messages short and descriptive.
- Avoid mixing unrelated changes in a single commit.

## Join The Project Team

If you are interested in joining the core team, open an issue to start a discussion.

<!-- omit in toc -->

## Attribution

This guide is based on the **contributing-gen**. [Make your own](https://github.com/bttger/contributing-gen)!
