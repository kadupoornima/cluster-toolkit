# How to Contribute

We'd love to accept your patches and contributions to this project. There are just a few small guidelines you need to follow.

## Code of Conduct

This project follows [Google's Open Source Community Guidelines](https://opensource.google/conduct/). Please adhere to these guidelines to ensure a welcoming and inclusive environment for everyone.

## Reporting Bugs

If you find a bug in the source code or a mistake in the documentation, you can help us by [submitting an issue](https://github.com/GoogleCloudPlatform/cluster-toolkit/issues/new).

**When filing an issue, please include:**
1. A clear and descriptive title.
2. Steps to reproduce the issue.
3. The expected behavior vs. the actual behavior.
4. Your environment details (OS, Go version, Terraform version, etc.).
5. Any relevant logs or error messages.

## Pull Request Submission Guidelines

We use GitHub Pull Requests (PRs) for all code changes.

1. **Fork the Repository**: Create a fork of the repository to your own GitHub account.
2. **Create a Branch**: Create a new branch for your feature or bugfix.
   - **Target Branch**: All PRs must be targeted at the `develop` branch. PRs targeting `main` will be closed.
   - **Naming**: Use a descriptive name (e.g., `fix/networking-bug` or `feat/new-scheduler`).
3. **Make Changes**: Implement your changes.
4. **Test Your Changes**:
   - Run the test suite to ensure no regressions:
     ```bash
     make tests
     ```
   - Ensure your code passes linting and pre-commit checks:
     ```bash
     make check-pre-commit
     ```
5. **Commit**: Use clear and descriptive commit messages.
6. **Submit PR**: Push your branch to your fork and open a Pull Request against the `develop` branch of this repository.
7. **CLA**: You must sign the [Contributor License Agreement (CLA)](#contributor-license-agreement) before your PR can be merged.

## Contributor License Agreement

Contributions to this project must be accompanied by a Contributor License Agreement. You (or your employer) retain the copyright to your contribution; this simply gives us permission to use and redistribute your contributions as part of the project. Head over to <https://cla.developers.google.com/> to see your current agreements on file or to sign a new one.

You generally only need to submit a CLA once, so if you've already submitted one (even if it was for a different project), you probably don't need to do it again.

## Code Reviews

All submissions, including submissions by project members, require review. We use GitHub pull requests for this purpose. Consult [GitHub Help](https://help.github.com/articles/about-pull-requests/) for more information on pull requests.

### Standard PR Response Times

Community submissions can take up to 2 weeks to be reviewed.
