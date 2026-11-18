# kready 检查清单 v0（草案）

> 只收录满足四条标准的项：只读、确定性、期望值明确、失败可直接给出动作。
> 期望值来源：`scclabs-k8s-in-action/k8s`（kubespray 2.32.0 文档与 os/ 目录）。
> 每条只陈述期望状态；不通过即为异常。

## A. 操作系统（每个节点）

| ID | 检查项 | 期望 | 来源 |
|---|---|---|---|
| OS-01 | 内核参数 `fs.inotify.max_user_instances` | `1280` | `os/80-inotify.conf` |
| OS-02 | 内核参数 `fs.inotify.max_user_watches` | `655360` | `os/80-inotify.conf` |
| OS-03 | 内核模块 `br_netfilter` 已加载 | 已加载 | `new-cluster-config-guide.md` §6 / 设计要求 |
| OS-04 | 参数 `net.bridge.bridge-nf-call-iptables` | `1` | 同上 |
| OS-05 | 参数 `net.ipv4.ip_forward` | `1` | 同上 |
| OS-06 | swap 已关闭 | 0 B | kubelet 要求，`os/README.md` 未涉及，待文档补充 |
| OS-07 | cgroup 版本 | v2 | `new-cluster-config-guide.md` §2.1 |
| OS-08 | 容器运行时 | containerd | `new-cluster-config-guide.md` §2.1 |
| OS-09 | 时间同步服务 `systemd-timesyncd` 或 `chrony` 运行中且已同步 | 已同步 | `new-cluster-config-guide.md` §2.1 |
| OS-10 | sshd 禁用密码登录 | `PasswordAuthentication no` | `os/50-cloud-init.conf` |
| OS-11 | 内核自动升级已停用（unattended-upgrades 未启用） | 未启用 | `os/noupgrade.sh` |

## B. 版本与组件（集群级）

| ID | 检查项 | 期望 | 来源 |
|---|---|---|---|
| VER-01 | 所有节点 kubelet 版本 | `v1.36.4` | `base-components-and-versions.md` §三 |
| VER-02 | 所有节点 containerd 版本 | `2.3.5` | 同上 §四 |
| VER-03 | etcd 版本 | `3.6.14` | 同上 §三 |
| VER-04 | kube-proxy 模式 | `nftables` | `k8s-cluster.yml` `kube_proxy_mode` |
| VER-05 | 节点操作系统 | Ubuntu 24.04 | 实验环境基线 |

## C. 控制面与 kube-system（集群级）

| ID | 检查项 | 期望 | 来源 |
|---|---|---|---|
| SYS-01 | 控制面节点数 | 3（奇数，≥3） | `new-cluster-config-guide.md` §2.1 |
| SYS-02 | kube-apiserver / kube-controller-manager / kube-scheduler 各 Pod 数 | 等于控制面节点数，全部 Ready | `new-cluster-config-guide.md` §7.1 |
| SYS-03 | kube-system 中 Pod 为 Ready 且无 CrashLoopBackOff | 全部 Ready | 通用 |
| SYS-04 | kube-system 中近 15 分钟内容器重启次数 | 0 | `new-cluster-config-guide.md` §8 |
| SYS-05 | kube-vip Pod 数 | 等于控制面节点数，且为 mirror Pod（可见于 `kubectl get po`） | `fixes/README.md` |
| SYS-06 | nginx-proxy Pod 数 | 等于工作节点数，且为 mirror Pod | `fixes/README.md` |
| SYS-07 | CoreDNS Pod | Ready，副本数 ≥ 2 | `new-cluster-config-guide.md`（`coredns_replicas`） |
| SYS-08 | Cilium / kube-proxy / nodelocaldns DaemonSet | 每个节点各一个 Pod，全部 Ready | 通用（与节点数一致） |
| SYS-09 | metrics-server 副本数 | `3` | `addons.yml` `metrics_server_replicas` |

## D. 配置终态

| ID | 检查项 | 期望 | 来源 |
|---|---|---|---|
| CFG-01 | kube-controller-manager 内存 request 与 limit | 均已设置（非空） | `new-cluster-config-guide.md` §4 |
| CFG-02 | kube-controller-manager 存在 startupProbe | 已配置 | 同上 §4.2 |
| CFG-03 | etcd 后端配额 `etcd_server_quota_backend_bytes` | `8589934592` | `etcd.yml` |
| CFG-04 | CoreDNS 上游 DNS | 与 `upstream_dns_servers` 一致 | `all.yml` |
| CFG-05 | 证书自动续期 | 与文档一致（生产要求为 `false`，或已完成串行编排） | `new-cluster-config-guide.md` §2.2、§10 |

## E. 对象规模（信息项，不判定）

- 节点数、Pod 数、CRD 数、EndpointSlice 数、Namespace 数。
- 仅输出当前值，不做通过或失败判定。

## F. 网络策略与虚拟集群（集群级）

| ID | 检查项 | 期望 | 来源 |
|---|---|---|---|
| NET-01 | `policyTypes` 含 `Egress` 的 NetworkPolicy 所选中、但未被放行 53 的 Pod | 空集（凡被 Egress 策略选中的 Pod 都应能解析 DNS） | 2026-11-18 `vc0168a` 故障 |
| NET-02 | 放行 53 的规则写法 | 规则**只写 `ports`、不写 `to` / `ipBlock`** | 同上；`sd1-maas-dev-deploy/compute/vcluster/*/values.yaml` 注释 |
| NET-03 | vcluster 命名空间中 `vcluster.loft.sh/managed-by=<vcluster 名>` 的同步 Pod → 本 vcluster 控制面 | 放行 443/8443（`to: podSelector` 命中 `app=vcluster`） | 同上 |

NET-02 的原因：本环境 CNI 为 Cilium（`k8s-cluster.yml` `kube_network_plugin: cilium`），宿主机 node-local-dns（`nodelocaldns_ip: 169.254.25.10`）在 Cilium 里是 host identity，任何带 `to` / `ipBlock` 的 53 规则都匹配不到它，等于没放行；只有“仅 `ports`”的规则才生效。

### SYS-07 失败时的故障指纹（2026-11-18 `vc0168a` 实例）

```
[INFO] plugin/kubernetes: waiting for Kubernetes API before starting server
[INFO] plugin/ready: Still waiting on: "kubernetes"
[ERROR] plugin/kubernetes: ... failed to list *v1.Namespace: Get "https://<ClusterIP>:443/api/v1/namespaces?...": dial tcp <ClusterIP>:443: i/o timeout
[ERROR] plugin/errors: ... HINFO: read udp <podIP>:...->169.254.25.10:53: i/o timeout
```

判读：**集群内目标（Service ClusterIP、`169.254.25.10`）与公网 DNS 同时大面积 `i/o timeout`** ⇒ 先查该命名空间 `policyTypes` 含 `Egress` 的 NetworkPolicy（NET-01~03），不要先去查 DNS 配置或机房网络；若只超时公网 DNS，才按 CFG-04 查出口与上游。

处理：给该策略补 53（仅 `ports`）与控制面 443/8443 放行；参考 `zg3-deploy/vcluster/vc0168a/hooks/networkpolicy.yaml`。

---

### 已排除（不进入工具）

- etcd fio 测试、控制面重启演练、删除风暴、kcm 参数调优效果：需要执行测试，不是只读检查。
- 网卡名、VIP 地址、DNS 地址是否与机房一致：需要人确认的事实。
- etcd WAL fsync p99：需要压测或长期监控，不在部署后单次只读检查范围。

### 待确认（需要你决定）

1. **CFG-05 的期望值**：文档对"证书续期"是二选一，工具应判定哪一种？建议先按 `false` 判定，否则需要区分两种合法状态。
2. **OS-06（swap）、OS-03~05（br_netfilter 与转发）** 在文档中没有明确条款，目前只来自 kubelet/Cilium 的通用要求。是否纳入需确认。
3. **SYS-09（metrics-server=3）、SYS-07（CoreDNS≥2）** 是 profile 的具体取值，若不同集群取值不同，应改为从 profile 读取。
4. **VER-05（Ubuntu 24.04）** 是本环境基线，不是通用要求，建议从清单中移出，或改为 profile 参数。
5. **NET-01~03 的生效范围**：按“存在 Egress NetworkPolicy 的命名空间”通用判定，还是只覆盖 vcluster 命名空间（按 `vcluster.loft.sh/managed-by` 收窄）？
