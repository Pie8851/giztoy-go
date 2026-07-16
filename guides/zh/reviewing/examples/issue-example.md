## Background

- Parent: #123
- Prerequisite of:
  - #126
- Follow up to:
  - #118

Issue 写作要求目前分散在 Agent 指令、审核文档和既有 Issue 中。开发者无法从单一入口确认标题、Issue Type、正文结构、关系字段和验收要求，导致不同 Issue 的信息完整度不一致。

## Goal

在审核指引中增加一份统一的 Issue 格式文档，让开发者和 Agent 可以在创建或审查 Issue 时使用同一 contract。

本 Issue 负责：

- 定义实现 Issue 的标题、Issue Type 和五段正文结构；
- 定义 tracking Issue 与实现 Issue 的边界；
- 提供一份由 VitePress 直接加载的完整 Markdown 示例。

本 Issue 不修改 GitHub Issue template、CLI、API 或产品行为。

## Code Changes Tree

```text
guides/.vitepress/
└── config.mts                              # sidebar and example exclusion
guides/zh/reviewing/
├── issue-format.md                         # canonical Issue format
├── issue_review.md                         # readiness review entrypoint
└── examples/
    └── issue-example.md                    # complete Markdown example
```

## Design

- 将 Issue 格式放在“审核指引”中，并在侧边栏中置于“Issue 审查”之前。
- 使用 `prefix: Subject` 标题格式，并保持 GitHub Issue Type 与标题 prefix 相互独立。
- 实现 Issue 固定使用 `Background`、`Goal`、`Code Changes Tree`、`Design` 和 `Test And Acceptance Criteria` 五个顶层章节。
- 示例保存在独立 Markdown 文件中，由 VitePress code snippet 语法加载；示例文件不生成独立文档页面。
- Issue 审查文档只描述 readiness 审核方法，通过链接复用格式 contract，不复制规则。

## Test And Acceptance Criteria

- “审核指引”侧边栏包含“Issue 格式”，并位于“Issue 审查”之前。
- Issue 格式页面显示完整 Markdown 示例、语法高亮和复制按钮。
- 修改示例文件后，VitePress 页面同步显示新内容，不需要在格式文档中复制示例正文。
- 示例文件不会生成可独立访问的文档页面。
- 以下验证通过：

```sh
git diff --check
npm --prefix guides run build
```
