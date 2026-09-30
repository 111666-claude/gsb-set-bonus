// Package setbonus 是装备套装加成：按套装计数判档、跨套装累加，换装只影响相关套装。
package setbonus

import (
	"fmt"
	"sort"
	"strings"
)

// Bonus 是某个套装的一个档位。
type Bonus struct {
	Set     string
	Pieces  int
	Attack  int
	Defense int
}

// Item 是一件装备。
type Item struct {
	Slot string
	Set  string
}

// Catalog 是各套装的档位表。
type Catalog struct {
	Bonuses []Bonus
}

// DefaultCatalog 是内置的档位表。
func DefaultCatalog() *Catalog {
	return &Catalog{Bonuses: []Bonus{
		{Set: "lion", Pieces: 2, Attack: 10},
		{Set: "lion", Pieces: 4, Attack: 25, Defense: 10},
		{Set: "lion", Pieces: 6, Attack: 40, Defense: 20},
		{Set: "moon", Pieces: 2, Defense: 15},
		{Set: "star", Pieces: 2, Attack: 8, Defense: 8},
	}}
}

// Inventory 是身上装备。
//
// 槽位与件数都在 Equip/Unequip 时增量维护：slots 保证一个槽位只有一件，
// counts 按套装各算各的，activeTier/attack/defense 只随受影响的套装变动，
// Active 只读缓存，不重扫任何装备。
type Inventory struct {
	catalog *Catalog

	slots  map[string]Item // 当前每个槽位上穿的装备
	counts map[string]int  // 每个套装当前还在身上的件数

	tiers map[string][]Bonus // 各套装的档位，按 Pieces 升序

	activeTier map[string]int // 各套装当前生效档位在 tiers 中的下标
	active     []string       // 按字典序排列的生效档位标签
	attack     int
	defense    int

	recomputed int
}

// NewInventory 建一个空背包。
func NewInventory(catalog *Catalog) *Inventory {
	inv := &Inventory{
		catalog:    catalog,
		slots:      map[string]Item{},
		counts:     map[string]int{},
		tiers:      map[string][]Bonus{},
		activeTier: map[string]int{},
	}
	for _, bonus := range catalog.Bonuses {
		inv.tiers[bonus.Set] = append(inv.tiers[bonus.Set], bonus)
	}
	for set := range inv.tiers {
		tiers := inv.tiers[set]
		sort.Slice(tiers, func(i, j int) bool { return tiers[i].Pieces < tiers[j].Pieces })
		inv.tiers[set] = tiers
	}
	return inv
}

// Equip 穿上一件装备；同一槽位已有旧件时替换旧件，件数只统计还在身上的装备。
func (inv *Inventory) Equip(item Item) {
	if old, ok := inv.slots[item.Slot]; ok {
		if old.Set == item.Set {
			inv.slots[item.Slot] = item
			return
		}
		inv.removePiece(old.Set)
		inv.slots[item.Slot] = item
		inv.addPiece(item.Set)
		inv.refreshSet(old.Set)
		inv.refreshSet(item.Set)
		return
	}
	inv.slots[item.Slot] = item
	inv.addPiece(item.Set)
	inv.refreshSet(item.Set)
}

// Unequip 脱下某个槽位上的装备；槽位为空时无操作。
func (inv *Inventory) Unequip(slot string) {
	old, ok := inv.slots[slot]
	if !ok {
		return
	}
	delete(inv.slots, slot)
	inv.removePiece(old.Set)
	inv.refreshSet(old.Set)
}

func (inv *Inventory) addPiece(set string) { inv.counts[set]++ }

func (inv *Inventory) removePiece(set string) {
	inv.counts[set]--
	if inv.counts[set] == 0 {
		delete(inv.counts, set)
	}
}

// Pieces 是某个套装当前还在身上的件数。
func (inv *Inventory) Pieces(set string) int {
	return inv.counts[set]
}

// refreshSet 只重算单个套装的生效档位，并把增量同步到总和与 active 列表。
func (inv *Inventory) refreshSet(set string) {
	tiers, ok := inv.tiers[set]
	if !ok {
		return
	}
	newIdx := -1
	if count := inv.counts[set]; count > 0 {
		if n := sort.Search(len(tiers), func(i int) bool { return tiers[i].Pieces > count }); n > 0 {
			newIdx = n - 1
		}
	}
	oldIdx := -1
	if idx, ok := inv.activeTier[set]; ok {
		oldIdx = idx
	}
	if oldIdx == newIdx {
		return
	}

	if oldIdx >= 0 {
		old := tiers[oldIdx]
		inv.attack -= old.Attack
		inv.defense -= old.Defense
		inv.deactivate(set)
	}
	if newIdx >= 0 {
		new := tiers[newIdx]
		inv.attack += new.Attack
		inv.defense += new.Defense
		inv.activate(set, fmt.Sprintf("%s:%d", set, new.Pieces))
	}
	if newIdx >= 0 {
		inv.activeTier[set] = newIdx
	} else {
		delete(inv.activeTier, set)
	}
}

// deactivate 从字典序 active 列表中删掉一个套装的标签。
func (inv *Inventory) deactivate(set string) {
	prefix := set + ":"
	pos := sort.Search(len(inv.active), func(i int) bool { return inv.active[i] >= prefix })
	if pos == len(inv.active) || !strings.HasPrefix(inv.active[pos], prefix) {
		return
	}
	copy(inv.active[pos:], inv.active[pos+1:])
	inv.active[len(inv.active)-1] = ""
	inv.active = inv.active[:len(inv.active)-1]
}

// activate 把一个套装的档位标签按字典序插进 active 列表。
func (inv *Inventory) activate(set, label string) {
	pos := sort.SearchStrings(inv.active, label)
	inv.active = append(inv.active, "")
	copy(inv.active[pos+1:], inv.active[pos:])
	inv.active[pos] = label
}

// Active 返回当前生效的总攻击、总防御与按字典序排列的档位标签。
// 直接读增量维护的缓存，不重扫装备。
func (inv *Inventory) Active() (int, int, []string) {
	active := append([]string(nil), inv.active...)
	return inv.attack, inv.defense, active
}

// Recomputed 是 Active 累计重算过的件数（规模观测）。
func (inv *Inventory) Recomputed() int { return inv.recomputed }

// Format 把生效档位拼成一行，空的时候返回 none。
func Format(active []string) string {
	if len(active) == 0 {
		return "none"
	}
	sorted := append([]string(nil), active...)
	sort.Strings(sorted)
	return strings.Join(sorted, ",")
}
