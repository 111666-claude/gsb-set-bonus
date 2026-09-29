# set-bonus

装备套装加成：按同一套装的件数判档、跨套装累加，只用 Go 标准库。

```
go test ./...
go run ./cmd/set-bonus --sample mixed
go run ./cmd/set-bonus --sample pair
go run ./cmd/set-bonus --sample full
```

## 口径（README 为准）

- **件数按套装各算各的**：只有同一套装的件数才算，不同套装的件数不能相加。
  `--sample mixed`（四个不同套装各一件）按口径是 `attack=0 defense=0 active=none`，
  现在是 `attack=18 defense=23 active=lion:2,moon:2,star:2`。
- **刚好够件数就要生效**：判定用 `Pieces >= 档位件数`。
  `--sample pair`（同一套装正好两件）按口径是 `attack=10 defense=0 active=lion:2`，
  现在是 `attack=0 defense=0 active=none`。
- **同一套装只拿最高档**：一个套装同时满足多个档位时只取件数最大的那一档，低档不叠加；
  多个套装同时达标则各自累加。`--sample full`（同一套装六件）按口径是
  `attack=40 defense=20 active=lion:6`，现在是 `attack=43 defense=33 active=lion:2,lion:4,moon:2,star:2`。
- 不变量：`active` 里不会出现同一套装的多个档位；件数越多的档位不会带来更低的攻击或防御；
  同一份装备任意顺序求值结果相同（`active` 输出按字典序）。
- 规模：一次在 5 万件装备、200 个套装的档位表上 O(n + m) 算完（扫一遍装备按套装计数，
  再扫一遍档位表），内存 O(套装数)；不许按件数展开，也不许对每件装备遍历整张档位表。

## 现在的行为

- `Apply` 用 `len(items)` 当件数，`pieces > bonus.Pieces` 当判定，`lion:2` 要凑到三件才生效。
- 每个达到 `pieces > Pieces` 的档位都会加一遍，同一套装的两档会叠在一起。

## 输出

```
attack=0 defense=0 active=none
attack=10 defense=0 active=lion:2
attack=40 defense=20 active=lion:6
```

## 目录

```
bonus.go                套装加成
cmd/set-bonus           命令行入口
bonus_test.go           go test 用例
```
