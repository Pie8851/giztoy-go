# Issue 格式

GizClaw Issue 是实现前的设计与验收 contract。Issue 必须让没有参与前置讨论的开发者能够确认问题边界、找到正确 ownership、完成实现并用明确证据判断是否可以关闭。

Issue 只描述当前要交付的最终状态。不要把 roadmap、迁移过程、未经代码或文档支持的推测写成已存在的行为。

## 标题

标题统一使用：

```text
prefix: Subject
```

规则：

- `prefix` 使用小写字母或数字，可以用 `/` 表示模块层级，不能包含空格、空 segment 或首尾 `/`。
- 冒号后保留一个空格。
- `Subject` 描述完成后的目标，不使用“处理一下”“优化相关逻辑”等无法独立判断的表述。
- 标题 prefix 与 GitHub Issue Type 是两个独立字段。prefix 表示问题或模块类别，Issue Type 表示工作的组织形态。

示例：

```text
bug: Fix token refresh race
firmware/ota: Add recovery status reporting
admin/ui: Add workspace editor
tracking: Split server mesh delivery
```

## GitHub Issue Type

仓库使用 `Bug`、`Feature` 和 `Task`：

| Issue Type | 使用场景 | 内容要求 |
| --- | --- | --- |
| `Bug` | 已有行为与 contract 或预期不一致 | 写清当前行为、正确行为和回归验证 |
| `Feature` | 新增用户可见能力或拥有具体实现的交付物 | 包含完整设计、改动范围和验收标准 |
| `Task` | 只用于 tracking/container Issue | 必须有 sub-issues，本身不拥有具体实现设计和验收细节 |

不要因为 Issue 有 sub-issues 就自动设为 `Task`。只要父 Issue 仍拥有具体代码改动、实现设计或直接验收标准，就应根据交付物选择 `Bug` 或 `Feature`。

## 实现 Issue 的正文结构

拥有具体实现范围的 Issue 必须按以下顺序使用五个顶层章节：

```markdown
## Background

## Goal

## Code Changes Tree

## Design

## Test And Acceptance Criteria
```

章节名和顺序保持一致。可以在 `Design` 中增加与任务匹配的三级章节，但不要另建同级章节分散 scope 或验收条件。

### Background

先写依赖关系，再说明当前状态、问题原因和必要上下文。

关系字段使用严格的 Markdown 列表：

```markdown
## Background

- Parent: #123
- Prerequisite of:
  - #130
  - #131
- Follow up to:
  - #118

当前状态与问题背景。
```

关系含义：

- `Parent`：当前 Issue 属于哪个 tracking 或更大交付物，只允许一个内联 Issue 引用。
- `Prerequisite of`：哪些 Issue 必须等待当前 Issue，始终使用嵌套列表。
- `Follow up to`：当前 Issue 依赖、扩展或清理哪些既有 Issue，始终使用嵌套列表。

不要写裸 `Parent: #123`、`Related` 或带重复标题的关系行。松散参考关系写入背景正文；GitHub 的原生 parent/sub-issue 关系仍需在 GitHub 中单独设置。

### Goal

`Goal` 描述完成后的 ownership 和边界：

- 当前 Issue 必须产生的行为或输出；
- 当前 Issue 拥有的模块、contract 或用户路径；
- 明确不负责的相邻能力；
- 完成后是否可以直接关闭 Issue。

如果一个 Issue 包含多个能够独立交付和验证的目标，应拆成 tracking Issue 与 sub-issues。

### Code Changes Tree

`Code Changes Tree` 使用仓库中的真实路径描述计划改动，不复制整个仓库目录：

```text
cmd/internal/commands/example/
└── command.go                 # CLI command and flags
pkgs/gizclaw/example/
├── service.go                 # domain behavior
└── service_test.go            # behavior and error coverage
guides/zh/using/
└── example.md                 # user-facing final behavior
```

规则：

- 先读取根目录和目标目录适用的 `AGENTS.md`，再确定 ownership、测试和生成文件位置。
- 只列当前 Issue 拥有的 source、schema、生成产物、测试、配置、脚本和文档。
- 删除项标记 `(delete)`；需要提交的生成文件标记 `(generated, committed)`。
- 不列 build output、cache、下载依赖、日志、临时文件和无关路径。
- 如果正确路径仍无法从仓库和 Guides 确认，保留设计问题，不要伪造目录树。

### Design

`Design` 写清实现必须遵守的 contract，而不是提前固定每一行代码。根据任务补充：

- API、RPC、CLI、配置、Schema 或文件格式；
- 调用路径、状态转换、ownership、异步行为与清理；
- 错误、超时、重试、取消、partial failure 和 unsupported 行为；
- Server、Edge、device、desktop、browser 和各 SDK 的平台差异；
- 第三方依赖、provider 边界和生成流程；
- compatibility、migration、security 和明确的 non-goals。

如果存在多个会改变产品或架构方向的可行方案，增加 `Open Design Questions` 三级章节并等待决策。不要用未经支持的假设把 Issue 标记为 ready。

### Test And Acceptance Criteria

每条 acceptance criterion 必须描述可观察、可复现的完成状态，包括：

- 哪个调用方完成什么行为；
- 正常路径、错误路径和兼容路径必须满足什么结果；
- 哪些 source contract、生成 surface、调用方和文档需要同步；
- 使用哪些实际命令或测试证明完成；
- 需要 Docker、网络、credential、provider account 或硬件时，允许的 `SKIP`、`UNSUPPORTED` 或 blocked 证据。

Go 行为变化默认包含 `go test ./...`；如果使用 scoped equivalent，Issue 必须说明范围为什么足够。文档、README 或 workflow-only 改动至少包含：

```sh
git diff --check
```

不要使用“完成开发”“测试通过”“优化代码”等无法单独证明行为的验收条件。

## 完整示例

下面示例展示一个拥有具体实现范围的 `Feature` Issue。示例中的 Issue 编号仅用于说明关系字段格式。

- Title: `guides: Document the repository Issue format`
- Issue Type: `Feature`

<<< ./examples/issue-example.md

## Tracking Issue

`Task` tracking Issue 只维护交付边界、sub-issue 列表、依赖关系和状态，不直接拥有具体代码树、实现设计或测试细节。具体内容由 sub-issues 承担。

如果仓库模板要求 tracking Issue 仍保留五个章节，这些章节只能说明 sub-issue ownership 和汇总完成条件，不能重新复制每个子任务的实现设计。

## Ready 检查

创建或更新 Issue 后，逐项确认：

- 标题符合 `prefix: Subject`。
- Issue Type 与工作形态一致，`Task` 只用于有 sub-issues 的 container。
- 实现 Issue 的五个章节存在且顺序正确。
- Background 关系字段使用严格 Markdown 列表，并同步设置原生 GitHub 关系。
- Goal 写清 ownership、交付结果和非目标。
- Code Changes Tree 来自当前仓库结构和适用的 `AGENTS.md`。
- Design 覆盖相关 contract、生命周期、错误、平台和兼容边界。
- Acceptance criteria 可观察，并列出与风险匹配的验证命令。
- 不包含 secret、credential、token、私人路径、临时状态或把 roadmap 当成当前事实的描述。
- 未决问题会改变实现方向时，Issue 保持未 ready 状态。

Issue readiness 的审查方法见 [Issue 审查](./issue_review)。
