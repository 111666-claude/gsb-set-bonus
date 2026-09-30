// Package setbonus 是装备套装加成：按套装计数判档、跨套装累加，换装只该影响相关套装。
// 缺陷：用身上总件数判档、阈值用严格大于、同一套装多档一起累加、换装不替换旧件、每次查询全量重算。
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
type Inventory struct {
	catalog    *Catalog
	items      []Item
	recomputed int
}

// NewInventory 建一个空背包。
func NewInventory(catalog *Catalog) *Inventory {
	return &Inventory{catalog: catalog}
}

// Equip 穿上一件装备。
// 缺陷：同一槽位重复换装只是追加，旧件继续留在列表里。
func (inv *Inventory) Equip(item Item) {
	inv.items = append(inv.items, item)
}

// Unequip 脱下某个槽位上的装备。
func (inv *Inventory) Unequip(slot string) {
	kept := inv.items[:0]
	for _, item := range inv.items {
		if item.Slot != slot {
			kept = append(kept, item)
		}
	}
	inv.items = kept
}

// Pieces 是某个套装当前的件数。
// 缺陷：直接数列表条目，换装留下的旧件也算。
func (inv *Inventory) Pieces(set string) int {
	count := 0
	for _, item := range inv.items {
		if item.Set == set {
			count++
		}
	}
	return count
}

// Active 返回当前生效的加成与档位。
// 缺陷：每次调用都重扫全部装备，用总件数判档、严格大于判阈值、同套装多档一起累加。
func (inv *Inventory) Active() (int, int, []string) {
	pieces := len(inv.items)
	inv.recomputed += pieces
	attack, defense := 0, 0
	active := make([]string, 0, len(inv.catalog.Bonuses))
	for _, bonus := range inv.catalog.Bonuses {
		if pieces > bonus.Pieces {
			attack += bonus.Attack
			defense += bonus.Defense
			active = append(active, fmt.Sprintf("%s:%d", bonus.Set, bonus.Pieces))
		}
	}
	sort.Strings(active)
	return attack, defense, active
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
