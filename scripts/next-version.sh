#!/usr/bin/env bash
# 按 docs/design/versioning.md 第六节计算下一个 kready 版本号。
#
# 只读：不写文件、不打 tag。输出版本号与依据，供人复核。
#
# 规则：
#   - 没有任何 tag：首个版本为 v0.1.0（不计算增量）。
#   - 0.x 阶段：破坏性变更 -> MINOR；feat、fix -> PATCH；其它类型不升版本。
#   - 1.0 之后：破坏性变更 -> MAJOR；feat -> MINOR；fix -> PATCH。
#
# 破坏性变更识别两种写法：类型后带 "!"（如 feat(output)!:），或提交正文含 "BREAKING CHANGE:"。
#
# 用法：scripts/next-version.sh [--json]

set -euo pipefail

FIRST_VERSION="0.1.0"
JSON=0
if [[ "${1:-}" == "--json" ]]; then
  JSON=1
elif [[ $# -gt 0 ]]; then
  echo "用法：$0 [--json]" >&2
  exit 2
fi

if ! git rev-parse --git-dir >/dev/null 2>&1; then
  echo "错误：不在 git 仓库中" >&2
  exit 2
fi

# 最近的 vX.Y.Z tag；没有则为空。
last_tag="$(git describe --tags --abbrev=0 --match 'v[0-9]*.[0-9]*.[0-9]*' 2>/dev/null || true)"

if [[ -z "$last_tag" ]]; then
  base_version="$FIRST_VERSION"
  range="HEAD"
  major=0; minor=0; patch=0
  # 首个版本：直接输出起始版本，不做增量计算。
  if [[ "$JSON" == 1 ]]; then
    printf '{"version":"v%s","base":"","reason":"无 tag，首个版本","commits":0}\n' "$base_version"
  else
    echo "v$base_version"
    echo "依据：无 tag，首个版本。" >&2
  fi
  exit 0
fi

base_version="${last_tag#v}"
IFS='.' read -r major minor patch <<<"$base_version"
range="${last_tag}..HEAD"

# 逐条提交解析：提取类型、破坏性标记。
# 以 NUL 分隔每条提交，避免正文中的换行干扰。
has_breaking=0
has_feat=0
has_fix=0
commit_count=0

while IFS= read -r -d '' msg; do
  commit_count=$((commit_count + 1))
  subject="${msg%%$'\n'*}"
  # 类型：行首的字母，可带 (scope)，可带 !，然后是冒号。
  # 类型正则放入变量，避免 [[ =~ ]] 中括号转义问题。
  re='^([a-z]+)(\([^)]*\))?(!)?:[[:space:]]'
  if [[ "$subject" =~ $re ]]; then
    type="${BASH_REMATCH[1]}"
    bang="${BASH_REMATCH[3]}"
    [[ -n "$bang" ]] && has_breaking=1
    case "$type" in
      feat) has_feat=1 ;;
      fix)  has_fix=1 ;;
    esac
  fi
  if printf '%s' "$msg" | grep -q '^BREAKING CHANGE:'; then
    has_breaking=1
  fi
done < <(git log --format='%B%x00' "$range")

bump=""
changed=0
if [[ "$commit_count" == 0 ]]; then
  reason="自 ${last_tag} 起无新提交，版本不变"
  next="$last_tag"
else
  if (( major == 0 )); then
    # 0.x 阶段
    if (( has_breaking )); then
      minor=$((minor + 1)); patch=0; bump="MINOR（0.x 破坏性变更）"; changed=1
    elif (( has_feat || has_fix )); then
      patch=$((patch + 1)); bump="PATCH（0.x 功能或修复）"; changed=1
    else
      bump="无版本变化（无 feat、fix 或破坏性变更）"
    fi
  else
    # 1.0 之后
    if (( has_breaking )); then
      major=$((major + 1)); minor=0; patch=0; bump="MAJOR（破坏性变更）"; changed=1
    elif (( has_feat )); then
      minor=$((minor + 1)); patch=0; bump="MINOR（新功能）"; changed=1
    elif (( has_fix )); then
      patch=$((patch + 1)); bump="PATCH（修复）"; changed=1
    else
      bump="无版本变化（无 feat、fix 或破坏性变更）"
    fi
  fi
  if (( changed )); then
    next="v${major}.${minor}.${patch}"
  else
    next="$last_tag"
  fi
  reason="${bump}；自 ${last_tag} 起 ${commit_count} 条提交"
fi

if [[ "$JSON" == 1 ]]; then
  printf '{"version":"%s","base":"%s","reason":"%s","commits":%s}\n' "$next" "$last_tag" "$reason" "$commit_count"
else
  echo "$next"
  echo "依据：$reason" >&2
fi
