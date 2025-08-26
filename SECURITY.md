# Security Policy

## Security Status

This repository is continuously scanned by Mend (formerly WhiteSource).  Mend automatically creates GitHub issues for any detected vulnerabilities.

For a complete list of changes, please refer to our [CHANGELOG.md](CHANGELOG.md).

## Version Security Status

| Version | Status                       | Details                                                    | Release Date |
| ------- | ---------------------------- | ---------------------------------------------------------- | ------------ |
| 2.0.0   | :white_check_mark: Supported | Initial release with all security vulnerabilities resolved | 2024-03-26   |
| 1.x.x   | :x: Deprecated               | Contains known vulnerabilities - upgrade to 2.0.0          | -            |

## Reporting a Vulnerability

To report a security issue, please email the maintainers listed in [MAINTAINERS.md](MAINTAINERS.md) with:

- Subject line: "Security Vulnerability: s3-iam-cosi-driver"
- Description of the vulnerability
- Steps to reproduce the issue
- Affected versions
- Any known mitigations
- Any potential impacts

Our team will:

1. Acknowledge receipt within 3 business days
2. Provide an initial assessment within 10 business days
3. Follow a 90-day disclosure timeline from the time the vulnerability is confirmed

Please do NOT report security vulnerabilities through public GitHub issues.

## Security Response Process

1. Reporter submits vulnerability report
2. Maintainers assess and confirm the issue
3. Maintainers develop and test a fix
4. A new release is prepared with the fix
5. The fix is released and the vulnerability is publicly disclosed