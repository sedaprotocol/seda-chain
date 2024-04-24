# SEDA Security Policy

Security researchers play a crucial role in identifying vulnerabilities within the SEDA ecosystem. If you've uncovered a security issue in the SEDA chain or any of its affiliated repositories, we urge you to report it promptly using the method outlined below.

> [!WARNING]
 Please refrain from opening public issues on the repository containing information about potential security vulnerabilities, as this makes it difficult to mitigate the impact of valid security concerns.

## Standard Priority Bugs 🐛

For non-sensitive operational bugs or issues, please submit a GitHub issue. If it remains unaddressed after a couple of days, feel free to tag a member of the SEDA team for attention.

## Critical Bugs or Security Issues 💥

For critical security vulnerabilities, please report to `security@seda.xyz`.

Upon receipt of your report, the SEDA team will provide an initial response outlining the subsequent steps. We'll keep you updated on the progress of remediation efforts and may ask for additional information or guidance.

Please include the following information along with your report:

- Your name and affiliation (if applicable).
- Type of vulnerability
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Explanation of exploitability
- Public or third-party awareness of the vulnerability.

> [!NOTE]
 If you believe an existing issue poses a critical security risk, please email `security@seda.xyz`, providing the issue ID and a brief rationale for escalation.

## Severity Classification 🛡️

| Severity     | Description                                                                       |
| ------------ | --------------------------------------------------------------------------------- |
| **CRITICAL** | Likely catastrophic impact (e.g., chain halts, funds at risk)                     |
| **HIGH**     | Significant impact on major functionality (e.g., DoS, consensus issues)           |
| **MEDIUM**   | Impacts specific modules or features (e.g., application logic, accounting errors) |
| **LOW**      | Minimal/marginal impact                                                           |

For a detailed explanation of each severity level, refer to the [Severity Classification Matrix](https://github.com/interchainio/security/blob/main/resources/CLASSIFICATION_MATRIX.md).

### Vulnerability Disclosure Process

The Vulnerability Disclosure Process encompasses the following steps:

1. Initial report: submission of vulnerability via email.
2. Confirmation: acknowledgement of receipt within 48 hours.
3. Assessment: evaluation by the security team, including severity determination and estimated resolution timeline.
4. Resolution: notification to verify the fix once implemented.
5. Public Disclosure: publication of vulnerability details after ensuring no further risk.

Throughout the disclosure process, we emphasize the importance of confidentiality and responsible disclosure. Should a security issue necessitate a network upgrade, additional time may be required to propose and execute the upgrade.

During this period, we request:
- Refraining from exploiting any discovered vulnerabilities.
- Demonstrating good faith by refraining from actions that disrupt or degrade SEDA services.

## Feedback on this Policy 💬

To provide feedback or suggestions for improvement, submit a pull request or email `security@seda.xyz`.