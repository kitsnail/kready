# 需求清单（backlog）

> 维护者：人（产品经理）。Agent 仅按 `docs/product/README.md`「字段级权限」写入（D-011）。
> 排列：按优先级，从高到低。每条需求使用 `REQ-nnn` 编号，编号不复用。

## 条目模板

复制下面的模板，按编号顺序追加到文件末尾。

```markdown
### REQ-000 <一句话标题>

- 状态：草稿 / 待开始 / 进行中 / 待验收 / 已验收 / 已取消
- 优先级：高 / 中 / 低
- 背景：为什么要做（一两句）
- 期望结果：用户能做到什么
- 验收标准：见 `acceptance.md` 中的 REQ-000
- 关联：规则 ID、PR、CHANGELOG 条目（由 Agent 回填）
```

## 需求列表

### REQ-001 GPU 节点基础检查（device plugin 与卡数标签）

- 状态：已验收
- 优先级：高
- 背景：GPU 节点的交付就绪缺少可自动判定的检查；`--pack gpu` 目前无规则，输出"正常"易被误读为 GPU 已检查通过。
- 期望结果：`kready check --pack gpu` 对 GPU 节点报出 device plugin 就绪、可分配卡数与标签一致、卡数标签存在三项判定与污点信息项；规则包无 active 规则时结论输出「未运行任何规则」。
- 验收标准：见 `acceptance.md` 中的 REQ-001
- 关联：GPU-01～GPU-04、D-012（规则包骨架）、D-014（期望值批准）、PR #21

### REQ-002 GPU 驱动加载检查

- 状态：已验收
- 优先级：中
- 背景：驱动未加载时 GPU 节点不可交付，需要只读、可自动判定的检查。
- 期望结果：GPU 节点上经 SSH 只读执行 `nvidia-smi -L`，退出码 0 判正常，非 0 判异常。
- 验收标准：见 `acceptance.md` 中的 REQ-002
- 关联：GPU-05、D-006（SSH 只读采集）、D-012、D-014（期望值与白名单批准）、PR #21、PR #22

### REQ-003 GPU-01 适配 GPU Operator 部署

- 状态：已验收
- 优先级：高
- 背景：GPU-01 选择器只覆盖独立 Helm 部署的 device plugin，GPU Operator 集群误报「未部署 FAIL」（sd1 测试环境实测复现）。
- 期望结果：GPU Operator 与独立 Helm 两种部署方式下，GPU-01 均能正确判定 device plugin 就绪状态。
- 验收标准：见 `acceptance.md` 中的 REQ-003
- 关联：GPU-01、D-014、D-015、PR #24

### REQ-004 控制面规模与组件检查（CP-02～CP-05）

- 状态：草稿
- 优先级：高
- 背景：草案 CP-02～05（etcd 成员数 3、apiserver/kcm/scheduler Pod 数等于控制面节点数且全部 Ready、kube-vip 与 nginx-proxy 镜像 Pod 数）已定稿且采集路径清晰（快照 Pod 清单 + 节点角色标签），但未实现，控制面规模缺陷目前检不出。
- 期望结果：`kready check` 对 CP-02～05 报出判定结果；控制面规模或组件 Pod 异常时报 FAIL 并列出实际值与期望值。
- 验收标准：见 `acceptance.md` 中的 REQ-004
- 关联：CP-02、CP-03、CP-04、CP-05（规则条目与期望值需人批准后写入 `rules/`）

### REQ-005 控制面资源与 etcd 配额检查（CP-06～CP-08）

- 状态：草稿
- 优先级：高
- 背景：草案 CP-06～08（kcm 内存 request/limit 均设置、kcm startupProbe 已配置、etcd 后端配额 8 GiB）需快照扩展 Deployment/Pod 资源字段，实现后控制面资源配置缺陷可自动检出。
- 期望结果：`kready check` 对 CP-06～08 报出判定结果；资源配置缺失或配额不符时报 FAIL。
- 验收标准：见 `acceptance.md` 中的 REQ-005
- 关联：CP-06、CP-07、CP-08（规则条目与期望值需人批准后写入 `rules/`）

### REQ-006 系统组件健康检查（SYS-01～SYS-03）

- 状态：草稿
- 优先级：高
- 背景：草案 SYS-01～03（kube-system Pod 全部 Ready 无 CrashLoopBackOff、近 15 分钟容器重启 0 次、Cilium/kube-proxy/nodelocaldns DaemonSet 每节点一个 Pod 且 Ready）是交付验收最直接的信号，目前未实现。
- 期望结果：`kready check` 对 SYS-01～03 报出判定结果；系统 Pod 异常、重启或 DaemonSet 缺副本时报 FAIL。
- 验收标准：见 `acceptance.md` 中的 REQ-006
- 关联：SYS-01、SYS-02、SYS-03（规则条目与期望值需人批准后写入 `rules/`）

### REQ-007 NTP 结果判定（POL-03a/POL-03b）

- 状态：草稿
- 优先级：中
- 背景：D-017 裁定 POL-03 采用结果判定：POL-03a 每节点 NTP 同步状态 = `yes`，POL-03b 节点间时钟偏差 ≤ `1000ms`；需白名单新增 2 条命令（`timedatectl show --property=NTPSynchronization --value`、`date +%s%3N`）与新判定类型（节点间 max−min）。
- 期望结果：NTP 未同步的节点报 FAIL（POL-03a）；节点间时钟偏差超 1000ms 报 FAIL 并列出偏差值（POL-03b）；NTP 服务器地址不检查。
- 验收标准：见 `acceptance.md` 中的 REQ-007
- 关联：POL-03a、POL-03b、D-017（白名单命令与期望值需人批准）

### REQ-008 OS 内核参数检查（OS-01～OS-03）

- 状态：草稿
- 优先级：中
- 背景：D-017 裁定纳入 OS 内核参数检查：OS-01 `br_netfilter` 已加载、OS-02 `net.ipv4.ip_forward=1`、OS-03 swap 关闭；需白名单新增命令（`lsmod`/`sysctl`/`swapon --show` 或等价只读方式）。
- 期望结果：模块未加载、转发未开启或 swap 未关闭时报 FAIL 并列出实际值与期望值。
- 验收标准：见 `acceptance.md` 中的 REQ-008
- 关联：OS-01、OS-02、OS-03、D-017（采集命令与期望值需人批准）
