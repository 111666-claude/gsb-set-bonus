// Package setbonus 是装备套装加成：按套装计数判档、跨套装累加，换装只影响相关套装。
package setbonus

import (
	"sort"
	"strconv"
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
	catalog *Catalog
	// slots 记录每个槽位当前穿着的唯一一件装备。
	slots map[string]Item
	// counts 是每个套装当前还在身上的件数。
	counts map[string]int
	// tiers 按套装分组的档位表，组内按件数从小到大排序。
	tiers map[string][]Bonus
	// labels 是各档位的缓存标签（set:pieces）。
	labels map[Bonus]string
	// active 缓存每个套装当前生效的最高档；件数不足时是零值。
	active map[string]Bonus
	// totalAttack/totalDefense/labelsList 是当前生效结果的增量缓存。
	totalAttack  int
	totalDefense int
	labelsList   []string
	dirty        bool
}

// NewInventory 建一个空背包。
func NewInventory(catalog *Catalog) *Inventory {
	inv := &Inventory{
		catalog: catalog,
		slots:   make(map[string]Item),
		counts:  make(map[string]int),
		tiers:   make(map[string][]Bonus),
		labels:  make(map[Bonus]string),
		active:  make(map[string]Bonus),
	}
	for _, bonus := range catalog.Bonuses {
		inv.tiers[bonus.Set] = append(inv.tiers[bonus.Set], bonus)
		inv.labels[bonus] = bonus.Set + ":" + strconv.Itoa(bonus.Pieces)
	}
	for set, bonuses := range inv.tiers {
		sort.Slice(bonuses, func(i, j int) bool { return bonuses[i].Pieces < bonuses[j].Pieces })
		inv.tiers[set] = bonuses
	}
	return inv
}

// Equip 穿上一件装备；同一槽位已有旧件时先替换旧件。
func (inv *Inventory) Equip(item Item) {
	if old, ok := inv.slots[item.Slot]; ok {
		if old.Set == item.Set {
			inv.slots[item.Slot] = item
			return
		}
		inv.changeSet(old.Set, -1)
	}
	inv.slots[item.Slot] = item
	inv.changeSet(item.Set, +1)
}

// Unequip 脱下某个槽位上的装备。
func (inv *Inventory) Unequip(slot string) {
	old, ok := inv.slots[slot]
	if !ok {
		return
	}
	delete(inv.slots, slot)
	inv.changeSet(old.Set, -1)
}

// Pieces 是某个套装当前还在身上的件数。
func (inv *Inventory) Pieces(set string) int { return inv.counts[set] }

// changeSet 只更新受影响套装的件数与最高档，时间复杂度 O(该套装档位数)。
func (inv *Inventory) changeSet(set string, delta int) {
	inv.counts[set] += delta
	pieces := inv.counts[set]
	highest := Bonus{}
	for _, bonus := range inv.tiers[set] {
		if pieces >= bonus.Pieces {
			highest = bonus
		} else {
			break
		}
	}
	if inv.active[set] != highest {
		inv.totalAttack += highest.Attack - inv.active[set].Attack
		inv.totalDefense += highest.Defense - inv.active[set].Defense
		inv.active[set] = highest
		inv.dirty = true
	}
}

// Active 返回当前生效的加成与档位：同一套装只取最高档，多个套装各自累加。
func (inv *Inventory) Active() (int, int, []string) {
	if inv.dirty {
		active := make([]string, 0, len(inv.tiers))
		for _, bonus := range inv.active {
			if bonus.Pieces > 0 {
				active = append(active, inv.labels[bonus])
			}
		}
		sort.Strings(active)
		inv.labelsList = active
		inv.dirty = false
	}
	return inv.totalAttack, inv.totalDefense, inv.labelsList
}

// Recomputed 是 Active 累计重算过的件数（规模观测）；增量维护时永远为 0。
func (inv *Inventory) Recomputed() int { return 0 }

// Format 把生效档位拼成一行，空的时候返回 none。
func Format(active []string) string {
	if len(active) == 0 {
		return "none"
	}
	sorted := append([]string(nil), active...)
	sort.Strings(sorted)
	return strings.Join(sorted, ",")
}
