# assay

<div align="center">
  <img src="./.github/assets/assay.png" alt="Assay Mascot" width="400" />
</div>

## Problem

LLMs are collapsing the cost of generating proofs and implementations. When proofs are cheap, the bottleneck shifts to specifications.

## Solution

Assay is a standalone specification language for software behavior with pluggable verification backends. The spec language is the domain. Verification backends are plug and play. Same spec, different guarantee levels; from probabilistic (property testing) to mathematical (formal proof).

|                       | property testing                  | formal proof (model)                               | formal proof (native)                   |
| --------------------- | --------------------------------- | -------------------------------------------------- | --------------------------------------- |
| **Backend generates** | e.g., go test file + test results | e.g., dafny scaffolding + theorems + proof results | e.g., verus annotations + proof results |
| **Proofs about**      | real code (directly)              | model of code                                      | real code (directly)                    |
| **Trust gap**         | none                              | model ↔ code                                       | none                                    |
| **Guarantee**         | probabilistic                     | mathematical (of model)                            | mathematical (of code)                  |

## Architecture

```mermaid
graph TD     
    subgraph "Handler Layer"
        CLI["cmd/assay (CLI)"]
    end

    CLI -->|1| SVC

    subgraph "Service Layer"
        SVC((Service))
    end

    SVC -->|2 parse spec| L
    SVC -->|3 parse binding| BP

    subgraph "Domain Layer"
        L[Lexer] --> P[Parser]
        P --> AST
        AST --> TC[Type Checker]
        TC --> VA[Validated AST]
        BP[Binding Parser] --> BA[Binding AST]
    end

    VA -->|4| SVC
    BA -->|4| SVC

    SVC -->|5 resolve| R
    SVC -->|6 verify| B
    SVC -->|7 store| VS

    subgraph "Client Layer"
        R[Resolver]
        B[Backend]
        VS[Verdict Store]
    end

    R -.-> GP["gopackage (go/types)"]
    B -.-> PT[Property Testing]
    B -.-> FP[Formal Proof]
    VS -.-> FS["fs (~/.assay/verdicts/)"]
```

## Usage

Coming soon.
