// Package setbonus 是装备套装加成：按同一套装的件数判档、跨套装累加。
// 缺陷：用身上总件数判档、阈值用严格大于，还把同一套装的多个档位叠起来加。
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

// Apply 按身上装备算出加成，返回总攻击、总防御和生效的档位。
// 缺陷：件数用的是身上总件数，判定用严格大于，每个达标档位都往上加。
func (c *Catalog) Apply(items []Item) (int, int, []string) {
	pieces := len(items)
	attack, defense := 0, 0
	active := make([]string, 0, len(c.Bonuses))
	for _, bonus := range c.Bonuses {
		if pieces > bonus.Pieces {
			attack += bonus.Attack
			defense += bonus.Defense
			active = append(active, fmt.Sprintf("%s:%d", bonus.Set, bonus.Pieces))
		}
	}
	return attack, defense, active
}

// Format 把生效档位拼成一行，空的时候返回 none。
func Format(active []string) string {
	if len(active) == 0 {
		return "none"
	}
	sorted := append([]string(nil), active...)
	sort.Strings(sorted)
	return strings.Join(sorted, ",")
}
