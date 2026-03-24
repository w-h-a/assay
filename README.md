# assay

<div align="center">
  <img src="./.github/assets/assay.png" alt="Assay Mascot" width="400" />
</div>

## Problem

LLMs are collapsing the cost of generating proofs and implementations. When proofs are cheap, the bottleneck shifts to specifications.

## Solution

Assay is a standalone specification language for software behavior with pluggable verification backends. The spec language is the domain. Verification backends are infrastructure. Same spec, different guarantee levels; from probabilistic (property testing) to mathematical (formal proof).

Assay separates three concerns:

1. **Specification** — what should this system do? (`.assay` file)
2. **Binding** — how does the spec connect to an implementation? (`.bind` file)
3. **Verification** — does the implementation satisfy the spec? (backend)

## Architecture

```mermaid
graph TD
    subgraph "Domain Layer"
        L[Lexer] --> P[Parser]
        P --> AST
        AST --> TC[Type Checker]
    end

    subgraph "Input"
        SF[".assay spec"] --> L
        BF[".bind file"] --> BP[Binding Parser]
    end

    TC --> VA[Validated AST]
    BP --> BR[Binding Resolver]
    BR --> VA

    VA --> B{Backend}
    B --> CG[Go Code Generator]
    B --> DC[Dafny Compiler]
    B --> LC[Lean Compiler]

    subgraph "Property Testing"
        CG --> TF["_assay_test.go"]
        TF --> GT["go test + rapid"]
    end

    subgraph "Dafny"
        DC --> DF[".dfy file"]
    end

    subgraph "Lean"
        LC --> LF[".lean file"]
    end

    GT --> V[Verdict]
    DF --> V
    LF --> V
```

## Usage

Coming soon.
