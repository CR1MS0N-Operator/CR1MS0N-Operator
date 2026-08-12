# 👋 Hi, I'm Darrius Grate — `@CR1MS0N`

**Red team operator · security tool builder · reverse engineer.**

I build the tooling that makes offensive security operations faster, stealthier, and more automated — from C2 infrastructure to operator workstations to the research behind them.

[![GitHub followers](https://img.shields.io/github/followers/CR1MS0N-Operator?style=for-the-badge&logo=github&color=181717)](https://github.com/CR1MS0N-Operator)
[![C](https://img.shields.io/badge/C-00599C?style=for-the-badge&logo=c&logoColor=white)](https://github.com/CR1MS0N-Operator/ACLGuard-Active-Directory-Permission-Auditor)
[![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://github.com/CR1MS0N-Operator/c4)
[![Shell](https://img.shields.io/badge/Shell-4EAA25?style=for-the-badge&logo=gnubash&logoColor=white)](https://github.com/CR1MS0N-Operator/veil)
[![TypeScript](https://img.shields.io/badge/TypeScript-3178C6?style=for-the-badge&logo=typescript&logoColor=white)](https://github.com/CR1MS0N-Operator/azrael-security)

---

## 🏛️ Three Pillars

```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│   🛡️ Red Team Infrastructure      (primary)                 │
│   🤖 Active Directory Security    (secondary)               │
│   🔬 RE & Threat Intelligence     (sustaining)              │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```
*"Below the abstraction."*

---

## 🛠️ Projects

### [veil](https://github.com/CR1MS0N-Operator/veil) — Red Team Infrastructure as Code
Production-grade infrastructure with a multi-node WireGuard mesh. NixOS C2 node (Mythic), Alpine TLS redirector, Arch edge node — with automated threat scoring, nftables-based IPS, and realistic engagement OPSEC built in.

### [c4](https://github.com/CR1MS0N-Operator/c4) — C2 Control Center
Single CLI to deploy, manage, and destroy C2 frameworks (Mythic, Sliver). Orchestrates infrastructure through GraphQL (Hasura) and Docker Compose — standing up and tearing down full environments on demand.

### [nightforge](https://github.com/CR1MS0N-Operator/nightforge) — Operator Workstation
Reproducible Arch Linux workstation for red team operators and AI agent orchestration. Niri Wayland compositor, Quickshell UI layer with glassmorphism bars, Matugen dynamic theming — plus rootless Podman profiles (`ad`, `re`, `web`) isolating offensive workflows.

### [ACLGuard](https://github.com/CR1MS0N-Operator/ACLGuard-Active-Directory-Permission-Auditor) — AD Permission Auditor
C-based Active Directory permission analyzer that flags DACL misconfigurations and privilege escalation paths. Lightweight single binary with CSV/JSON export for continuous assessment pipelines.

### [security-research](https://github.com/CR1MS0N-Operator/security-research) — Container Boundary Research
Original research into container boundary weaknesses. Published analysis of OCI hook privilege escalation in rootless Podman deployments (CVE-2025-23266), plus a MITRE ATT&CK-mapped technique library covering Kerberoasting, AS-REP Roasting, DCSync, Golden Ticket, and ADCS abuse.

### [azrael-security](https://github.com/CR1MS0N-Operator/azrael-security) — Portfolio & Research Site
Personal red team portfolio built with Next.js/TypeScript — a public home for projects, writeups, and research as they're released.

---

## 🧭 Philosophy

**Proactive over reactive.** The industry spends most of its energy responding to incidents that have already happened. I build tooling and research that anticipates adversary behavior — counter-reconnaissance, adversary emulation, behavioral baselining, and continuous attack surface validation.

**Depth over breadth.** One thoroughly understood system is worth more than ten superficially known. Every component in my infrastructure is justified, documented, and manually validated before automation wraps it.

**Local-first.** Self-hosted toolchain where possible — Gitea, SearXNG, Vaultwarden, Open WebUI on internal infrastructure. Cloud when necessary, local when feasible.

**OPSEC-conscious.** Lab infrastructure mirrors real engagement constraints — WireGuard-only C2 access, dedicated redirector VMs, network segmentation, honeypot layers. Security research should not compromise the researcher.

---

## 🎓 Certifications & Training

**Completed**
- **Active Directory Red Team Specialist (AD-RTS)** — CyberWarfare Labs, 2025

**In Progress**
- **Certified Red Team Analyst (CRTA)** — CyberWarfare Labs
- **Red Team Operator Specialist (ROPS-RT1)** — Rogue Labs

**Planned**
- **CRT-ID, MCRTA, OCO-AI** — CyberWarfare Labs (post-CRTA pipeline)
- **BTF, PTF, CPIA, CRT-COI** — CyberWarfare Labs (long-term track)

---

## 🌐 Community

- **OWASP Las Vegas** — Chapter Leader (Dec 2025 – Mar 2026)
- **HackSpaceCon 2026** — Volunteer

---

*Independent security researcher. Building offensive tools, one commit at a time.*