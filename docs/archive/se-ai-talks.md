# 请深入调查研究一下，目前AI软件开发，软件工程中的最佳实践，主要研究成熟的工作流，举个例子：比如有一个软件项目，要开发一套k8s上的推理平台，只有一个人+AI Agent，来完成从项目的启动到最终项目的落地，其中人主要的角色就是产品经理，剩下的所有事情都交给AI Agent完成，

# **AI Agent 驱动的软件工程：一个人如何指挥一支 AI 软件研发团队，从 0 到 1 交付 Kubernetes 推理平台**

我认为这是 AI 软件工程中最值得认真研究的方向之一。

你提出的模式，本质上不是传统意义上的 AI 辅助编程（AI-assisted coding），而是：

**Human-led, Agent-executed Software Engineering：人类主导决策，AI Agent 承担软件工程的绝大部分执行工作。**

这两者有本质区别。

传统开发模式：

- 产品经理提出需求。
- 架构师设计系统。
- 开发工程师编写代码。
- 测试工程师负责质量。
- DevOps 工程师负责部署。
- 运维工程师负责生产运行。

你设想的新模式：

- 人类：产品负责人、最终决策者、需求验收者。
- AI Agent：产品分析师、架构师、开发工程师、测试工程师、DevOps、SRE、安全工程师、技术文档工程师。
- Harness（执行框架）：负责上下文管理、任务编排、工具调用、权限控制、验证、状态持久化和错误恢复。

这里最重要的变化是：**人不再主要负责生产代码，而是负责定义目标、约束、判断和验收；Agent 不再只是回答问题，而是对交付结果承担可验证的执行责任。**

但有一个前提：不能把所有工作简单地交给一个 Agent，然后期待它自主完成一个复杂项目。

真正成熟的方案，需要把软件工程拆成可管理、可验证、可回溯的工作流。

我会从软件工程方法论、AI Agent 工作流、工具与架构、Kubernetes 推理平台实战四个层面展开。

## **一、先研究已经被验证的 AI 软件工程实践**

截至 2026 年 11 月，我认为有四类实践最值得借鉴。它们分别解决不同层面的问题，而不是互相替代。

### **1. OpenAI：Harness Engineering**

[Chatte mit den besten KI-Modellen der Welt.](https://images.openai.com/static-rsc-4/5WMgkKYZPRc8FKemY4U2q-9SgbBfKC6LnjgRJhp12FkyjqYQq4WTU8kliZxAQUobXMJLDXGnT0xo0D9bj1RUhAZeS-AFAyZmL4-OKw_eCN2bGTcubYYvlMYXbichcH6tDbjOAkp088W09tpvel4Hvv9kgFV3iKobqn5tdJksZZU?purpose=inline)

**OpenAI — Harness Engineering**

2026-02-11 · Agent-first 软件工程实践

[Harness engineering: leveraging Codex in an agent-first world](https://openai.com/index/harness-engineering/?trk=article-ssr-frontend-pulse_x-social-details_comments-action_comment-text\&utm_source=chatgpt.com)

这是与你提出的设想最接近的案例。

OpenAI 的工程团队公开分享了一个实验：

- 人类不直接编写代码。
- Codex 负责应用代码、测试、CI、文档、可观测性和内部工具。
- 人类负责定义目标、架构约束、优先级和验收。
- 团队持续改进 Agent 的工作环境和反馈机制。

他们称之为：

> Humans steer. Agents execute.

文章披露，他们在大约五个月内构建并发布了一个内部产品，代码规模达到百万行级别，并报告了约 10 倍的开发时间效率提升。

但这不是一个可以直接套用的生产力基准：它是特定团队、产品和基础设施条件下的实验结果。

**真正值得学习的不是零人工代码，而是三项工程原则：**

1. Repository is the source of truth：项目知识必须沉淀在仓库，而不是依赖某个 Agent 的聊天历史。
2. Agent-readable environment：项目必须具备 Agent 可以理解和操作的工具、脚本、文档及验证机制。
3. Feedback loops：Agent 每次执行后都必须得到明确的验证反馈，而不是仅仅报告自己完成了任务。

尤其值得注意的是，他们发现，早期开发效率低并非单纯因为模型能力不足，而是因为环境没有充分定义，Agent 缺少完成任务所需要的工具和约束。

这意味着：**AI 软件工程的瓶颈，正在从代码生产能力转向工程环境和反馈系统的设计能力。**&#x20;

[image](https://www.google.com/s2/favicons?domain=https://openai.com\&sz=32)

OpenAI+1



### **2. Anthropic：Long-running Agent Harness**

[Logo Anthropic – Logos PNG](https://images.openai.com/static-rsc-4/3J2P6R10bAtwn--wFbQs2ZYhjXL1tAL8xMv2PSI17u7YJ5W3w8hJlqn6G-QMv5laU0lWr49unOPa55OQIH7dn8K8Uej39bGVOWTJnYnFCAvRi0TO7EV4ww2eurqVOYncsfDUFQvKXGP_DM5y4hmD51rLnFT0YxuXoKXk0eeWuA0?purpose=inline)

**Anthropic — 长周期自主软件开发**

2025–2026 · 多阶段 Agent 工作流

[Effective harnesses for long-running agents](https://www.anthropic.com/research/effective-harnesses-for-long-running-agents?trk=public_post_comment-text\&utm_source=chatgpt.com)

[Harness design for long-running apps](https://www.anthropic.com/research/harness-design-long-running-apps?utm_source=chatgpt.com)

Anthropic 的研究重点是一个实际问题：

一个 Agent 能不能持续开发一个复杂软件项目，跨越多个上下文窗口，而不丢失项目状态？

他们提出的关键做法包括：

- Initializer Agent：初始化项目、理解目标、拆分任务。
- Coding Agent：每次只完成一个明确的功能任务。
- Structured Artifacts：通过持久化文件传递项目状态、任务结果和技术决策。
- Incremental Development：每次完成一个可验证的增量，而不是一次生成整个应用。
- Evaluator：独立评价生成结果，必要时反馈修改。

2026 年的进一步研究采用了 Planner、Generator、Evaluator 三类角色，重点强化生成与独立评价之间的循环。

这对于一个人管理多个 AI Agent 的场景尤其重要。

**一个关键认识：Agent 的长期记忆不应该主要依靠对话记录，而应该依靠可读取、可验证、可交接的工程产物。**&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.anthropic.com\&sz=32)

Anthropic+1



### **3. OpenAI Codex：Issue → Plan → Implement → Verify**

[Liberando el poder de OpenAI Codex: Una guía completa para desarrolladores y entusiastas de la IA](https://images.openai.com/static-rsc-4/bumDCT8DT3qlEHXutJuoc9dvpwr38EUB2V7Jbd3RLAkeQ3ISmi-Bxb5rlFERBcOjjYPnOAlR3UBO-sf_ptWCzBtbmzm73Ua-e8VwqfBW9owcVt7lY35XQoDHliS2Mk7qHVJJeq2Lce2XDvOW_q2txRiBPG6RrXroSMO_s0QTaBk?purpose=inline)

**Codex 开发工作流**

2026 · 实际开发流程建议

[How OpenAI uses Codex](https://openai.com/business/learn/how-openai-uses-codex/?utm_source=chatgpt.com)

[Iterating Development Workflows with Codex](https://developers.openai.com/cookbook/examples/codex/iterating-development-workflows-with-codex?utm_source=chatgpt.com)

OpenAI 分享的日常工程实践中，有几个非常有价值的建议：

- 大任务先使用 Ask Mode 分析和制定计划，再进入 Code Mode。
- 把任务描述成 GitHub Issue，而不是模糊的自然语言指令。
- 单个任务尽量控制在一个工程师大约一小时能完成的范围。
- 使用 `AGENTS.md` 保存项目级约束。
- 将测试、构建、运行环境逐步自动化。
- 通过多个候选方案比较，避免过早接受第一个结果。

他们还公开了一个相当实用的项目文件组织方式：

```
project/
├── AGENTS.md
├── GOALS.md
├── PLANS.md
├── PROMPTS.md
├── README.md
└── harness/
    ├── build/
    ├── context/
    ├── code_review/
    └── build-log.md
```

这里要区分：`AGENTS.md` 是 Codex 识别的仓库指导文件，而 `GOALS.md`、`PLANS.md` 等是工作流建议，并不是 Codex 强制规定的标准。

它们分别承担：

| 文件           | 主要职责                  |
| ------------ | --------------------- |
| AGENTS.md    | Agent 的工作规范、约束、项目知识入口 |
| GOALS.md     | 项目目标、成功条件、非目标         |
| PLANS.md     | 项目整体路线图               |
| PROMPTS.md   | 可重复调用的工作流入口           |
| build/       | 每个开发阶段的执行计划           |
| context/     | 阶段性技术发现和决策            |
| build-log.md | 实际执行过程和验证证据           |
| code_review/ | 代码审查发现                |

这个方法很适合单人长期管理一个 Agent 团队。

[image](https://www.google.com/s2/favicons?domain=https://developers.openai.com\&sz=32)

OpenAI Developers+1



### **4. GitHub：Agent 与标准软件交付流程结合**

[GitHub 코파일럿 AI 코딩 라이선스 구독: Pro/Business 플랜 정품 보장](https://images.openai.com/static-rsc-4/eLjHBMbzNpunLWb5Zp2LEu5ihqSbRBmWja5idXhF7SSh_SPWWVWsrzshQDjB7I0BYfi_3hnAfY6DiOUCu6NQJUsACMKi7wgZjg7kg5RTXRv33uvf-UFmOYPhA-dqW9n0ebpDIyAWOj60dqcmGMigk2Zg2wFXSVAUowReGXDq4D4?purpose=inline)

**GitHub — Coding Agent 工作流**

Issue、Pull Request、Review、CI

[GitHub Copilot coding agent](https://docs.github.com/en/copilot/concepts/agents/coding-agent/about-coding-agent?utm_source=chatgpt.com)

GitHub 这条路线的核心价值，是让 AI Agent 不脱离成熟的软件交付体系。

Agent 不直接在生产环境随意修改代码，而是：

1. 从 Issue 获取任务。
2. 在独立分支或开发环境中工作。
3. 提交代码变更。
4. 创建 Pull Request。
5. 执行自动化测试和 CI。
6. 接受独立审查。
7. 满足条件后合并。

这非常重要：AI Agent 的工作结果应该成为一个可以审查、验证、回滚的工程变更，而不是一个聊天窗口里的承诺。

## **二、综合这些实践：我建议采用的总体工程模型**

我不建议简单照搬某一家公司的方法。

对于一个人加 AI Agent 的项目，应该融合：

- 传统软件工程的需求管理、架构设计和质量控制。
- Spec-Driven Development（规格驱动开发）。
- Agentic Coding（Agent 自主执行）。
- Harness Engineering（工程环境与执行约束）。
- GitOps / CI/CD（可重复交付）。
- Independent Evaluation（独立质量评价）。

我把它归纳为一个六层模型。

## **AI-Native Software Engineering**

人类负责目标和最终决策，Agent 负责执行，自动化系统负责验证。

**L1 · Human Governance**

产品目标、优先级、预算、风险、最终验收

**L2 · Product & Specification**

需求分析、PRD、用户故事、验收标准、规格管理

**L3 · Architecture & Planning**

系统设计、ADR、技术选型、任务拆解、依赖管理

**L4 · Agent Execution**

开发、测试、代码审查、文档、DevOps、安全

**L5 · Engineering Verification**

自动化测试、集成测试、性能测试、独立审查

**L6 · Delivery & Operations**

构建、发布、部署、监控、故障处理、持续迭代

**横向贯穿：Harness & Engineering Memory**

Agent 上下文、工具、权限、任务状态、执行记录、验证证据、知识沉淀、恢复机制。

这里有一个重要区别：

Harness 不是第七个开发阶段，而是贯穿六层的基础设施。没有它，Agent 可能在需求、设计、代码、测试之间丢失状态，也可能反复犯已经解决过的错误。

### **整个软件生命周期应该是什么样？**

1. **Discover — 产品探索**

   明确用户、场景、问题、产品边界。
2. **Specify — 需求规格化**

   PRD、用户故事、验收标准、非功能需求。
3. **Design — 系统设计**

   架构图、模块边界、API、数据模型、ADR。
4. **Plan — 工程计划**

   拆分 Epic、Feature、Task，确定依赖和交付顺序。
5. **Build — Agent 实施**

   Agent 在独立工作区完成具体任务，提交可审查变更。
6. **Verify — 独立验证**

   测试、Review、安全扫描、性能评估、验收。
7. **Release — 交付上线**

   版本构建、部署、灰度、回滚和发布说明。
8. **Operate — 运行和演进**

   监控、故障分析、用户反馈、下一轮需求。

注意，这不是严格单向的瀑布模型。每个阶段都可以反馈到之前的阶段。例如，性能测试发现架构不合理，就应该回到 Design，而不是让开发 Agent 在错误架构上不断打补丁。

## **三、真正落地时，应该怎样组织 AI Agent 团队？**

现在进入最重要的部分。

假设只有一个人，没有其他工程师，也没有传统的研发团队。你负责产品，AI Agent 负责其余工作。

我的建议是：**不要按照传统公司的部门数量创建 Agent，而是按照软件交付过程中的职责和质量边界创建 Agent。**

### **1. 推荐的 Agent 组织架构**

**Human · Product Owner**

唯一人类决策者

目标、需求、优先级、预算、重大技术决策、上线批准

任务分配 / 结果验收

**Orchestrator Agent**

总调度

理解目标、拆解任务、安排专业 Agent、跟踪状态、处理依赖、汇总结果

**Product Agent**

需求分析、PRD、用户故事、验收标准

**Architect Agent**

系统架构、技术选型、ADR、API 设计

**Backend Agent**

后端、业务逻辑、API、数据库

**Frontend Agent**

前端、交互、管理控制台

**Kubernetes Agent**

Operator、CRD、调度、集群集成

**DevOps Agent**

CI/CD、镜像、Helm、GitOps

**QA Agent**

测试设计、自动化、端到端验收

**Security Agent**

权限、漏洞、镜像、供应链审查

**Independent Reviewer / Evaluator**

独立检查设计、代码、测试证据和实际运行结果，不直接继承实现 Agent 的自我评价。

统一工作区、Git、CI 和自动化验证

**Software Delivery**

可运行、可测试、可部署、可监控、可回滚的软件产品

这里有几个设计原则。

**第一，Orchestrator 不应该亲自包办所有工作。**

它的主要工作是管理任务状态和依赖关系，而不是把所有代码都写一遍。

**第二，Architect 和 Implementer 应当适当分离。**

例如 Kubernetes Agent 可以负责某个 Operator 的设计与实现，但重大架构方案应当先由 Architect 形成设计，再由独立 Reviewer 检查。

**第三，QA 不应该只是开发 Agent 的一个附属角色。**

开发 Agent 说测试通过，并不代表测试真正覆盖了需求。QA 应当从 PRD 和验收条件出发独立设计测试。

**第四，不需要始终运行九个 Agent。**

这些是职责角色，不是九个必须常驻的进程。一个 Agent 可以承担多个角色，不同模型也可以承担不同任务。

单人项目中，我更倾向于：

- 1 个主调度 Agent。
- 1 个主力开发 Agent。
- 1 个独立 Review Agent。
- 按任务需要临时启动架构、测试、Kubernetes、DevOps 等专业 Agent。

这比创建一支庞大、相互沟通成本很高的虚拟研发团队更实际。

### **2. Agent 之间应该如何协作？**

不建议采用自由聊天式的多 Agent 协作。

建议采用：

**Task-driven + Artifact-driven + State-driven。**

也就是：

- Task-driven：所有工作都必须有明确任务。
- Artifact-driven：Agent 通过文件、代码、设计文档交接。
- State-driven：每个任务都有明确状态和完成条件。

例如：

**任务：实现 InferenceService 创建接口**

TASK-024 · Backend

示例：已验收

**输入**

- `specs/inference-service/api.md`
- `specs/inference-service/data-model.md`
- `docs/architecture/adr-003.md`

**执行要求**

- 实现创建 API。
- 完成请求参数校验。
- 持久化资源定义。
- 增加单元测试。
- 更新 API 文档。

**完成条件**

- API 符合规格。
- 单元测试通过。
- 错误处理符合约定。
- 没有引入未授权依赖。
- Reviewer 无未解决的阻断问题。

**交付物**

- Code changes
- Test report
- Review report
- Implementation summary

Agent 的执行状态可以统一定义为：

TASK STATE MACHINE

统一状态机

**BACKLOG**

**READY**

**IN_PROGRESS**

**REVIEW**

**VERIFY**

DONE

REVIEW 或 VERIFY 失败时回到 IN_PROGRESS；遇到阻塞进入 BLOCKED；任务被否决则进入 REJECTED。

这样即使一个任务跨越多个 Agent、多个会话、多个工作日，也能恢复执行状态。

### **3. 多 Agent 并行工作的关键：任务依赖图**

不能因为有多个 Agent 就随意并行开发。

例如，一个 Kubernetes 推理平台：

```
                Product Specification
                         |
                  System Architecture
                         |
               Domain Model / API Design
                         |
             +-----------+-----------+
             |           |           |
          Backend     Frontend    K8s Operator
             |           |           |
             +-----------+-----------+
                         |
                   Integration
                         |
                  QA / Security
                         |
                    Release
                         |
                    Production
```

实际上，Backend、Frontend 和 Operator 可以在 API 契约明确之后并行开发。

但如果 API 契约还没有确定，就让三个 Agent 同时开发，最终很可能出现接口不一致、字段命名不同、状态机不兼容等问题。

因此要在任务管理中定义：

- Dependencies：依赖任务。
- Inputs：输入规格。
- Outputs：预期交付物。
- Acceptance Criteria：验收标准。
- Verification：验证方法。

**并行化的前提不是 Agent 数量，而是任务之间具有稳定的接口和边界。**

## **四、实例：一个人 + AI Agent，开发 Kubernetes 推理平台**

现在进入具体的软件项目。

假设项目暂定名称为 `KubeInference`。

目标是开发一套部署在 Kubernetes 上的模型推理平台，为用户提供模型部署、推理服务管理、GPU 资源管理、自动扩缩容、服务监控等能力。

这里我建议不要从第一天就尝试实现一个完整的商业化 AI 平台。

先定义一个可交付的 MVP，再通过真实使用逐步扩展。

### **1. 首先明确产品究竟是什么**

产品不是 Kubernetes 上的一组 YAML，也不是一个管理后台。

它应该是一个能够完成完整业务闭环的系统：

[A Reference Architecture for Adopters of Envoy AI Gateway | Envoy AI Gateway](https://images.openai.com/static-rsc-4/1b3niQwtHGYOCOwkbMvh-H5-LwrNbAA_JY-MLaKLTloVG7gxPYHUObMq_8loYiLW-iBgXa0d3kmpkJXHhKB7s1fd9gpZpeB2iokDa-VSORoX5V9pcj3X57kXfy3tucm0wp7wXRc3Bnkg9RRlZ_TjRwCxByLF-Nv37ZZ0s4xZApQ?purpose=inline)

[aigateway.envoyproxy.io](https://aigateway.envoyproxy.io/blog/envoy-ai-gateway-reference-architecture/?utm_source=chatgpt.com)

**核心业务流程：**

1. 用户注册或接入一个模型。
2. 选择模型运行配置和 GPU 资源。
3. 创建一个推理服务。
4. 平台将其部署到 Kubernetes。
5. 调度 GPU 节点，启动模型推理容器。
6. 暴露推理 API。
7. 监控推理服务运行状态和资源使用情况。
8. 支持更新、扩缩容、停止和删除。

我建议把平台划分为以下领域：

| 模块                | 主要职责         | MVP  |
| ----------------- | ------------ | ---- |
| Model Management  | 模型元数据、模型来源   | 必须   |
| Inference Service | 创建、更新、删除推理服务 | 必须   |
| GPU Resource      | GPU 资源发现和分配  | 必须   |
| Scheduling        | 推理实例调度       | 基础能力 |
| Runtime           | vLLM 等推理引擎集成 | 必须   |
| API Gateway       | 推理 API 暴露    | 必须   |
| Observability     | 指标、日志、状态     | 必须   |
| Autoscaling       | 实例自动扩缩容      | 后续   |
| Multi-tenancy     | 多租户和配额       | 后续   |
| Billing           | 计量计费         | 后续   |

这里有一个非常重要的产品决策：

**不要一开始重新发明 Kubernetes 推理运行时。**

可以研究并复用 KServe、vLLM、Kubernetes、Prometheus 等成熟项目。

KServe 已经具备控制面与数据面分离、InferenceService 生命周期管理和自动扩缩容等成熟设计。其 LLM 推理相关的 LLMInferenceService 还针对高级 LLM 工作负载提供了更细粒度的能力。

[image](https://www.google.com/s2/favicons?domain=https://kserve.github.io\&sz=32)

KServe+1



你真正要做的，是把这些能力组织成一套符合自己产品定位的完整系统。

### **2. 技术架构建议**

## **KubeInference · Reference Architecture**

**User Interface**

**Web Console**

模型、服务、GPU、监控

**CLI / SDK**

API、命令行、自动化

**Platform Control Plane**

API Server

Auth / RBAC

Model Manager

Inference Manager

GPU Manager

Scheduler Adapter

Quota Manager

Status Controller

Kubernetes API / CRD / Controller

**Kubernetes Cluster**

**GPU Worker Nodes**

**vLLM**

Inference Runtime

**GPU Devices**

NVIDIA Device Plugin / GPU Operator

**Networking / Storage / Scheduling**

Gateway API、PVC、节点标签、资源约束

**Observability & Operations**

Prometheus · Grafana · Logs · Traces · Alerts · Audit

这是一个参考架构，不代表必须将所有模块做成独立微服务。

对于单人开发，我反而建议先采用模块化单体控制面 + Kubernetes Controller 的方式，避免一开始就承担微服务治理成本。

核心技术栈可以这样选择：

| 领域             | 建议技术                                |
| -------------- | ----------------------------------- |
| Backend        | Go                                  |
| Kubernetes SDK | controller-runtime / client-go      |
| Frontend       | Vue 3 + Vben Admin                  |
| API            | REST / OpenAPI                      |
| Database       | PostgreSQL                          |
| Model Runtime  | vLLM                                |
| GPU Management | NVIDIA GPU Operator / Device Plugin |
| Packaging      | Helm                                |
| Deployment     | Argo CD                             |
| Monitoring     | Prometheus + Grafana                |
| Testing        | Go test + envtest + Kind + E2E      |

这套技术组合的考虑是：Go 和 Kubernetes 原生生态比较契合；模块化单体可以降低维护复杂度；Helm 与 Argo CD 可以让部署过程可复现、可审计。

特别是 Kubernetes Operator 的开发，必须有独立测试策略。Kubernetes 官方明确区分了单元测试、集成测试和端到端测试，并强调重要功能需要对应的测试覆盖。

[image](https://www.google.com/s2/favicons?domain=https://kubernetes.io\&sz=32)

Kubernetes+1



## **五、真正从 0 到 1：完整项目研发阶段**

这是整个研究中我认为最有价值的部分。

不是简单列出开发阶段，而是明确每个阶段由哪个 Agent 执行、产生什么文件、如何验收，以及人类什么时候介入。

我建议将整个项目分成 8 个阶段。

### **Phase 0：项目启动与初始化**

人类重点参与

**人类负责：**

- 提出产品愿景。
- 明确目标用户。
- 描述核心业务场景。
- 明确 MVP 边界。
- 确定预算和资源约束。

**Agent 负责：**

| Agent        | 工作             |
| ------------ | -------------- |
| Product      | 产品需求访谈、需求整理    |
| Architect    | 初步技术调研、方案比较    |
| Research     | 竞品、开源项目、技术生态研究 |
| Orchestrator | 汇总并建立项目计划      |

产出物：

```
project/
├── README.md
├── AGENTS.md
├── docs/
│   ├── product/
│   │   ├── vision.md
│   │   ├── prd.md
│   │   ├── mvp-scope.md
│   │   └── personas.md
│   ├── architecture/
│   │   ├── system-overview.md
│   │   ├── architecture-options.md
│   │   └── decisions/
│   └── research/
│       ├── competitors.md
│       └── technology-evaluation.md
└── project-plan.md
```

阶段验收：产品定位、MVP 边界、主要技术路线和关键风险获得人类确认。

特别强调：Agent 可以提出架构建议，但不能自行把尚未验证的假设变成既定产品需求。

### **Phase 1：需求工程与规格定义**

Agent 主导，人类审批

Product Agent 将 PRD 转换成可开发的规格。

例如：

```
specs/
├── 001-model-management/
│   ├── requirements.md
│   ├── design.md
│   └── tasks.md
├── 002-inference-service/
│   ├── requirements.md
│   ├── design.md
│   └── tasks.md
├── 003-gpu-resource/
│   ├── requirements.md
│   ├── design.md
│   └── tasks.md
└── 004-observability/
    ├── requirements.md
    ├── design.md
    └── tasks.md
```

每一个功能规格都必须回答：

1. 为什么需要这个功能？
2. 谁使用？
3. 用户如何操作？
4. 输入是什么？
5. 输出是什么？
6. 异常情况有哪些？
7. 性能和可靠性要求是什么？
8. 怎样证明功能实现正确？

举个例子：

**Inference Service 创建功能**

用户提交：

```
name: llama-8b
model:
  source: huggingface
  name: meta-llama/Llama-3.1-8B-Instruct

runtime:
  engine: vllm
  replicas: 1

resources:
  gpu: 1
  memory: 24Gi
```

验收标准：

- 能够成功创建推理服务。
- 能够生成对应 Kubernetes 工作负载。
- 能够等待模型加载完成。
- 能够获取 Ready 状态。
- 能够通过 HTTP API 完成推理请求。
- GPU 不满足要求时，明确报告调度失败原因。
- 删除服务后，相关资源按规定清理。

注意：这些是示例规格，具体字段、状态和资源模型需要在架构阶段正式确认。

**这一阶段最重要的产物不是 PRD，而是可验证的 Specification。**

### **Phase 2：架构设计与技术方案**

Architect Agent 主导，人类审批关键决策

Architect Agent 负责：

- 系统架构。
- 模块划分。
- 数据模型。
- API 契约。
- CRD 设计。
- 状态机设计。
- 安全模型。
- 技术选型。
- 性能和可靠性设计。

这里建议强制采用 ADR（Architecture Decision Record）。

例如：

```
docs/architecture/decisions/
├── ADR-001-control-plane-architecture.md
├── ADR-002-inference-runtime.md
├── ADR-003-crd-design.md
├── ADR-004-database-selection.md
└── ADR-005-gpu-resource-management.md
```

每份 ADR 包括：

```
# ADR-003: Inference Service CRD Design

## Context
为什么需要这个设计？

## Decision
采用什么方案？

## Alternatives
还考虑过哪些方案？

## Consequences
优点、缺点和影响。

## Validation
如何验证该决策？
```

不要让 Agent 每次遇到问题都重新发明一套架构。

### **Phase 3：工程基础设施建设**

DevOps Agent + Backend Agent

这是很多 AI 编程项目容易忽略的阶段。

先不要急着开发业务功能。

先建立 Agent 可以工作的工程环境。

任务包括：

- Git 仓库和分支规范。
- Go 项目结构。
- Frontend 项目初始化。
- Dockerfile。
- Helm Chart。
- Makefile。
- Lint。
- Unit Test。
- CI Pipeline。
- 自动化开发环境。
- 本地 Kubernetes 测试集群。

例如：

```
Makefile

make lint
make test
make test-unit
make test-integration
make test-e2e
make build
make docker-build
make deploy-dev
make clean
```

要求 Agent 能通过标准命令完成开发和验证。

**工程环境应该先于业务功能成熟。**

这也是 Harness Engineering 的重要思想：不要只让 Agent 编写代码，还要让 Agent 拥有稳定、可重复、可验证的执行环境。

### **Phase 4：MVP 功能开发**

多个 Agent 并行

此时才开始大规模实施。

建议按照垂直业务能力划分任务，而不是按照文件或代码行数划分。

例如：

| Sprint   | 交付目标                     | 主要 Agent         |
| -------- | ------------------------ | ---------------- |
| Sprint 1 | Model Management         | Backend          |
| Sprint 2 | InferenceService API     | Backend + K8s    |
| Sprint 3 | vLLM Runtime Integration | K8s              |
| Sprint 4 | GPU Resource Integration | K8s              |
| Sprint 5 | Frontend Management      | Frontend         |
| Sprint 6 | Monitoring & Status      | Backend + DevOps |
| Sprint 7 | End-to-End Integration   | QA               |

每个 Sprint 交付的都应该是可运行的功能，而不是孤立代码。

例如 Sprint 3 完成后，应该能够真正通过 Kubernetes 创建一个 vLLM 推理服务，而不只是生成一个 Deployment YAML。

### **Phase 5：测试、审查与质量保障**

必须设置独立质量关卡

这部分不能交给开发 Agent 自己宣布完成。

建议建立五层质量体系：

**L1 · Static Analysis**

格式、Lint、类型检查、依赖检查、安全扫描

自动化

**L2 · Unit Tests**

业务逻辑、状态机、参数校验、边界条件

自动化

**L3 · Integration Tests**

数据库、Kubernetes API、Controller、组件集成

自动化

**L4 · End-to-End Tests**

真实集群部署、模型启动、API 推理、更新、删除

必须有真实环境验证

**L5 · Performance & Reliability**

并发、延迟、GPU 利用率、故障恢复、稳定性

必须有真实环境验证

对于推理平台，尤其要关注：

- 模型加载失败。
- GPU 不可用。
- Pod Pending。
- 节点故障。
- 模型下载中断。
- 推理进程崩溃。
- Kubernetes API 暂时不可用。
- 多次重试导致重复资源。
- 删除资源时产生孤儿 Pod。
- 滚动更新时服务中断。
- GPU 显存不足。
- 大模型加载超时。

这些问题仅靠 Mock 测试无法充分验证。

### **Phase 6：部署与发布**

DevOps Agent 主导，人类授权

建议采用：

```
Git Repository
      |
      v
Pull Request
      |
      v
CI Pipeline
      |
      +-- Unit Tests
      +-- Integration Tests
      +-- Security Scan
      +-- Build Image
      |
      v
Container Registry
      |
      v
Helm / GitOps
      |
      v
Argo CD
      |
      v
Kubernetes Cluster
      |
      v
Health Checks / Smoke Tests
      |
      v
Release
```

建议将开发环境、测试环境、生产环境分开。

生产环境部署应当使用独立授权，不能让普通开发 Agent 随意修改。

Google SRE 的发布工程实践强调可复现构建、自动化测试、自动化部署、小批量发布和 Canary。Argo CD 则通过 Git 作为期望状态的来源，持续比较实际集群状态与声明状态。

[image](https://www.google.com/s2/favicons?domain=https://sre.google\&sz=32)

sre.google+2



### **Phase 7：运行、反馈与持续演进**

Agent 持续运行

软件交付并不是结束。

后续应该建立：

- Monitoring Agent：分析指标和趋势。
- Incident Agent：分析故障。
- SRE Agent：提出稳定性改进建议。
- Product Agent：整理用户反馈。
- Development Agent：修复问题和实现新功能。

形成：

```
Production
    |
    v
Metrics / Logs / Traces
    |
    v
Incident Detection
    |
    v
Agent Investigation
    |
    v
Root Cause Analysis
    |
    v
Issue Creation
    |
    v
Fix + Tests + Review
    |
    v
Release
```

但这里有一条红线：

Agent 可以自动分析、提出修复、创建 PR、执行测试；涉及生产数据、权限、资源销毁和重大变更时，必须按照事先批准的权限策略执行。

## **六、最关键的基础设施：Harness Engineering 应该怎么建设？**

上面的 Agent 团队和工作流，如果没有一个统一的执行环境，很快就会出现混乱。

例如：

- Agent A 修改了 API，Agent B 还在使用旧接口。
- 上一个会话已经决定采用某种架构，新会话却重新设计。
- Agent 说测试通过，但没有实际运行测试。
- 开发 Agent 修改了代码，却忘记同步更新文档。
- 多个 Agent 同时修改相同文件。
- 任务已经完成，但项目管理状态没有更新。
- Agent 中途退出，下一个 Agent 不知道之前做到了哪里。

这些问题并不是简单增加 Prompt 就能解决的。

我建议将 Harness 设计成五个核心子系统。

### **1. Context Engineering：上下文管理**

建立分层知识体系。

| 层次                  | 内容           | 存储位置      |
| ------------------- | ------------ | --------- |
| Project Context     | 项目目标、架构、约束   | docs/     |
| Engineering Context | 编码规范、测试规范    | AGENTS.md |
| Feature Context     | 功能规格、接口、设计   | specs/    |
| Task Context        | 当前任务、依赖、执行状态 | tasks/    |
| Execution Context   | 实际操作和验证记录    | build-log |
| Decision Context    | 已经确认的技术决策    | ADR       |

最重要的原则：

**一个事实只能有一个权威来源，其他文件通过链接引用。**

例如，CRD 的字段定义应该以正式 API Schema 为准，而不是在 PRD、架构文档、Agent 记忆里维护三份互相独立的定义。

OpenAI 的 Harness Engineering 实践也专门指出，过于庞大的 `AGENTS.md` 会导致上下文拥挤、规则失效和知识过期。应该提供知识地图，让 Agent 按任务读取相关资料。

[image](https://www.google.com/s2/favicons?domain=https://openai.com\&sz=32)

OpenAI+1



### **2. Task Orchestration：任务编排**

Harness 必须知道：

- 现在有哪些任务？
- 谁在执行？
- 依赖是否满足？
- 是否存在冲突？
- 哪些任务可以并行？
- 哪些任务需要人工批准？
- 任务失败后如何恢复？

初期不需要开发复杂的任务管理平台。

GitHub Issues + Projects、仓库里的任务文件，或者一个轻量任务数据库，都可以承担这个职责。

### **3. Tooling：工具环境**

Agent 应该拥有明确的工具能力：

| 能力         | 工具                   |
| ---------- | -------------------- |
| 代码操作       | Git、OpenCode、Codex   |
| 项目检索       | ripgrep、LSP、代码索引     |
| 文档         | Markdown、OpenAPI     |
| 构建         | Make、Go、Node.js      |
| 测试         | Go test、Kind、envtest |
| 容器         | Docker / BuildKit    |
| Kubernetes | kubectl、Helm         |
| GitOps     | Argo CD              |
| 安全         | Trivy、Gitleaks、依赖扫描  |
| 可观测性       | Prometheus、Grafana   |

不要只考虑 Agent 能不能调用这些工具，还要考虑它是否能看到执行结果、识别失败、修正问题。

### **4. Verification：验证机制**

这是整个 Harness 的核心。

Agent 的任务完成状态不应该由它自己决定。

例如：

```
Agent: TASK-024 completed

Harness:
  [1] Code diff exists       PASS
  [2] Formatting             PASS
  [3] Unit tests             PASS
  [4] Integration tests      PASS
  [5] Static analysis        PASS
  [6] Independent review     PASS
  [7] Acceptance criteria    PASS

Task Status: VERIFIED
```

这里的 `PASS` 必须来自实际执行结果，而不是 Agent 在总结里写了一句 PASS。

对于 Kubernetes 这种复杂软件，最好将验证命令标准化，并把日志、测试报告、版本、Git Commit 等信息一并保存。

### **5. Recovery：故障恢复**

必须考虑 Agent 本身也是不可靠组件。

例如：

- 执行超时。
- API 调用失败。
- 上下文压缩导致信息丢失。
- Agent 退出。
- 工具调用中断。
- 任务重复执行。
- Git 冲突。
- 多 Agent 并发写入。

我的建议：

1. 每个任务有唯一 ID。
2. 每个任务有明确输入和输出。
3. 代码变更在独立 Git Worktree 或分支进行。
4. 重要操作有执行日志。
5. 状态变化持久化。
6. 失败可以重试，但要考虑幂等性。
7. 不允许通过重试绕过失败的质量检查。

## **七、推荐的项目仓库结构**

将以上实践融合起来，我建议 KubeInference 的仓库从一开始就采用下面的结构。

```
kubeinference/
│
├── AGENTS.md
├── README.md
├── Makefile
├── go.mod
├── package.json
│
├── docs/
│   ├── product/
│   │   ├── vision.md
│   │   ├── prd.md
│   │   └── mvp-scope.md
│   │
│   ├── architecture/
│   │   ├── overview.md
│   │   ├── components.md
│   │   ├── data-model.md
│   │   └── decisions/
│   │
│   ├── api/
│   ├── operations/
│   └── security/
│
├── specs/
│   ├── model-management/
│   ├── inference-service/
│   ├── gpu-resource/
│   └── observability/
│
├── tasks/
│   ├── backlog/
│   ├── active/
│   ├── review/
│   └── completed/
│
├── harness/
│   ├── workflows/
│   ├── context/
│   ├── logs/
│   ├── reviews/
│   └── reports/
│
├── api/
│   └── openapi/
│
├── backend/
│   ├── cmd/
│   ├── internal/
│   └── pkg/
│
├── controllers/
│   ├── api/
│   ├── controllers/
│   └── config/
│
├── frontend/
│
├── deploy/
│   ├── helm/
│   ├── manifests/
│   └── argocd/
│
├── tests/
│   ├── unit/
│   ├── integration/
│   ├── e2e/
│   └── performance/
│
├── scripts/
│
└── .github/
    ├── workflows/
    ├── ISSUE_TEMPLATE/
    └── PULL_REQUEST_TEMPLATE.md
```

不需要第一天就把所有目录创建出来。随着项目推进逐步生成即可。

其中最值得重视的是三个目录：

- `specs/`：定义软件应该做什么。
- `tasks/`：定义 Agent 应该做什么。
- `harness/`：记录 Agent 实际做了什么，以及验证结果。

三者构成完整的工程闭环。

## **八、Spec-Driven Development：我建议把 OpenSpec 纳入整个流程**

这里单独谈一下规格驱动开发。

我认为，对于你提出的一个人管理 Agent 团队的模式，Spec-Driven Development（SDD）非常重要。

原因很简单：

传统研发团队可以通过会议、口头交流、代码 Review 解决大量模糊问题。

但一个人管理多个 Agent 时，这些隐性沟通成本无法承受。

必须把重要需求、接口、决策转化成明确的工程规格。

你之前关注过 OpenSpec、Spec Kit 等工具，我建议将它们作为工作流的一部分，而不是整个研发体系。

**推荐的规格驱动流程**

**Product Requirement**

产品需求

**Specification**

功能规格和验收条件

**Technical Design**

技术设计、API、数据模型

**Implementation Tasks**

可执行的开发任务

**Implementation**

Agent 编码

**Verification**

测试和验收

**Specification Complete**

实现与规格一致

我建议采用如下分工：

| 工具 / 方法       | 职责             |
| ------------- | -------------- |
| OpenSpec      | 功能规格、变更提案、实施任务 |
| ADR           | 架构决策及其理由       |
| GitHub Issues | 任务管理和协作        |
| AGENTS.md     | 仓库级 Agent 规范   |
| CI            | 自动化验证          |
| Harness       | 将以上环节组织成可重复工作流 |

特别注意：Spec 不是越多越好。

一个功能的规格应该足够明确，让 Agent 能独立完成，但不要把每一行代码都写成设计文档。

## **九、一个人应该如何控制整个研发过程？**

这里涉及最重要的管理问题：

**你不能同时审查所有 Agent 的每一次操作。**

如果 Agent 生产代码的速度远高于你阅读代码的速度，最终瓶颈会转移到人类的判断能力和审核带宽。

所以应该设计人工介入关卡（Human Approval Gates）。

我建议采用四级权限。

| 等级 | Agent 权限              | 是否需要人批准   |
| -- | --------------------- | --------- |
| L1 | 阅读、调研、分析、生成文档         | 通常不需要     |
| L2 | 本地编码、测试、修改开发环境        | 按项目规则授权   |
| L3 | Git 推送、合并、修改共享环境      | 需要授权或预先批准 |
| L4 | 生产发布、数据迁移、删除资源、修改安全权限 | 必须明确授权    |

### **人类应该重点审核什么？**

不是每个函数怎么写。

而是以下五件事：

1. 产品方向有没有偏离？
2. 架构是否存在不可接受的设计缺陷？
3. Agent 是否满足真实业务需求？
4. 质量和安全是否达标？
5. 是否值得进入下一个阶段？

举个例子：

开发 Agent 可以自行决定某个函数怎么实现，但不应该自行决定将原有 PostgreSQL 架构整体替换成另一种数据库。

这属于架构决策，必须经过审批。

### **人类审批的具体节点**

1. **Gate 1：产品确认**

   PRD、MVP 范围、成功标准。
2. **Gate 2：架构确认**

   系统架构、技术选型、数据模型、关键 ADR。
3. **Gate 3：重大功能确认**

   核心业务行为、API 契约、安全边界。
4. **Gate 4：正式发布**

   测试报告、风险清单、部署计划、回滚计划。

其他低风险、可逆的工程任务尽可能自动化。

这是单人管理 Agent 团队能够成立的关键。

## **十、工具选型：究竟用什么构建这支 AI 研发团队？**

结合上述工作流，我把工具分成四层。

## **AI Engineering Toolchain**

**1. Coding Layer**

真正负责代码生产

[What Is Oh My OpenAgent (OMO)? Complete 2026 Guide to OpenCode Multi-Agent Orchestration Harness | Blog | a2a mcp](https://images.openai.com/static-rsc-4/N_twF1GO9SW6FueQ52YRjAHcCGWkTwWpXZH9o_jsnCLMDvXOVVma-gwYTbBAWgPgHCXzG6s9UeTIFy_ZikYcHkUHtTwoQ_vlJQc0fo1wK2ez0wXrhdhTQHXoAMA15Cla8UGPZ736R63XoJkNYDuNcDNbtZTUbAiAqSbipWjvQXA?purpose=inline)

**OpenCode**

主力开发 Agent

[官方网站](https://opencode.ai/?utm_source=chatgpt.com)

[OpenAI releases Codex for Windows • Межа](https://images.openai.com/static-rsc-4/tlznayY5fdqsRRGWoxgyImUU2-A0e8kYOxHdfSrPy3J1X4rYgsotkrtm0SQG8ei9P-MvH6196dVA88TeGpymNe1pHb6NePAEOx2Dp-VzXCuvcdfVy9a7DpMKGlLfhEKaESsTtJacwuHpd0cjACfcOeJgx3m9fqx2w30L2PtlcSE?purpose=inline)

**Codex**

复杂任务、独立审查、疑难问题

[官方网站](https://developers.openai.com/codex/?utm_source=chatgpt.com)

**2. Workflow Layer**

管理规格和开发流程

[うさぎでもわかる🐰OpenSpec - AIコーディングを変える仕様駆動開発の新常識｜taku_sid🐰](https://images.openai.com/static-rsc-4/ntNCMtoVbJQkvyBBAe5KW1foIBR4cORzYtyMAldBkgEvAtgELrwqGbEhS0crkBYm9a6C5NHnOYkNGEABcPmyFHaA9lWEXPu5Vng-cG0i1LdwsgZKdlMgZtqm02rqsHpYQ7D61PcUZZzD4VAce0JcXzHchXa-Mfem4j7b0cgZxi8?purpose=inline)

**OpenSpec**

需求规格和变更流程

[GitHub](https://github.com/Fission-AI/OpenSpec?utm_source=chatgpt.com)

[Oakstone & Github | Executive Search for GTM SaaS](https://images.openai.com/static-rsc-4/HQnzkeZepdcbz6yki5F911xaI7qZFrq-EjoAnoh5focKQlZBBR_xVRxmdszYIh8CF9ZImt1zDUWQIDbeJm9vx6gJmTSUYpEpChiHmY_gLt1gXp9uq4lq-oxfEEM9g7FJhQpCYi7WHrUwBF9vcV_4-kwArSHYrPmmA6I2uBI6wFE?purpose=inline)

**GitHub**

Issues、PR、版本控制和 CI

[官方网站](https://github.com/?utm_source=chatgpt.com)

**3. Agent Orchestration Layer**

任务分配、多 Agent 协作、长期执行

[万有商城_大模型_AIGC_电商提效_流量营销_企业管理_火山引擎](https://images.openai.com/static-rsc-4/bhutmBtfYi1HQF1xPfptWIPW1-bT7UHZGD_eDB216O_1Qhfy-9xdLb_3FgYjHbF-X8oEkAKhqYUGhXeAwjDdH254yryWFPtNl53yHszeuAGyDOoNz1QS2UpTRtiJPWfcpx0-RzG-9_qqNEP1DI5fhoixk9yaRdh0S58njoe4YEc?purpose=inline)

**OpenClaw**

个人 Agent 团队和工具编排

[官方网站](https://openclaw.ai/?utm_source=chatgpt.com)

[Brand | Temporal](https://images.openai.com/static-rsc-4/Fqt2Irxq-kWE1k-v9FydEy3AbVdqQhRwCtIuYAlxlowBdpWAQ9ce2Z6HJE06Su6vgcK-xI7inbL9irURsjm-1mCuUoChvnHB51DIG3eIzvh-7SRlWaGQd0TRKoih9uLQJMPwwwb6phOY0sb_RAxO-dXHyLvXy0rxOSyhfmV8Gl4?purpose=inline)

**Temporal**

可恢复、持久化工作流（复杂化后考虑）

[官方网站](https://temporal.io/?utm_source=chatgpt.com)

**4. Delivery Layer**

软件构建、测试、发布和运行

[Docker Logo PNG, AI, EPS, CDR, PDF, SVG - iconLogoVector](https://images.openai.com/static-rsc-4/x_Ge8Om-lmdDyESi_E1do4drGQT9DSkX52vTn2ewhosAAWrXUQb75wMc1ir1acbib9ghrUC265cv3GoWGy3jQpD5Eo8uaFw2K4vVUP6L8ro7xoeh077PcgfTkj0A9dAagm-ASFMY2S_Q_IAVAVpZCEojLDWXxsXMojeGIp5mFLE?purpose=inline)

**Docker**

[PaperMC/Velocity - k8s 中文文档](https://images.openai.com/static-rsc-4/xi6TSQ0skPuUlRXG_yQsCenukJdUTpwX0I5p70Fxu9gutts9s7jCWqCVYcgRyFR4JHGah6XHYlF7eIOCBqDU_f2bSjztMqsMRPWmofsZICIVf3tV_VeXgKYG3K3Zsrgc9excjLt3HqqqLIEcUS0ULnVL1IlqTDhafuNXJmeNShM?purpose=inline)

**Kubernetes**

[Argocd Original Wordmark PNG Logo — Download Free Transparent Logo](https://images.openai.com/static-rsc-4/Q-fXe2G_dFty_FMkvGwURV60z3v_mZD0KYAzNa6Xmzh1MhkL5fp4Ou65S1x0aR3RNebo49s5TRALWQzmjlqP-0vV1oPRwiJbPJxOt0nDE8NXxCdb4RC5Da3q7ajpZrqi4wKHqMz95C0QaNFPr9bfILdQfIbfeho9fl3e1H-GpPw?purpose=inline)

**Argo CD**

### **对于你的场景，我的实际建议**

我不会建议你立即开发一个复杂的多 Agent 编排系统。

你已经有 OpenClaw、OpenCode、本地模型、多个 Agent、OpenSpec 等相关实践基础。真正需要解决的不是再增加一个 Agent 框架，而是将它们组织成一个稳定的软件交付系统。

我会这样分工：

| 工具         | 定位                     |
| ---------- | ---------------------- |
| OpenClaw   | 团队入口、任务调度、Agent 生命周期管理 |
| OpenCode   | 主要开发执行环境               |
| Codex      | 独立评审、复杂问题处理、第二意见       |
| OpenSpec   | 规格驱动开发                 |
| GitHub     | 项目知识、任务、PR、CI 的事实来源    |
| Make + CI  | 可重复执行的验证环境             |
| Kubernetes | 最终交付环境                 |

有一点需要特别强调：不要把 OpenClaw 的对话记忆直接当成项目数据库，也不要把 Agent 的会话状态当成任务状态。项目的真实状态必须由 Git、任务记录和测试证据确定。

另外，我建议先不使用复杂的多 Agent 通信协议，也不需要一开始就构建独立的 Agent 管理后台。

先把工作流跑通，再考虑工程化编排。

## **十一、最容易失败的地方：AI 软件工程的七个陷阱**

这一部分我认为比工具选择更重要。

| 陷阱    | 典型表现            | 解决办法                       |
| ----- | --------------- | -------------------------- |
| 需求不明确 | Agent 不断猜测业务    | Spec + Acceptance Criteria |
| 过度并行  | 多 Agent 接口冲突    | 依赖图 + 明确契约                 |
| 过度设计  | 架构复杂度不断膨胀       | MVP + 架构决策门槛               |
| 自我验证  | 开发 Agent 自己宣布成功 | 独立 QA + 自动化验证              |
| 上下文丢失 | 换会话就重新开始        | 仓库知识 + 持久化状态               |
| 无限修复  | Agent 不断尝试、反复修改 | 重试上限 + 根因分析                |
| 过度信任  | 未验证代码直接上线       | 权限隔离 + 发布审批                |

我再强调三个特别容易被忽略的问题。

### **1. Agent 数量不等于研发能力**

十个 Agent 不一定比三个 Agent 效率高。

Agent 越多，协调成本、上下文同步成本和代码冲突的概率也越高。

最优团队不是人数最多的团队，而是职责边界清晰、依赖关系稳定、验证闭环完整的团队。

### **2. 代码量不等于产品进度**

AI 非常容易生成大量看起来合理的代码。

但是：

- 代码编译通过，不代表功能正确。
- 单元测试通过，不代表系统能够运行。
- Kubernetes YAML 正确，不代表实际能够调度 GPU。
- 模型启动成功，不代表推理性能达标。
- Demo 能跑，不代表可以生产使用。

所以应该使用可运行的业务闭环衡量进度，而不是使用代码行数、Commit 数量衡量。

### **3. 必须让 Agent 能够发现自己的错误**

一个成熟的 Agent 工作流不是：

```
任务 → 生成代码 → 汇报完成
```

而是：

```
任务
  ↓
理解规格
  ↓
生成方案
  ↓
实施
  ↓
自动测试
  ↓
独立审查
  ↓
发现问题
  ↓
修复
  ↓
再次验证
  ↓
提交交付
```

这个循环才是 Agent 工程能力真正发挥作用的地方。

## **十二、如何衡量这个 AI 研发团队是否真正成功？**

建议从四个维度建立指标。

| 维度       | 指标                   |
| -------- | -------------------- |
| 交付效率     | 从需求到可运行功能的 Lead Time |
| 质量       | 缺陷率、回归失败率、线上故障率      |
| Agent 效率 | 首次任务成功率、返工次数、人工介入次数  |
| 产品价值     | 用户场景完成率、可用性、实际使用情况   |

特别建议增加一个指标：

**Human Intervention Ratio（人工介入比例）**

例如：

一个任务从需求到交付总共需要 10 次人工介入：

- 3 次产品确认。
- 2 次架构决策。
- 2 次设计修正。
- 2 次质量审查。
- 1 次发布批准。

随着 Harness 不断成熟，目标是让低价值的人工介入逐渐减少，把人的时间集中在真正需要判断的事情上。

不过，不应该把人工介入比例越低当成唯一目标。安全、架构和产品决策需要人类介入，是正常且必要的。

# **十三、最终建议：先做一个 30 天的 AI-Native Engineering MVP**

如果是我来实施这个项目，不会第一天就建设完整的 Agent 团队和 Harness 平台。

我会先用 30 天验证这套方法是否真的可行。

以下是一个建议的实验计划，不是对实际研发工期的保证。

# **30-Day AI Engineering Experiment**

目标：一个人 + AI Agent，完成 KubeInference 的最小可运行闭环。

1. **Week 1 · 建立工程体系**

   核心成果：工程基础设施
   - 建立 Git 仓库。
   - 配置 AGENTS.md。
   - 建立 OpenSpec 工作流。
   - 建立 CI、Lint、测试框架。
   - 建立 Agent 任务管理。
   - 确定 MVP 和架构。

   验收：Agent 能独立接收任务、修改代码、执行测试并提交结果。
2. **Week 2 · 完成核心推理能力**

   核心成果：推理服务闭环
   - Kubernetes CRD。
   - Controller。
   - vLLM Runtime。
   - 创建推理服务。
   - GPU 资源请求。
   - HTTP 推理接口。

   验收：通过声明式配置部署真实模型，并成功完成推理。
3. **Week 3 · 完成平台管理能力**

   核心成果：可管理的平台
   - Backend API。
   - Frontend Console。
   - 模型管理。
   - 推理服务管理。
   - 状态查询。
   - 基础监控。

   验收：用户可以通过 Web 控制台完成核心业务操作。
4. **Week 4 · 质量与交付**

   核心成果：可交付产品
   - Integration Tests。
   - E2E Tests。
   - 故障测试。
   - 安全检查。
   - Helm Chart。
   - CI/CD。
   - 部署文档。
   - 发布验证。

   验收：在目标 Kubernetes 环境完成可重复部署、推理和基本故障恢复。

30 天的范围应该严格控制：单集群、单一主要推理运行时、基础 GPU 管理、单模型服务生命周期、基础管理界面。多租户、复杂调度、分布式推理、计费等全部暂缓。

## **十四、我认为最值得进一步研究的三个方向**

最后，我想提出三个比单纯选择 Agent 框架更有长期价值的问题。

**方向一：Agent Engineering Team**

如何让一个人管理多个专业 Agent，形成真正的虚拟研发团队？

重点研究：

- Agent Role Design。
- Agent Orchestration。
- Task Dependency Graph。
- Multi-Agent Collaboration。
- Independent Review。

**方向二：Software Engineering Harness**

如何让 Agent 可以长期、稳定、可靠地开发一个复杂项目？

重点研究：

- Context Engineering。
- Persistent Project Memory。
- Specification Management。
- Automated Verification。
- Recovery Mechanism。
- Engineering Observability。

**方向三：Autonomous Software Delivery**

如何让 Agent 从需求一直走到生产交付？

重点研究：

- CI/CD。
- GitOps。
- Automated Testing。
- Security Governance。
- Deployment Approval。
- Production Monitoring。
- Incident Response。

我认为，这三个方向共同构成未来 AI 软件工程的核心基础设施。

# **十五、最终结论**

我想把这项研究收敛成一个核心判断：

**AI 时代，一个人能否完成过去一个研发团队的工作，关键不在于他能指挥多少个 Agent，而在于他能否构建一套让 Agent 持续、可靠、可验证地完成软件交付的工程体系。**

传统软件工程的主要生产资料是人的时间、知识和协作能力。

AI-Native Software Engineering 则逐渐转向：

| 传统软件工程 | AI-Native Software Engineering |
| ------ | ------------------------------ |
| 人编写代码  | Agent 生产代码                     |
| 人员组织团队 | Harness 组织 Agent               |
| 人工传递知识 | 工程知识库                          |
| 会议协调任务 | Task Orchestration             |
| 人工测试   | 自动化验证 + 独立评估                   |
| 人工部署   | 自动化交付                          |
| 人员经验积累 | 可复用工程规范和执行经验                   |

人类并没有失去工程责任，而是向更高层次转移：定义问题、设计约束、判断方案、建立反馈机制、承担最终责任。

这也意味着，未来一个优秀的个人开发者，可能不再需要自己精通每一个技术细节，但必须具备足够的系统思维，能够识别错误的架构、错误的需求和错误的验证结论。

**从软件工程的角度看，真正值得追求的不是 Vibe Coding，而是 Agentic Software Engineering。**

前者强调让 AI 快速生成代码，后者强调让 AI 按照工程规范，持续交付真正可用的软件。

我会把这套体系命名为：

# **One-Person AI Engineering**

## 一个人，一支 AI 团队，一套完整的软件交付体系。

**核心公式**

\\[ \text{Human Direction} + \text{Agent Execution} \\]

\\[ +\text{Engineering Harness} + \text{Automated Verification} \\]

\\[ \Downarrow \\]

**Autonomous Software Delivery**

我的建议是，下一步不要急着开发 KubeInference 的业务代码，而是先建立一套可复用的 **AI Software Engineering Harness**，然后用 KubeInference 作为第一个真实项目验证它。

这样最终得到的就不只是一个 Kubernetes 推理平台，而是两项成果：

1. 一个真正可运行的 K8s 推理平台。
2. 一套以后开发其他软件项目也能复用的个人 AI 研发体系。

这两项成果的长期价值，可能远远大于单独完成一个软件项目。

## **参考研究资料**

以下是本次研究中最值得保留的原始资料，建议优先阅读：

1. [OpenAI — Harness engineering: leveraging Codex in an agent-first world](https://openai.com/index/harness-engineering/?trk=article-ssr-frontend-pulse_x-social-details_comments-action_comment-text\&utm_source=chatgpt.com)：AI 原生研发团队的核心实践。
2. [Anthropic — Harness design for long-running apps](https://www.anthropic.com/research/harness-design-long-running-apps?utm_source=chatgpt.com)：长周期 Agent 软件开发。
3. [Anthropic — Effective harnesses for long-running agents](https://www.anthropic.com/research/effective-harnesses-for-long-running-agents?trk=public_post_comment-text\&utm_source=chatgpt.com)：跨上下文持续开发。
4. [OpenAI Cookbook — Iterating Development Workflows with Codex](https://developers.openai.com/cookbook/examples/codex/iterating-development-workflows-with-codex?utm_source=chatgpt.com)：可直接参考的项目文件和工作流。
5. [OpenAI — How OpenAI uses Codex](https://openai.com/business/learn/how-openai-uses-codex/?utm_source=chatgpt.com)：实际工程团队如何使用 Coding Agent。
6. [Google SRE — Release Engineering](https://sre.google/sre-book/release-engineering/?utm_source=chatgpt.com)：成熟软件发布工程。
7. [Google SRE — Canarying Releases](https://sre.google/workbook/canarying-releases/?trk=article-ssr-frontend-pulse_little-text-block\&utm_source=chatgpt.com)：生产环境渐进式发布。
8. [KServe — System Architecture](https://kserve.github.io/website/docs/concepts/architecture?utm_source=chatgpt.com)：Kubernetes 推理平台架构参考。

**我最希望你关注的是 OpenAI 的 Harness Engineering 和 Anthropic 的 Long-running Agent Harness。** 这两项研究最接近你想实现的目标，而且比单纯研究哪一个 Coding Agent 更有价值。

# **T3 Code：一个统一管理 AI 编程 Agent 的控制台**

你说的应该是 **T3 Code**，官网是 [t3.codes](https://t3.codes/?utm_source=chatgpt.com)，GitHub 仓库是 [pingdotgg/t3code](https://github.com/pingdotgg/t3code?utm_source=chatgpt.com)。

它值得关注，尤其是你现在使用 DeepSeek Harness、多个模型和 Agent 的场景。

先说结论：

**T3 Code 不是一个新的 AI 编程 Agent，也不是一个模型 Harness。它更像是多个 Coding Agent 的统一控制台（Control Plane）。**

它解决的问题是：

- Claude Code、Codex、OpenCode 等工具各有自己的终端和交互界面。
- 多个 Agent 并行工作时，任务、会话和代码修改容易分散。
- 想从手机查看电脑上正在运行的 Agent，不太方便。
- 不同 Agent 产生的代码变更，需要频繁切换终端和 Git 工具来检查、提交、推送。

T3 Code 把这些 Agent 放在一个界面中管理。底层 Agent 仍然运行在你的机器上，使用你已经配置好的账号和订阅。

[image](https://www.google.com/s2/favicons?domain=https://t3.codes\&sz=32)

T3 Code+1



## **一、T3 Code 的定位和架构**

[T3 Code - The open-source control plane for coding agents. - Aitoolnet](https://images.openai.com/static-rsc-4/Zi024Ia_wmXarYLL0mJe93E2G4nJz1_oJ3tMdowsq8ipzUVaDsvdSAW8zzhBRA2-xYlAbtmRVkO7GGN1qdwt3hPzQ5ds9wd1SwNNDFv_Uo-U8dy23axgw_Mb1VU18crZTxI1aDV_f4MRY0ckWOiJdyhbl6DZnx2_gyHAwqKppi0?purpose=inline)

[T3 Code](https://images.openai.com/static-rsc-4/MQM40YcIh-S1wjLe2XUGOKpcKOxF6yvL33SGJupByFCGV5BGYcFeDkbqa2i9ERhHvoXcMgo_3GKPYJa-neGaWfG-cUd8tX56ri2KThQswtlX49y4j79RoJEZ0v03vgAT4TiSVgFGlKGIN64YQWxCxTwEcs6wUTrkgDtK6xYlup4?purpose=inline)

[Theo - t3.gg (@theo) on X](https://images.openai.com/static-rsc-4/V_fRjtpBwZdF28B-iOxlg8rsHUHrY4LEoEtN79yKTciFLs7OWPkEOfm6O7oAY2AEw574Y-QKaFWfxUBn3B8TWSLY2fuSq8CReE-jbZSr9JsAnKVYjskx5Q0aX2kB6IvxwwsC6wCy-8o_Sb9QklA4KXrazUdvSULkBylV3WtPwug?purpose=inline)

8

可以把它理解成三层：

**T3 Code — Control Plane**

统一 UI · 会话管理 · Git · Worktree · 远程控制

**Claude Code**

**Codex**

**OpenCode**

**Cursor / Others**

**Local Development Environment**

本地代码仓库 · Git · Terminal · 文件系统 · 模型认证

它的核心设计理念是：

**不重新实现每一个 Agent，而是统一管理已经存在的 Agent。**

目前官方列出的集成包括 Claude Code、Codex、Cursor、Grok Build、OpenCode、Antigravity 等，此外还支持通过 ACP 接入其他 Agent。实际可用性要以具体版本和 Provider 适配情况为准。

[image](https://www.google.com/s2/favicons?domain=https://github.com\&sz=32)

GitHub+1



## **二、它主要有哪些能力？**

### **1. 多 Agent 统一管理**

[Ben Davis (@davis7) on X](https://images.openai.com/static-rsc-4/77LCFU9u0Kc5nei9mQ_IEvk17ItdBqRnahx3ubefWBjRDC0wDUHZ_KVY69Q4NFHRBlMunGdb_U3T-OA0yipZdKkx1t48qDf-aNZFeO8odtE7chwzL-DOXjWjqbhux8AuDn65e7HZAb2NQWyJ13sNMumX3KuCok_4xM-R6BnlXIs?purpose=inline)

[x.com](https://x.com/davis7/status/2039100311693107613?utm_source=chatgpt.com)

你可以在同一个界面中创建多个任务：

- 一个 Agent 负责实现后端 API。
- 一个 Agent 负责 Kubernetes Controller。
- 一个 Agent 负责前端。
- 一个 Agent 负责代码审查。

每个任务都有独立的 Thread，可以查看各自的执行进度和修改结果。

它也支持在会话中切换 Agent 或模型，例如让一个 Agent 开始工作，再交给另一个 Agent 继续处理。

[image](https://www.google.com/s2/favicons?domain=https://github.com\&sz=32)

GitHub+1



### **2. Git Worktree 隔离**

[Dominik Szaradowski | Senior Software Engineer](https://images.openai.com/static-rsc-4/SJQ7k7-aCaomcA64h5w54uMf7LLcOomu_VHqXT3kUF0ReapLsvlDsZ4hVZE3m8D0E7TBaAAuRb5y0vyQb4NAHjWE76wVkfBHob5o-7gIcKAWLy280m2PZ06Nsga9vx_NjTyC1C2917aDLy8Ow_HsC30BKCvcOxDOxkqjTkCenGY?purpose=inline)

[szaradowski.com](https://szaradowski.com/?utm_source=chatgpt.com)

这是我认为 T3 Code 最有工程价值的能力之一。

假设你要并行开发三个功能：

| Agent   | 工作内容          | 工作目录          |
| ------- | ------------- | ------------- |
| Agent A | Model CRD     | 独立 Worktree A |
| Agent B | Inference API | 独立 Worktree B |
| Agent C | Frontend      | 独立 Worktree C |

每个任务有自己的 Git 工作区和分支，避免多个 Agent 同时修改同一份工作目录。

完成后，可以查看 Diff、Commit、Push，并创建 GitHub PR。

注意：Worktree 解决的是文件工作区隔离，不会自动解决接口冲突、依赖冲突和架构不一致。这些仍然需要任务规划和集成验证。

[image](https://www.google.com/s2/favicons?domain=https://github.com\&sz=32)

GitHub



### **3. 桌面、Web、手机统一控制**

[T3 Code](https://images.openai.com/static-rsc-4/MQM40YcIh-S1wjLe2XUGOKpcKOxF6yvL33SGJupByFCGV5BGYcFeDkbqa2i9ERhHvoXcMgo_3GKPYJa-neGaWfG-cUd8tX56ri2KThQswtlX49y4j79RoJEZ0v03vgAT4TiSVgFGlKGIN64YQWxCxTwEcs6wUTrkgDtK6xYlup4?purpose=inline)

[Dominik Szaradowski | Senior Software Engineer](https://images.openai.com/static-rsc-4/SJQ7k7-aCaomcA64h5w54uMf7LLcOomu_VHqXT3kUF0ReapLsvlDsZ4hVZE3m8D0E7TBaAAuRb5y0vyQb4NAHjWE76wVkfBHob5o-7gIcKAWLy280m2PZ06Nsga9vx_NjTyC1C2917aDLy8Ow_HsC30BKCvcOxDOxkqjTkCenGY?purpose=inline)

[Theo - t3.gg (@theo) on X](https://images.openai.com/static-rsc-4/cTD0Hzpw-h_GinzVGI_25sThB1LFzvyPjoJ7_prHpCo2RlhFSfTlGlNDD7pGerY8iTr0PrOQKlMfdEwEviT0d2zfQF9LDLWvQiqEcRQLH7VvTAJYglUo70Pgf7sF1waru09rVpimkpZHFJjshB-MamDtGhGs03HRk9Od00uJW_Q?purpose=inline)

8

T3 Code 提供桌面、Web 和移动端入口。

典型使用方式：

1. MacBook 上运行 T3 Code 和底层 Coding Agent。
2. Agent 在本地代码仓库中执行任务。
3. 你离开电脑后，用手机查看 Agent 的工作情况。
4. 需要审批时，在移动端查看并处理。

远程连接可以通过局域网、Tailscale 或 T3 Connect 等方式实现。

[image](https://www.google.com/s2/favicons?domain=https://github.com\&sz=32)

GitHub+1



### **4. Git 提交和 PR 管理**

它还把一部分 Git 操作集成到了界面中：

- 查看代码 Diff。
- Commit。
- Push。
- 创建 Draft PR。
- 生成 PR 标题、描述和变更摘要。

这让它不只是一个聊天窗口，而是把 Agent 执行结果与软件交付流程连接起来。

[image](https://www.google.com/s2/favicons?domain=https://t3.codes\&sz=32)

T3 Code+1



## **三、最值得关注的地方：T3 Code 与 DeepSeek Harness 可以组合**

这里有一个与你当前环境直接相关的新信息。

**T3 Code 官方文档已经提供 DeepSeek Harness 的 ACP 接入方式。**

在 T3 Code 中：

1. 打开 Settings → Providers。
2. 选择 Add provider → Show more。
3. 选择 Local ACP command。
4. 配置本地命令：

```
dsh
```

参数逐项填写：

```
--profile
acp
```

这条接入路径已经出现在 T3 Code 官方文档中。

[image](https://www.google.com/s2/favicons?domain=https://github.com\&sz=32)

GitHub



已有官方接入路径

这意味着，你不必为了使用 T3 Code 而放弃 DSH，可以尝试让 T3 Code 直接启动 DSH 的 ACP 接口。

### **但这里有一个重要限制**

2026 年 11 月 12 日，T3 Code GitHub 上有一个专门讨论 DeepSeek Harness 集成的议题：

- DSH 可以通过 ACP 正常启动会话。
- 可以报告模型和推理参数。
- 可以完成提示词交互。
- 但是 DSH 原生 `subagent` / `subagent_fork` 创建的子 Agent，目前不会自动显示在 T3 Code 的 Subagent UI 中。

这是因为 ACP 对子 Agent 的层级关系尚没有统一的可移植表达方式。

[image](https://www.google.com/s2/favicons?domain=https://github.com\&sz=32)

GitHub+1



因此要区分：

| 能力                      | DSH 接入 T3 Code 后                 |
| ----------------------- | -------------------------------- |
| 启动 DSH 会话               | 支持                               |
| 与 DSH 对话                | 支持                               |
| 使用 DSH 暴露的模型选项          | 取决于 ACP 返回的配置                    |
| 使用 DSH 自带 Subagents     | DSH 内部可以执行                       |
| 在 T3 中展示 DSH 子 Agent 层级 | 当前存在限制                           |
| DSH Workflows           | 应由 DSH 自己负责运行，不应假定 T3 能完整管理其内部状态 |
| T3 Git Worktree 管理      | T3 自身提供                          |

这也是我认为你应该特别注意的地方：**ACP 接入不等于把 DSH 的整个内部 Harness 完整映射到了 T3。**

## **四、T3 Code 与 DSH 的职责如何划分？**

我建议采用这种组合架构：

**T3 Code — Developer Control Center**

统一交互界面、项目管理、Thread、Git Worktree、Diff、PR、远程控制

ACP

**DeepSeek Harness — Engineering Harness**

主 Agent、Subagents、Workflows、任务编排、模型路由

**DeepSeek**

**Claude**

**GLM**

**Git / Repository / CI / Kubernetes**

这里的职责划分是：

- **T3 Code：** 负责让你方便地管理、观察和控制开发任务。
- **DSH：** 负责真正的 AI 研发组织、Agent 协作和 Workflow。
- **三个模型：** 负责具体的分析、编码和审查。
- **Git 和 CI：** 负责代码变更和客观质量验证。

这比用 T3 Code 替换 DSH 更符合你目前的实际需求。

## **五、T3 Code、DSH、OpenCode、Codex 的区别**

| 维度           | T3 Code     | DeepSeek Harness  | OpenCode       | Codex                |
| ------------ | ----------- | ----------------- | -------------- | -------------------- |
| 核心定位         | Agent 控制平面  | Agent Harness     | Coding Harness | Coding Agent/Harness |
| 模型调用         | 通过底层 Agent  | 自身模型与 Provider 体系 | 通过 Provider    | OpenAI 模型及支持的配置      |
| Subagents    | 取决于底层 Agent | 原生能力              | 支持             | 支持相关任务委派能力           |
| Workflows    | 不是其核心定位     | 原生能力              | 可通过工具和扩展实现     | 以任务执行和委派为主           |
| Git Worktree | 原生管理        | 依赖具体配置和工具         | 支持相关工作流        | 支持相关工作流              |
| 手机远程控制       | 原生支持        | 取决于自身部署方式         | 取决于部署方式        | 取决于产品形态              |
| 工程流程编排       | 不作为主要职责     | 核心职责              | 可扩展            | 可扩展                  |

T3 官方 FAQ 也明确回答了：T3 Code 本身不是 Harness，而是调用其他 Harness 的界面和控制层。

[image](https://www.google.com/s2/favicons?domain=https://github.com\&sz=32)

GitHub+1



## **六、对你的实际建议**

值得试用，但不要替换 DSH

我会这样安排：

| 使用场景            | 推荐工具           |
| --------------- | -------------- |
| 项目需求分析          | DSH            |
| 架构设计与技术讨论       | DSH            |
| 多 Agent 任务分解    | DSH Subagents  |
| 多步骤研发流程         | DSH Workflows  |
| 模型按任务切换         | DSH            |
| 多任务开发界面         | T3 Code        |
| Git Worktree 隔离 | T3 Code        |
| Diff、Commit、PR  | T3 Code        |
| 质量门禁和 CI        | GitHub Actions |

尤其是 KubeInference 项目，可以这样尝试：

- T3 Thread A：模型 CRD 开发。
- T3 Thread B：InferenceService Controller 开发。
- T3 Thread C：前端管理界面开发。

每个 Thread 运行一个独立的 DSH ACP 会话，分别使用独立 Worktree。

然后由每个 DSH 会话内部使用自己的 Subagents 和 Workflows 完成工作。

不过，建议先用一个简单任务验证 DSH ACP 与 T3 Worktree 的配合：确认 DSH 能正确读取独立工作目录、执行测试、保持任务上下文，并且切换模型后行为正常。不要一开始就并行三个复杂功能。

## **七、我的最终判断**

从 AI 软件工程的架构角度看，T3 Code 值得研究的并不是它的聊天界面，而是它提供了一种独立于具体模型和 Harness 的 **Agent Control Plane**。

它解决的是：

> 当你拥有多个 Coding Agent 时，如何统一管理它们的运行环境、工作区、会话和代码交付？

而 DSH 解决的是：

> 如何组织 Agent、分解任务、协作执行，并完成复杂的软件工程工作？

两者解决的是不同层次的问题。

**我更倾向于你采用 DSH 做研发操作系统，T3 Code 做研发控制台。** 先通过 ACP 连接，验证真实项目中的兼容性，再决定是否将它纳入长期工具链。

相关入口：

- [T3 Code 官网](https://t3.codes/?utm_source=chatgpt.com)
- [T3 Code GitHub](https://github.com/pingdotgg/t3code?utm_source=chatgpt.com)
- [T3 Code ACP Provider 文档](https://github.com/pingdotgg/t3code/blob/main/docs/user/providers-acp.md?utm_source=chatgpt.com)
- [T3 Code 与 DeepSeek Harness 集成讨论](https://github.com/pingdotgg/t3code/discussions/34675?utm_source=chatgpt.com)

这套组合值得做一次小规模实测，尤其是 DSH 原生 Subagents 在 T3 中的可见性和 Worktree 隔离效果。

# **T3 Code 社区主流 Agent 搭配及优缺点**

先说我的判断：**T3 Code 最有价值的用法不是选出一个“最强 Agent”，而是把不同 Harness 组合成一个研发团队。**

从目前能看到的社区实践、T3 官方支持情况和工程团队的技术评估来看，值得重点关注的有五类：

1. **Claude Code**：复杂任务、架构分析和代码实现。
2. **OpenAI Codex**：工程实现、代码修改、测试和交付。
3. **OpenCode**：开源、模型自由切换、自定义 Harness。
4. **Pi Coding Agent**：轻量、可扩展、适合构建个人 Agent 工作流。
5. **DeepSeek Harness**：多 Agent 协作、Workflow 和模型混合路由。

这里需要先区分两个概念：

- **Agent / Harness**：例如 Claude Code、Codex、OpenCode、DSH。
- **模型**：例如 DeepSeek V4.1 Flash、Claude Haiku 5.5、GLM 5.3。

T3 Code 主要统一管理前者；不同 Harness 内部具体使用什么模型，则由各自的模型和 Provider 配置决定。

T3 官方目前列出的主要集成包括 Claude Code、Codex、Cursor、Grok、OpenCode、Antigravity、Pi 等，也支持 ACP Registry 接入其他 Agent。

[image](https://www.google.com/s2/favicons?domain=https://t3.codes\&sz=32)

T3 Code+2



## **一、社区中值得关注的五种 Agent**

### **1. Claude Code：复杂研发任务的主力**

[Claude Code in Production: 40% Productivity Increase on a Large Project - DEV Community](https://images.openai.com/static-rsc-4/2_fVAyjh2A7rIE9SxJrra2xCv4ZlpBBEPC9M4stzaB1bUAOTIGECnWZxuPRa23fPYhIVPWUS3do-g52ox0TQAMr429CjCkeQUgX_zMLlw-F0X8q0uJLyMxkzWozWi99IGc4f_p3ld6pg-HHL2mUqMtDP3nKHvvj6W4e0CiCoJTk?purpose=inline)

**定位：成熟的 Coding Harness**

复杂代码理解、跨文件修改、架构分析、工具调用

[官方文档](https://docs.anthropic.com/en/docs/claude-code/overview?utm_source=chatgpt.com) · [GitHub](https://github.com/anthropics/claude-code?utm_source=chatgpt.com)

**优点**

- 对大型代码仓库的理解和跨文件修改能力值得重点考虑。
- Agent 工具调用、终端操作、文件修改和代码验证形成了较成熟的闭环。
- 支持 Subagents、Skills、Hooks 等扩展机制。
- 适合复杂任务分解、代码重构和多步骤工程工作。

**缺点**

- 使用成本可能较高，尤其是长时间、多轮工具调用。
- 对模型订阅、使用额度和服务可用性有依赖。
- 如果没有严格限制任务范围，容易扩大修改范围。
- 需要维护 Claude 专属配置、规则和 Skills。

**适合在 T3 中承担：**

- 复杂架构改造。
- 跨模块功能开发。
- 复杂 Bug 分析。
- 重要代码审查。

我的判断：如果不考虑成本，Claude Code 是值得放在 T3 里的重要对照组。但你没有必要为了它改变目前的模型体系。

### **2. OpenAI Codex：工程执行与独立验证**

[🚨 AI News | TestingCatalog (@testingcatalog) on X](https://images.openai.com/static-rsc-4/c2zFI940OGngXaTnP1M7m6IHV93bVtdY17cAdejdSNVKKbzcrDy5_bmfQ2hOXi9wp43Oct70G6OHP85-a7OOl7zS1YkaD8ngKTccu1E52NYeZqiEewTXKHh6vl4NwwfW9JHOkRFYlvKN9N1lR5rHdZsOSuFRPwez6B7yJX1pAnI?purpose=inline)

**定位：工程任务执行型 Harness**

代码修改、测试、任务委派、隔离执行

[官方文档](https://developers.openai.com/codex/?utm_source=chatgpt.com) · [GitHub](https://github.com/openai/codex?utm_source=chatgpt.com)

**优点**

- 有比较完整的 Coding Agent 执行环境。
- 支持任务委派和并行开发。
- 沙箱、权限控制和审批机制是重要的工程优势。
- 适合独立执行一个定义清楚的 Issue，完成代码修改和测试。
- 可以作为与其他模型不同体系的交叉验证者。

**缺点**

- 与 OpenAI 模型和产品生态结合较深。
- 与其他 Harness 的配置、Skills 和执行行为不完全一致。
- 复杂任务仍然需要良好的需求和上下文。
- 如果已经有成熟的 DSH 开发流水线，额外引入 Codex 会增加配置和维护成本。

**适合在 T3 中承担：**

- 独立功能实现。
- Bug 修复。
- 测试补充。
- 独立代码审查。
- 对其他 Agent 的实现进行交叉验证。

我的判断：Codex 最大的价值，是提供一套不同于 DSH 的执行和验证环境。特别适合拿同一个任务做交叉实验。

### **3. OpenCode：模型自由度和可定制性**

[What Is Oh My OpenAgent (OMO)? Complete 2026 Guide to OpenCode Multi-Agent Orchestration Harness | Blog | a2a mcp](https://images.openai.com/static-rsc-4/N_twF1GO9SW6FueQ52YRjAHcCGWkTwWpXZH9o_jsnCLMDvXOVVma-gwYTbBAWgPgHCXzG6s9UeTIFy_ZikYcHkUHtTwoQ_vlJQc0fo1wK2ez0wXrhdhTQHXoAMA15Cla8UGPZ736R63XoJkNYDuNcDNbtZTUbAiAqSbipWjvQXA?purpose=inline)

**定位：开源、模型无关的 Coding Harness**

多 Provider、开源、自定义工具和配置

[官网](https://opencode.ai/?utm_source=chatgpt.com) · [GitHub](https://github.com/anomalyco/opencode?utm_source=chatgpt.com)

**优点**

- 开源，模型和 Provider 选择比较灵活。
- 支持多种模型服务，包括自托管和开放模型生态。
- 可扩展性强，适合定制工具和开发环境。
- 可以作为统一接入不同模型服务的 Coding Harness。
- 对希望减少单一模型厂商绑定的人比较有吸引力。

**缺点**

- 自由度高意味着配置和维护工作也更多。
- 不同 Provider 的工具调用、上下文和模型能力存在差异。
- 多 Agent 协作和跨 Harness 状态管理，需要额外设计。
- 其会话共享和协作能力在 Thoughtworks 2026 年 11 月的技术雷达中仍被指出有不足。

  [image](https://www.google.com/s2/favicons?domain=https://www.thoughtworks.com\&sz=32)

  Thoughtworks



**适合在 T3 中承担：**

- GLM、DeepSeek 等模型的 Coding 实验。
- 自定义模型路由。
- 本地模型开发。
- 需要特殊工具链的项目。

我的判断：如果你的核心目标是模型自由度，OpenCode 很有价值。但你已经使用 DSH，二者在部分能力上重叠，不建议没有明确需求就同时维护两套 Harness。

### **4. Pi Coding Agent：轻量和高度可定制**

[Package Catalog · Pi](https://images.openai.com/static-rsc-4/DQvJIoTaKGwQMm1pQMiHoUR7U99fcUY9ukwa4VGxoxjNfDkpfYs-nSHsm3d9sHbcIwqyOAi-8yuQXwJXATXIcPo9LfXlR2MzfGQJT5BmFwYNV7D47J1d5Zgbt1EP73zT7blREYb-o9ZhVaQ5s0AOAeHJlfcuxP-tqnskb8e_mU4?purpose=inline)

**定位：轻量级 Agent 基础设施**

极简核心、Extensions、Skills、自定义工作流

[GitHub](https://github.com/badlogic/pi-mono?utm_source=chatgpt.com)

**优点**

- 核心设计相对轻量。
- 扩展机制灵活，适合开发者自行定制。
- 可以根据需求构建自己的 Agent 工作流。
- 适合探索 Harness Engineering，而不是只使用现成的 Coding Agent。
- 在 T3 中有相应的原生集成。

**缺点**

- 很多高级工程能力需要自己组合。
- 需要投入时间维护 Extensions 和工作流。
- 对希望安装后立即拥有完整研发团队的人来说，初期体验可能不够完整。
- 不适合把大量工程流程都交给一个尚未搭建好的自定义扩展体系。

**适合在 T3 中承担：**

- 自定义研发 Agent。
- 自动化代码检查。
- 专用小工具和脚本开发。
- Harness 实验。

我的判断：Pi 更像一套可以自己搭建厨房的工具，而不是开箱即用的完整研发流水线。

### **5. DeepSeek Harness：你的主研发 Harness**

[DSH 插件商店 · DeepSeek Harness Plugin Store](https://images.openai.com/static-rsc-4/OWFQ5VWTm9zYAyN6GKcIRwnCYp2oIe_wSYod778i4l1cilS-lAVUi0KbXd0r_rsrFMSXqurwv6E4EEujSbZO6BFOxyuEY-Mf0rrVBGW1dSUbTMWoEeRl_1yWor9wZsFYcLqgPioSMpviZKE8b-cFY2Yb-cFPG2QbnxpNa4lVD18?purpose=inline)

**定位：多 Agent 与 Workflow 编排**

原生 Agent、Subagents、Workflows、Plugins、模型路由

[官网](https://deepseek.com/en/harness/?utm_source=chatgpt.com) · [GitHub](https://github.com/DeepSeek-ai/deepseek-harness?utm_source=chatgpt.com)

**优点**

- 原生支持你目前最重视的 Agent、Subagents 和 Workflows。
- 能够组合多个模型完成不同任务。
- 适合把软件工程流程固化成可重复执行的工作流。
- 可以把个人研发团队的组织方式沉淀为项目资产。
- 对你当前的 DeepSeek、Claude Haiku、GLM 模型组合有直接价值。

**缺点**

- 生态和成熟度仍需持续观察。
- 通过 ACP 接入其他控制台时，不一定能完整暴露 DSH 的内部能力。
- 工作流和 Agent 定义需要自己建立工程规范。
- 如果把复杂任务全部交给一个主 Agent，仍然会遇到上下文、状态和恢复问题。

**适合在 T3 中承担：**

- 产品需求分析。
- 架构和任务规划。
- 多 Agent 协作。
- 软件工程 Workflow。
- 多模型路由和研发自动化。

我的判断：对你而言，DSH 不应该只是五种 Agent 中的一种，而应该是研发组织的核心。

## **二、社区实际在怎么搭配？**

这里我不想把“社区最佳”简单理解成 GitHub Star 数量。更有参考价值的是看真实用户如何组合，以及为什么这样组合。

我找到一个比较有代表性的 Hacker News 用户实践：一位用户在个人电脑和 VPS 上统一使用 T3 Code，底层组合是：

- Claude Code：使用 Claude 订阅。
- Codex：使用 OpenAI 订阅。
- Pi：使用 OpenCode Go 订阅。

他会根据任务复杂程度调整模型推理强度：复杂架构和大型功能使用更高推理级别，普通实现则拆成小任务完成。另一位参与讨论的用户也特别认可 T3 的统一工作流、跨 Harness 协作和远程控制。

[image](https://www.google.com/s2/favicons?domain=https://news.ycombinator.com\&sz=32)

Hacker News



这不是大规模用户统计，但它提供了一个很实用的组合思路。

### **四种值得借鉴的搭配**

**组合 A：Claude Code + Codex**

综合型

Claude 负责复杂分析和方案，Codex 负责实现、测试和独立验证。

优点：两个不同生态的 Harness 互相补充，适合复杂工程。\
缺点：成本较高，仍需要统一需求和接口。

**组合 B：Codex + OpenCode**

开源与工程型

Codex 执行工程任务，OpenCode 连接不同模型服务，做补充开发和交叉审查。

优点：模型选择自由度较高。\
缺点：两个 Harness 的配置、上下文和工具能力需要协调。

**组合 C：Claude Code + Codex + Pi**

多 Harness 型

复杂任务、工程执行、轻量自定义 Agent 分工。

优点：覆盖面广，适合多个项目和多种开发习惯。\
缺点：维护成本和工具重叠明显增加。

**组合 D：DSH + Codex + Claude Code**

适合你的方向

DSH 负责研发组织，Codex 负责独立工程执行，Claude Code 作为复杂任务和审查的备用 Harness。

优点：保留 DSH 的多 Agent / Workflow 能力，同时引入独立执行环境。\
缺点：需要验证 ACP、跨 Thread 委派和工作区隔离。

特别值得注意的是，T3 Code 的 Orchestrator V2 在 2026 年 11 月的稳定版中增加了 `delegate_task`：一个 Agent 可以委派子任务给不同 Provider 或模型，每个子任务作为独立 Thread 运行，父 Agent 可以等待并汇总结果。官方给出的典型例子就是 Claude 负责规划、Codex 负责实现。

[image](https://www.google.com/s2/favicons?domain=https://newreleases.io\&sz=32)

New Releases



这让 T3 从单纯的多 Agent 展示界面，进一步向跨 Harness 协作控制平面发展。

## **三、各 Agent 的核心优缺点横向比较**

下面是我的工程选型判断，不是统一基准测试排名。

| 维度           | Claude Code    | Codex  | OpenCode | Pi       | DSH      |
| ------------ | -------------- | ------ | -------- | -------- | -------- |
| 复杂工程任务       | 强              | 强      | 取决于模型    | 取决于模型和扩展 | 取决于模型和编排 |
| 开源和可定制       | 较有限            | 部分开源组件 | 强        | 强        | 强        |
| 多模型自由度       | 有限             | 有限但在扩展 | 强        | 强        | 强        |
| 多 Agent 编排   | 强              | 强      | 支持       | 可扩展      | 核心能力     |
| 自定义 Workflow | Hooks/Skills 等 | 任务和自动化 | 可扩展      | 高度可扩展    | 核心能力     |
| 开箱即用         | 较好             | 较好     | 较好       | 偏轻量      | 需要配置     |
| 维护复杂度        | 中              | 中      | 中高       | 中高       | 中高       |

一个重要补充：Thoughtworks 2026 年 11 月的技术雷达将 Claude Code、Codex、OpenCode 列入 Adopt。其评价也特别提到 Codex 的沙箱、审批与企业治理能力，以及 OpenCode 的开源、可定制和自托管模型优势。

[image](https://www.google.com/s2/favicons?domain=https://www.thoughtworks.com\&sz=32)

Thoughtworks



因此，我不会把 OpenCode 或 Pi 简单归类为弱于 Claude Code、Codex。它们的价值更多体现在自由度和可定制性，而不是单纯的模型能力。

## **四、结合你的 DSH 和三个模型，我会怎么选？**

我的建议是：**第一阶段只接入三个 Harness，不要把五个都装上。**

## **推荐组合：DSH + Codex + Claude Code**

**DeepSeek Harness**

主力

# **研发总控**

- DeepSeek V4.1 Flash：需求、规划、协调
- GLM 5.3：代码实现
- Claude Haiku 5.5：辅助审查、测试设计
- DSH Subagents：专业任务执行
- DSH Workflows：研发流程编排

**Codex**

第二执行引擎

- 独立实现复杂功能
- 修复 DSH/GLM 未解决的问题
- 独立运行测试
- 对关键变更进行交叉验证

**Claude Code**

备用专家

- 复杂代码库分析
- 跨模块架构改造
- 重要技术方案审查
- 复杂 Bug 定位

这里有个细节：你目前使用的是 Claude Haiku 5.5，并不意味着接入 Claude Code 后就自动拥有相同的模型配置。Claude Code 的模型和认证由 Anthropic 自身的 Provider 管理，需要单独确认账号、模型可用性和额度。

### **为什么不优先推荐 OpenCode？**

因为你已经有 DSH。

两者都具备模型接入、工具调用和 Agent 执行能力。如果只是为了多一个 Coding Agent 而引入 OpenCode，收益可能不够覆盖维护成本。

但如果你未来希望：

- 大规模使用自托管模型；
- 通过不同 Provider 运行 GLM、DeepSeek；
- 自己构建 Coding Harness；
- 深度定制模型调用、工具和执行流程；

那么 OpenCode 就值得重新评估。

### **为什么暂时不推荐 Pi？**

Pi 很适合喜欢自己搭建 Agent 的工程师，但你现在的核心目标是完整软件工程工作流，而不是再造一个 Harness。

DSH 已经承担了这个职责。Pi 可以留作未来探索项目。

## **五、T3 Code 里最值得实验的三个工作模式**

### **模式 1：DSH 主导，Codex 执行**

**DSH：分析需求、设计方案、拆解任务**

**Codex：独立执行开发任务**

**DSH + Claude：审查实现和验证结果**

**CI：自动测试和质量门禁**

适合 KubeInference 的正式功能开发。

### **模式 2：三个 Agent 并行分析**

例如开发 GPU 资源管理功能：

- DSH：分析产品需求和整体设计。
- Claude Code：分析 Kubernetes GPU 资源管理和架构风险。
- Codex：独立研究实现方案和测试策略。

最终由 DSH 汇总，形成技术决策材料。

适合复杂需求和架构研究。

### **模式 3：不同 Harness 交叉修复**

例如 GLM 5.3 完成了 Controller，但 Kind E2E 测试失败：

1. DSH/GLM 先尝试定位和修复。
2. 如果连续两次修复失败，暂停自动重试。
3. 把错误日志、代码 Diff、测试步骤交给 Codex。
4. Codex 独立分析，提出修复方案。
5. Claude 或 DSH 检查修复是否引入新问题。

这种方式比三个 Agent 同时修改同一份代码更可靠。

## **六、最终选型建议**

## **推荐优先级**

**1. DSH**

必须保留

研发组织、模型路由、Subagents、Workflows。

**2. Codex**

优先试用

提供独立的工程执行与交叉验证能力。

**3. Claude Code**

按需引入

复杂代码分析和高难度任务的备用 Harness。

**4. OpenCode**

未来可选

当模型自由度和自托管成为核心诉求时引入。

**5. Pi**

探索性

用于自定义 Agent 和 Harness 实验。

还有一个需要特别留意的 T3 新变化：当前 ACP 文档指出，第三方 ACP Agent 的子 Agent 层级展示依赖各 Agent 自己暴露的信息，ACP 并没有统一的原生子 Agent 血缘协议。T3 自己的 `delegate_task` 可以跨 Provider 创建子任务，但这不意味着 DSH 内部的 Subagents 自动完整映射到 T3 的界面。

[image](https://www.google.com/s2/favicons?domain=https://github.com\&sz=32)

GitHub+1



**我的核心建议：你不需要把 T3 变成一个 Agent 大卖场，而应该把它变成 DSH 研发团队的统一操作台。**

先让 DSH + Codex 共同完成 KubeInference 的一个真实功能，再决定 Claude Code 是否值得长期加入。这样能用实际任务的成功率、返工率和成本来决定工具组合，而不是凭社区热度选型。