# 历史记录书写规范
---

1. 历史记录文件放在模块根目录下，文件名为 `HISTORY.md`。

2. 书写格式参照 https://raw.github.com/allmobilize/amui/master/HISTORY.md

3. 一切有价值的修改都应该切实记录在文件中，推荐关联上对应的 issue 地址。

4. Amaze UI 的修改类型共有五项：
    - `NEW` [#3](##) 新增的属性、功能、方法、特性等等。
    - `FIXED` [#15](##) 修复 bug 和影响使用的性能问题等。
    - `IMPROVED` 接口增强、健壮性和性能提升、代码优化、依赖模块升级等。
    - `CHANGED` 涉及到兼容性变化的改动。
    - `UNRESOLVED` 已知的但本版本暂未修复的问题。