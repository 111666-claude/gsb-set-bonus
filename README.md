# set-bonus

装备套装加成：按套装计数判档、跨套装累加，换装只该影响相关套装，只用 Go 标准库。

```
go test ./...
go run ./cmd/set-bonus --sample mixed
go run ./cmd/set-bonus --sample pair
go run ./cmd/set-bonus --sample full
go run ./cmd/set-bonus --sample reslot
go run ./cmd/set-bonus --sample cost
```

## 口径（README 为准）

- **一个槽位只有一件装备**：`Equip` 换装必须替换同槽位的旧件，`Pieces(set)` 只数当前还在身上的件数。
- **件数按套装各算各的**：只有同一套装的件数才算，不同套装的件数不能相加。
- **刚好够件数就生效**：判定用 `Pieces >= 档位件数`。
- **同一套装只拿最高档**：同时满足多个档位时只取件数最大的那一档，低档不叠加；
  多个套装同时达标则各自累加；`active` 按字典序输出。
- **换装增量更新**：`Equip`/`Unequip` 只更新受影响的套装统计，`Active()` 不许重扫全部装备；
  `Recomputed()` 是 `Active()` 累计重算过的件数，必须是 0。
- **规模**：5 万件装备、200 个套装的档位表、每秒 10 万次换装与 10 万次查询，
  内存 O(套装数 × 档位数)；不许按件数展开。

## 输出契约（不改格式）

```
attack=0 defense=0 active=none
attack=10 defense=0 active=lion:2
attack=40 defense=20 active=lion:6
lion_pieces=2 attack=10 defense=0 active=lion:2
recomputed=0
```

## 目录

```
bonus.go                套装加成
cmd/set-bonus           命令行入口
bonus_test.go           go test 用例
```
