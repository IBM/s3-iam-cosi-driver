# Contributing Guide

Thank you for your interest in contributing to the s3-iam-cosi-driver project. Before you begin, please read through our legal requirements.

## Legal Requirements

### 1. Contributor License Agreement (CLA)

All contributors must sign and submit one of the following CLAs before their contributions can be accepted:

#### Individual Contributors

If you're contributing as an individual, sign the [Individual Contributor License Agreement (ICLA)](https://www.apache.org/licenses/icla.pdf).

#### Corporate Contributors

If you're contributing on behalf of a corporation, sign the [Corporate Contributor License Agreement (CCLA)](https://www.apache.org/licenses/cla-corporate.txt). The CCLA allows a corporation to authorize contributions from its designated employees.

Submit your signed CLA to one of the maintainers listed in [MAINTAINERS.md](MAINTAINERS.md).

### 2. Developer Certificate of Origin (DCO)

After submitting your CLA, all commits must include a DCO sign-off. The [Developer Certificate of Origin](https://developercertificate.org/) certifies that you wrote or have the right to submit the code you are contributing.

Add the following sign-off statement to your commits:

```text
Signed-off-by: Your Name <your.email@example.com>
```

You can add this automatically using:

```bash
git commit -s
```

### 3. License Headers

Each source file must include the appropriate license header:

For MIT licensed files (new code):

```golang
/*
Copyright (c) 2024 <holder>

Licensed under the MIT License.
*/
```

For Apache 2.0 licensed files (original codebase):

```golang
/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
```

## Contributing Process

- For **new features**: Please [raise an issue](https://github.ibm.com/graphene/s3-iam-cosi-driver/issues) first to discuss your proposal

- For **bug fixes**: Please [raise an issue](https://github.ibm.com/graphene/s3-iam-cosi-driver/issues) to track the fix

### Contribution Workflow

1. **Fork and Clone**
   - Fork the repository
   - Clone your fork locally
   - Add the upstream repository as a remote

2. **Create a Branch**
   - For features: `feature/short-description`
   - For bugs: `fix/short-description`
   - For docs: `docs/short-description`

3. **Make Changes**
   - Write clear, concise commit messages
   - Include DCO sign-off in all commits
   - Add tests for new functionality
   - Update documentation as needed

4. **Submit Pull Request**
   - Push changes to your fork
   - Create a [pull request](https://github.ibm.com/graphene/s3-iam-cosi-driver/pulls) against the main branch
   - Fill out the pull request template
   - Link any related issues

### Pull Request Guidelines

- Keep changes focused and atomic
- Maintain a clean commit history
- Follow existing code style and conventions
- Include relevant tests
- Update documentation

## Communication

- [GitHub Issues]([issues](https://github.ibm.com/graphene/s3-iam-cosi-driver/issues)): Bug reports and feature requests
- [Pull Requests](https://github.ibm.com/graphene/s3-iam-cosi-driver/pulls): Code review discussions
- [Project Discussions](https://github.ibm.com/graphene/s3-iam-cosi-driver/wiki): General questions and discussions

## Additional Resources

- [Code of Conduct](CODE_OF_CONDUCT.md)
- [README](README.md)
- [MAINTAINERS](MAINTAINERS.md)
