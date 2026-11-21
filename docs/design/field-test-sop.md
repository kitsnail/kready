# 真实环境实测规范（field-test SOP）

> 状态：标准。在真实测试集群上验证 kready 的固定流程，每次版本验证重复使用。
> 本地自动化测试（`make check`）不在此列，见 AGENTS.md。

## 一、原则

- 只读：kubectl 仅 get/list，SSH 仅白名单只读命令。节点可能跑生产负载，不得做破坏性操作。
- 前置：`make check` 通过后才实测；实测发现的问题回流为修复或规则草案「待确认」条目。

## 二、前置条件

| 项 | 要求 | 检查方式 |
|---|---|---|
| 测试机 | 集群内节点，root 登录 | `ssh <测试机>` |
| kubectl | 默认上下文指向目标集群 | `kubectl get no` |
| SSH 采集 | 到全部节点的 root 免密（`~/.ssh/id_*` 或 agent），known_hosts 已含节点 | `ssh <节点名> true` |
| 二进制 | 本地交叉编译 linux/amd64 | 见部署标准 |

## 三、部署标准

所有文件统一放 `/root/.kready`（目录标准见 `kready-home.md`）：

1. 编译：`GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.version=$(git describe --tags --always)" -o kready ./cmd/kready`
2. 传输：`scp kready <测试机>:/root/.kready/bin/kready`，`chmod 700`
3. 目录布局 `bin/ config/ logs/ reports/ snapshots/ tmp/`，权限 0700（运行时也会自建）
4. 验证：`/root/.kready/bin/kready version` 输出版本号

## 四、测试流程（每次重复）

1. **live 默认包**：`kready check --live`，记录结论与退出码。
2. **live GPU 包**：`kready check --live --pack gpu`，记录结论与退出码。
3. **快照重放**：对两次 live 保存的快照分别离线重跑，结论与退出码须与 live 一致（判定可复现）。快照不记录包选择，重放须带与 live 相同的包参数：
   - 默认包：`kready check --snapshot <快照目录>`
   - GPU 包：`kready check --snapshot <快照目录> --pack gpu`
4. **产物核对**：`reports/` 有对应报告，`snapshots/` 有时间戳目录；同秒冲突加 `-2` 后缀，不覆盖旧文件。
5. **结果对照**：与第五节基线表逐条对照，偏差须解释——集群新发现记入基线备注；工具缺陷走修复 PR；期望值疑问记入规则草案「待确认」。

## 五、测试集群基线（历次实测的预期结论）

| 集群 | 日期 | 版本 | 2.32 包 | gpu 包 | 备注 |
|---|---|---|---|---|---|
| sd1-dev1-maas | 2026-11-21 | v0.1.0-46-gd3ff985 | 异常（NODE-02/03 inotify 为发行版默认值；NODE-04 控制面 cpu02-04 密码登录未禁用） | 正常 | NODE-02/03 裁定中（2.32-draft 待确认 6）；首次按本规范部署于 cpu02 |
| sd1-dev1-mlops | 2026-11-21 | v0.1.0-44-g4ffd3ab | 异常（NODE-02/03 同上，其余通过） | 正常（全部 SKIP：无 GPU 节点） | 快照模式验证，未部署 |

## 六、发现记录

- 集群问题（如 NODE-04）：记入基线表备注，处置由人决定。
- 工具缺陷：开 `fix/` 分支走 PR 流程。
- 规则期望值疑问：记入 `docs/rules/<pack>-draft.md`「待确认」。
