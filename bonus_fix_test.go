package setbonus

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

func TestMixedSetsDoNotStack(t *testing.T) {
	inventory := NewInventory(DefaultCatalog())
	inventory.Equip(Item{Slot: "head", Set: "lion"})
	inventory.Equip(Item{Slot: "chest", Set: "moon"})
	inventory.Equip(Item{Slot: "legs", Set: "star"})
	inventory.Equip(Item{Slot: "feet", Set: "wind"})
	attack, defense, active := inventory.Active()
	if attack != 0 || defense != 0 || len(active) != 0 {
		t.Fatalf("四个不同套装的散件不该有加成：%d %d %v", attack, defense, active)
	}
}

func TestExactTierCountActivates(t *testing.T) {
	inventory := NewInventory(DefaultCatalog())
	inventory.Equip(Item{Slot: "head", Set: "lion"})
	inventory.Equip(Item{Slot: "chest", Set: "lion"})
	attack, defense, active := inventory.Active()
	if attack != 10 || defense != 0 || !reflect.DeepEqual(active, []string{"lion:2"}) {
		t.Fatalf("正好两件应该生效 lion:2：%d %d %v", attack, defense, active)
	}
}

func TestReslotReplacesOldItem(t *testing.T) {
	inventory := NewInventory(DefaultCatalog())
	inventory.Equip(Item{Slot: "head", Set: "lion"})
	inventory.Equip(Item{Slot: "chest", Set: "lion"})
	inventory.Equip(Item{Slot: "head", Set: "moon"})
	inventory.Equip(Item{Slot: "legs", Set: "lion"})
	if got := inventory.Pieces("lion"); got != 2 {
		t.Fatalf("换装后 lion 应该还是两件：%d", got)
	}
	if got := inventory.Pieces("moon"); got != 1 {
		t.Fatalf("换装后 moon 应该是一件：%d", got)
	}
	attack, defense, active := inventory.Active()
	if attack != 10 || defense != 0 || !reflect.DeepEqual(active, []string{"lion:2"}) {
		t.Fatalf("换装后应该只有 lion:2：%d %d %v", attack, defense, active)
	}
}

func TestOnlyHighestTierApplies(t *testing.T) {
	inventory := NewInventory(DefaultCatalog())
	slots := []string{"head", "chest", "legs", "feet", "hands", "back"}
	for _, slot := range slots {
		inventory.Equip(Item{Slot: slot, Set: "lion"})
	}
	attack, defense, active := inventory.Active()
	if attack != 40 || defense != 20 || !reflect.DeepEqual(active, []string{"lion:6"}) {
		t.Fatalf("六件 lion 只该取最高档 lion:6：%d %d %v", attack, defense, active)
	}
	seen := map[string]int{}
	for _, entry := range active {
		set := entry[:len(entry)-2]
		seen[set]++
		if seen[set] > 1 {
			t.Fatalf("同一套装不该出现多个档位：%v", active)
		}
	}
}

func TestMultipleSetsAccumulate(t *testing.T) {
	inventory := NewInventory(DefaultCatalog())
	inventory.Equip(Item{Slot: "head", Set: "lion"})
	inventory.Equip(Item{Slot: "chest", Set: "lion"})
	inventory.Equip(Item{Slot: "legs", Set: "moon"})
	inventory.Equip(Item{Slot: "feet", Set: "moon"})
	inventory.Equip(Item{Slot: "hands", Set: "star"})
	inventory.Equip(Item{Slot: "back", Set: "star"})
	attack, defense, active := inventory.Active()
	if attack != 18 || defense != 23 || !reflect.DeepEqual(active, []string{"lion:2", "moon:2", "star:2"}) {
		t.Fatalf("三个套装应该各自累加：%d %d %v", attack, defense, active)
	}
}

func TestActiveNeverRecomputes(t *testing.T) {
	inventory := NewInventory(DefaultCatalog())
	for index := 0; index < 2000; index++ {
		inventory.Equip(Item{Slot: fmt.Sprintf("slot-%d", index), Set: "lion"})
	}
	for query := 0; query < 5000; query++ {
		inventory.Active()
	}
	if got := inventory.Recomputed(); got != 0 {
		t.Fatalf("Active 不许重算件数：%d", got)
	}
}

// reference 是参考实现：按槽位去重、按套装计数、取最高档。
type reference struct {
	slots map[string]Item
	tiers map[string][]Bonus
}

func newReference(catalog *Catalog) *reference {
	ref := &reference{slots: map[string]Item{}, tiers: map[string][]Bonus{}}
	for _, bonus := range catalog.Bonuses {
		ref.tiers[bonus.Set] = append(ref.tiers[bonus.Set], bonus)
	}
	return ref
}

func (ref *reference) equip(item Item)     { ref.slots[item.Slot] = item }
func (ref *reference) unequip(slot string) { delete(ref.slots, slot) }

func (ref *reference) pieces(set string) int {
	count := 0
	for _, item := range ref.slots {
		if item.Set == set {
			count++
		}
	}
	return count
}

func (ref *reference) active() (int, int, []string) {
	counts := map[string]int{}
	for _, item := range ref.slots {
		counts[item.Set]++
	}
	attack, defense := 0, 0
	var active []string
	for set, bonuses := range ref.tiers {
		highest := 0
		for _, bonus := range bonuses {
			if counts[set] >= bonus.Pieces && bonus.Pieces >= highest {
				highest = bonus.Pieces
			}
		}
		for _, bonus := range bonuses {
			if bonus.Pieces == highest && highest > 0 {
				attack += bonus.Attack
				defense += bonus.Defense
				active = append(active, fmt.Sprintf("%s:%d", set, highest))
			}
		}
	}
	sort.Strings(active)
	return attack, defense, active
}

func randomCatalog(rng *rand.Rand) *Catalog {
	catalog := &Catalog{}
	for set := 0; set < 5; set++ {
		name := fmt.Sprintf("set-%d", set)
		tiers := 1 + rng.Intn(3)
		thresholds := rng.Perm(6)[:tiers]
		sort.Ints(thresholds)
		for _, threshold := range thresholds {
			catalog.Bonuses = append(catalog.Bonuses, Bonus{
				Set:     name,
				Pieces:  threshold + 1,
				Attack:  rng.Intn(20),
				Defense: rng.Intn(20),
			})
		}
	}
	return catalog
}

// sameActive 把 nil 与空切片视为一致地比较两个档位列表。
func sameActive(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for index := range a {
		if a[index] != b[index] {
			return false
		}
	}
	return true
}

func TestRandomSequencesMatchReference(t *testing.T) {
	rng := rand.New(rand.NewSource(20260930))
	sets := []string{"set-0", "set-1", "set-2", "set-3", "set-4", "stray"}
	for trial := 0; trial < 200; trial++ {
		catalog := randomCatalog(rng)
		inventory := NewInventory(catalog)
		ref := newReference(catalog)
		slots := make([]string, 0, 8)
		for slot := 0; slot < 8; slot++ {
			slots = append(slots, fmt.Sprintf("slot-%d", slot))
		}
		var worn []Item
		for step := 0; step < 40; step++ {
			slot := slots[rng.Intn(len(slots))]
			switch rng.Intn(3) {
			case 0:
				inventory.Unequip(slot)
				ref.unequip(slot)
			default:
				item := Item{Slot: slot, Set: sets[rng.Intn(len(sets))]}
				inventory.Equip(item)
				ref.equip(item)
				worn = append(worn, item)
			}
			attack, defense, active := inventory.Active()
			wantAttack, wantDefense, wantActive := ref.active()
			if attack != wantAttack || defense != wantDefense {
				t.Fatalf("第 %d 组第 %d 步加成不一致：got %d/%d want %d/%d",
					trial, step, attack, defense, wantAttack, wantDefense)
			}
			if !sameActive(active, wantActive) {
				t.Fatalf("第 %d 组第 %d 步 active 不一致：got %v want %v",
					trial, step, active, wantActive)
			}
			for _, set := range sets {
				if got, want := inventory.Pieces(set), ref.pieces(set); got != want {
					t.Fatalf("第 %d 组第 %d 步 %s 件数不一致：got %d want %d",
						trial, step, set, got, want)
				}
			}
		}
		// 同一份装备任意顺序求值结果相同：把最终穿戴集合打乱重放。
		final := map[string]Item{}
		for _, item := range worn {
			if _, ok := inventory.slots[item.Slot]; ok {
				final[item.Slot] = inventory.slots[item.Slot]
			}
		}
		replay := NewInventory(catalog)
		order := make([]Item, 0, len(final))
		for _, item := range final {
			order = append(order, item)
		}
		rng.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
		for _, item := range order {
			replay.Equip(item)
		}
		attack, defense, active := inventory.Active()
		reAttack, reDefense, reActive := replay.Active()
		if attack != reAttack || defense != reDefense || !sameActive(active, reActive) {
			t.Fatalf("第 %d 组乱序重放不一致：%d/%d/%v vs %d/%d/%v",
				trial, attack, defense, active, reAttack, reDefense, reActive)
		}
	}
}

func BenchmarkEquipAndQuery(b *testing.B) {
	catalog := &Catalog{}
	for set := 0; set < 200; set++ {
		for tier := 1; tier <= 6; tier++ {
			catalog.Bonuses = append(catalog.Bonuses, Bonus{
				Set: fmt.Sprintf("set-%d", set), Pieces: tier, Attack: tier, Defense: tier,
			})
		}
	}
	inventory := NewInventory(catalog)
	for index := 0; index < 50000; index++ {
		inventory.Equip(Item{Slot: fmt.Sprintf("slot-%d", index), Set: fmt.Sprintf("set-%d", index%200)})
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		slot := fmt.Sprintf("slot-%d", n%50000)
		inventory.Equip(Item{Slot: slot, Set: fmt.Sprintf("set-%d", n%200)})
		inventory.Active()
	}
}
