# Security policy

## Supported versions

Only the latest release gets security fixes. Run `claude-dx update` to get it.

## Report a vulnerability

Do not open a public issue.

Use the private report form of GitHub: go to the **Security** tab of this repository and click **Report a vulnerability**. Only the maintainers can see the report.

Give:

- the affected version (`claude-dx --version`) and operating system,
- the steps to reproduce the problem,
- the effect (example: code execution, file access outside the project, a bad binary install).

The maintainers reply as soon as possible. When a fix is available, the maintainers publish a release and a security advisory, and give credit to the reporter if the reporter agrees.

## Scope

These parts are important for security:

- `claude-dx update` and the install scripts: download and checksum verification of binaries.
- The MCP doc tools: input checks that keep file operations inside the `docs/` folders of the project.
- The `SessionStart` hook and the launch of the `claude` process.

## Verify a release

The checksum in `checksums.txt` comes from the same release as the binaries. `claude-dx update` and the install scripts check only this SHA-256 checksum. It detects a corrupt download but does not prove the origin of the file. To prove the origin, run `gh attestation verify`. Each release file, and `checksums.txt`, also has a GitHub build provenance attestation from the release workflow. To verify a file, run:

```sh
gh attestation verify <file> --repo EqualSoftware/claude-dx
```
