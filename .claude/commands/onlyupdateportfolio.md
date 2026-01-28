---
description: onlyupdateportfolio
---
```mermaid
flowchart TD
    start_node_default([Start])
    end_node_default([End])
    skill_1769569353276[[Skill: portfolio-meta]]

    start_node_default --> skill_1769569353276
    skill_1769569353276 --> end_node_default
```

## Workflow Execution Guide

Follow the Mermaid flowchart above to execute the workflow. Each node type has specific execution methods as described below.

### Execution Methods by Node Type

- **Rectangle nodes**: Execute Sub-Agents using the Task tool
- **Diamond nodes (AskUserQuestion:...)**: Use the AskUserQuestion tool to prompt the user and branch based on their response
- **Diamond nodes (Branch/Switch:...)**: Automatically branch based on the results of previous processing (see details section)
- **Rectangle nodes (Prompt nodes)**: Execute the prompts described in the details section below

## Skill Nodes

#### skill_1769569353276(portfolio-meta)

**Description**: |

**Scope**: user

**Validation Status**: valid

**Skill Path**: `C:\Users\jinpuruen\.claude\skills\portfolio-meta\SKILL.md`

This node executes a Claude Code Skill. The Skill definition is stored in the SKILL.md file at the path shown above.
