# Gitusr Context

gitusr 管理 git 用户身份,并通过 shell hook 在 clone/commit/cd 时按规则自动应用身份。

## Language

**Host 规则（Host rule）**:
将仓库 host 或通配模式映射到已保存用户邮箱的路由条目；规则以有序集合维护，顺序决定同级别下的匹配优先级。
_Avoid_: host 映射、host 路由表

**通配匹配（Wildcard match）**:
`*.byted.org` 形式的后缀规则，匹配任意深度的子域；匹配级别低于精确匹配，裸域规则（如 `byted.org`）只匹配自身。

**先配置者胜（First-rule-wins）**:
多个同级别规则匹配同一 host 时，配置靠前的规则生效；精确匹配永远先于通配匹配被评估。

**克隆后身份裁决（Clone identity resolution）**:
`git clone` 完成后为该仓库确定 git 用户的流程，优先级依次为：显式 `--gu-*` 参数 > 仓库内 `.gitusrrc` > host 规则 > 交互式选择。

**Host 路由配置**:
host → 用户邮箱的规则集合，由 gitusr 统一增删调序；运行时若规则引用的用户已不存在，输出警告并跳过，而不是失败。
