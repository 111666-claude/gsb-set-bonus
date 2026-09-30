package setbonus

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"testing"
)

func TestSingleItemGivesNothing(t *testing.T) {
	inventory := NewInventory(DefaultCatalog())
	inventory.Equip(Item{Slot: "head", Set: "lion"})
	attack, defense, active := inventory.Active()
	if attack != 0 || defense != 0 || len(active) != 0 {
		t.Fatalf("一两件散件不该有加成：%d %d %v", attack, defense, active)
	}
}

func TestEmptyInventory(t *testing.T) {
	attack, defense, active := NewInventory(DefaultCatalog()).Active()
	if attack != 0 || defense != 0 || len(active) != 0 {
		t.Fatalf("空装备不该有加成：%d %d %v", attack, defense, active)
	}
	if Format(active) != "none" {
		t.Fatalf("没有档位时应该打成 none：%s", Format(active))
	}
}

func TestPiecesCountsDistinctSets(t *testing.T) {
	inventory := NewInventory(DefaultCatalog())
	inventory.Equip(Item{Slot: "head", Set: "lion"})
	inventory.Equip(Item{Slot: "chest", Set: "moon"})
	inventory.Equip(Item{Slot: "legs", Set: "lion"})
	if got := inventory.Pieces("lion"); got != 2 {
		t.Fatalf("lion 应该有两件：%d", got)
	}
	if got := inventory.Pieces("moon"); got != 1 {
		t.Fatalf("moon 应该有一件：%d", got)
	}
}

func TestUnequipRemovesSlot(t *testing.T) {
	inventory := NewInventory(DefaultCatalog())
	inventory.Equip(Item{Slot: "head", Set: "lion"})
	inventory.Equip(Item{Slot: "chest", Set: "lion"})
	inventory.Unequip("head")
	if got := inventory.Pieces("lion"); got != 1 {
		t.Fatalf("脱下之后 lion 应该只剩一件：%d", got)
	}
}

func TestCatalogTiers(t *testing.T) {
	catalog := DefaultCatalog()
	if len(catalog.Bonuses) != 5 {
		t.Fatalf("档位表被改了：%d", len(catalog.Bonuses))
	}
	seen := map[string]int{}
	for _, bonus := range catalog.Bonuses {
		seen[bonus.Set]++
	}
	if seen["lion"] != 3 || seen["moon"] != 1 || seen["star"] != 1 {
		t.Fatalf("档位表的分套件数不对：%v", seen)
	}
}

func TestMixedSetPiecesDoNotStack(t *testing.T) {
	inventory := NewInventory(DefaultCatalog())
	inventory.Equip(Item{Slot: "head", Set: "lion"})
	inventory.Equip(Item{Slot: "chest", Set: "moon"})
	inventory.Equip(Item{Slot: "legs", Set: "star"})
	inventory.Equip(Item{Slot: "feet", Set: "wind"})
	attack, defense, active := inventory.Active()
	if attack != 0 || defense != 0 || len(active) != 0 {
		t.Fatalf("四个不同套装的散件不该有加成：%d %d %v", attack, defense, active)
	}
	if Format(active) != "none" {
		t.Fatalf("散件 active 应该是 none：%s", Format(active))
	}
}

func TestExactlyAtThreshold(t *testing.T) {
	inventory := NewInventory(DefaultCatalog())
	inventory.Equip(Item{Slot: "head", Set: "lion"})
	if attack, defense, active := inventory.Active(); attack != 0 || defense != 0 || len(active) != 0 {
		t.Fatalf("一件 lion 不该到 2 件档：%d %d %v", attack, defense, active)
	}
	inventory.Equip(Item{Slot: "chest", Set: "lion"})
	attack, defense, active := inventory.Active()
	if attack != 10 || defense != 0 || Format(active) != "lion:2" {
		t.Fatalf("刚好两件 lion 该生效 2 件档：%d %d %v", attack, defense, active)
	}
}

func TestReslotReplacesOldItem(t *testing.T) {
	inventory := NewInventory(DefaultCatalog())
	inventory.Equip(Item{Slot: "head", Set: "lion"})
	inventory.Equip(Item{Slot: "chest", Set: "lion"})
	inventory.Equip(Item{Slot: "head", Set: "moon"})
	inventory.Equip(Item{Slot: "legs", Set: "lion"})
	if got := inventory.Pieces("lion"); got != 2 {
		t.Fatalf("换装后 lion 只能算还在身上的两件：%d", got)
	}
	if got := inventory.Pieces("moon"); got != 1 {
		t.Fatalf("moon 应该有一件：%d", got)
	}
	attack, defense, active := inventory.Active()
	if attack != 10 || defense != 0 || Format(active) != "lion:2" {
		t.Fatalf("换装后只有 lion 2 件档：%d %d %v", attack, defense, active)
	}
}

func TestOnlyHighestTierCounts(t *testing.T) {
	inventory := NewInventory(DefaultCatalog())
	slots := []string{"head", "chest", "legs", "feet", "hands", "back"}
	for index, slot := range slots {
		inventory.Equip(Item{Slot: slot, Set: "lion"})
		if index == 1 {
			attack, defense, active := inventory.Active()
			if attack != 10 || defense != 0 || Format(active) != "lion:2" {
				t.Fatalf("两件时只该有 2 件档：%d %d %v", attack, defense, active)
			}
		}
		if index == 3 {
			attack, defense, active := inventory.Active()
			if attack != 25 || defense != 10 || Format(active) != "lion:4" {
				t.Fatalf("四件时只该有 4 件档，低档不叠加：%d %d %v", attack, defense, active)
			}
		}
	}
	attack, defense, active := inventory.Active()
	if attack != 40 || defense != 20 || Format(active) != "lion:6" {
		t.Fatalf("六件时只该有 6 件档，低档不叠加：%d %d %v", attack, defense, active)
	}
}

func TestMultipleSetsStackAndSorted(t *testing.T) {
	inventory := NewInventory(DefaultCatalog())
	inventory.Equip(Item{Slot: "a1", Set: "star"})
	inventory.Equip(Item{Slot: "a2", Set: "star"})
	inventory.Equip(Item{Slot: "b1", Set: "moon"})
	inventory.Equip(Item{Slot: "b2", Set: "moon"})
	inventory.Equip(Item{Slot: "c1", Set: "lion"})
	inventory.Equip(Item{Slot: "c2", Set: "lion"})
	attack, defense, active := inventory.Active()
	if attack != 18 || defense != 23 {
		t.Fatalf("三个套装同时达标各自累加，期望 18/23：%d/%d", attack, defense)
	}
	want := []string{"lion:2", "moon:2", "star:2"}
	if len(active) != 3 {
		t.Fatalf("该有三个生效档位：%v", active)
	}
	for i, label := range want {
		if active[i] != label {
			t.Fatalf("active 必须按字典序，期望 %v，得到 %v", want, active)
		}
	}
}

func TestRepeatedEvaluationIsStable(t *testing.T) {
	inventory := NewInventory(DefaultCatalog())
	inventory.Equip(Item{Slot: "head", Set: "lion"})
	inventory.Equip(Item{Slot: "chest", Set: "lion"})
	var first []string
	for i := 0; i < 5; i++ {
		attack, defense, active := inventory.Active()
		if attack != 10 || defense != 0 || Format(active) != "lion:2" {
			t.Fatalf("第 %d 次求值结果漂移：%d %d %v", i, attack, defense, active)
		}
		if first == nil {
			first = active
		} else if len(active) != len(first) {
			t.Fatalf("重复求值结果不一致：%v vs %v", first, active)
		}
	}
	if inventory.Recomputed() != 0 {
		t.Fatalf("Active 不许重扫装备：%d", inventory.Recomputed())
	}
}

func TestEquipOrderIndependence(t *testing.T) {
	items := []Item{
		{Slot: "head", Set: "lion"},
		{Slot: "chest", Set: "lion"},
		{Slot: "legs", Set: "moon"},
		{Slot: "feet", Set: "moon"},
		{Slot: "hands", Set: "star"},
	}
	want := build(NewInventory(DefaultCatalog()), items)
	for seed := int64(0); seed < 20; seed++ {
		shuffled := append([]Item(nil), items...)
		rng := rand.New(rand.NewSource(seed))
		rng.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
		inventory := NewInventory(DefaultCatalog())
		for _, item := range shuffled {
			inventory.Equip(item)
		}
		attack, defense, active := inventory.Active()
		if attack != want.attack || defense != want.defense || Format(active) != want.active {
			t.Fatalf("seed=%d 顺序不同结果不同：%d %d %v", seed, attack, defense, active)
		}
	}
}

type activeSnapshot struct {
	attack  int
	defense int
	active  string
}

func build(inventory *Inventory, items []Item) activeSnapshot {
	for _, item := range items {
		inventory.Equip(item)
	}
	attack, defense, active := inventory.Active()
	return activeSnapshot{attack, defense, Format(active)}
}

// referenceInventory 是 README 口径的直白参考实现：
// 按槽位去重 + 按套装计数 + 每套装取满足 pieces>=档位 的最高档。
type referenceInventory struct {
	slots map[string]Item
	tiers map[string][]Bonus
}

func newReference(catalog *Catalog) *referenceInventory {
	ref := &referenceInventory{slots: map[string]Item{}, tiers: map[string][]Bonus{}}
	for _, bonus := range catalog.Bonuses {
		ref.tiers[bonus.Set] = append(ref.tiers[bonus.Set], bonus)
	}
	for set := range ref.tiers {
		tiers := ref.tiers[set]
		sort.Slice(tiers, func(i, j int) bool { return tiers[i].Pieces > tiers[j].Pieces })
		ref.tiers[set] = tiers
	}
	return ref
}

func (ref *referenceInventory) equip(item Item)     { ref.slots[item.Slot] = item }
func (ref *referenceInventory) unequip(slot string) { delete(ref.slots, slot) }
func (ref *referenceInventory) pieces(set string) int {
	count := 0
	for _, item := range ref.slots {
		if item.Set == set {
			count++
		}
	}
	return count
}

func (ref *referenceInventory) active() (int, int, []string) {
	counts := map[string]int{}
	for _, item := range ref.slots {
		counts[item.Set]++
	}
	attack, defense := 0, 0
	var active []string
	for set, tiers := range ref.tiers {
		for _, tier := range tiers {
			if counts[set] >= tier.Pieces {
				attack += tier.Attack
				defense += tier.Defense
				active = append(active, fmt.Sprintf("%s:%d", set, tier.Pieces))
				break
			}
		}
	}
	sort.Strings(active)
	return attack, defense, active
}

func TestRandomDifferentialAgainstReference(t *testing.T) {
	const sets = 6
	catalog := &Catalog{}
	for i := 0; i < sets; i++ {
		set := fmt.Sprintf("set%02d", i)
		catalog.Bonuses = append(catalog.Bonuses,
			Bonus{Set: set, Pieces: 2, Attack: i + 1, Defense: i},
			Bonus{Set: set, Pieces: 4, Attack: (i + 1) * 3, Defense: i * 2},
			Bonus{Set: set, Pieces: 6, Attack: (i + 1) * 7, Defense: i * 5},
		)
	}
	rng := rand.New(rand.NewSource(20240930))
	for iteration := 0; iteration < 200; iteration++ {
		inventory := NewInventory(catalog)
		ref := newReference(catalog)
		slots := make([]string, 4+rng.Intn(8))
		for i := range slots {
			slots[i] = fmt.Sprintf("slot%02d", i)
		}
		allSets := make([]string, 0, sets+2)
		for i := 0; i < sets; i++ {
			allSets = append(allSets, fmt.Sprintf("set%02d", i))
		}
		allSets = append(allSets, "junk0", "junk1")
		for step := 0; step < 40; step++ {
			slot := slots[rng.Intn(len(slots))]
			switch rng.Intn(3) {
			case 0:
				set := allSets[rng.Intn(len(allSets))]
				inventory.Equip(Item{Slot: slot, Set: set})
				ref.equip(Item{Slot: slot, Set: set})
			case 1:
				inventory.Unequip(slot)
				ref.unequip(slot)
			default:
				for _, set := range allSets {
					if got, want := inventory.Pieces(set), ref.pieces(set); got != want {
						t.Fatalf("iter=%d step=%d set=%s 件数不一致：%d vs %d",
							iteration, step, set, got, want)
					}
				}
				a1, d1, l1 := inventory.Active()
				a2, d2, l2 := ref.active()
				if a1 != a2 || d1 != d2 || Format(l1) != Format(l2) {
					t.Fatalf("iter=%d step=%d 加成不一致：(%d,%d,%s) vs (%d,%d,%s)",
						iteration, step, a1, d1, Format(l1), a2, d2, Format(l2))
				}
				for _, set := range allSets {
					count := 0
					for _, label := range l1 {
						if strings.HasPrefix(label, set+":") {
							count++
						}
					}
					if count > 1 {
						t.Fatalf("iter=%d step=%d set=%s 出现多个档位：%v", iteration, step, set, l1)
					}
				}
				// 再来一次，重复求值必须一致。
				a3, d3, l3 := inventory.Active()
				if a3 != a1 || d3 != d1 || Format(l3) != Format(l1) {
					t.Fatalf("iter=%d step=%d 重复求值漂移", iteration, step)
				}
			}
		}
	}
}

func TestScaleNoFullRescan(t *testing.T) {
	const setCount, tiersPerSet, itemCount = 200, 5, 50000
	catalog := &Catalog{}
	for i := 0; i < setCount; i++ {
		set := fmt.Sprintf("set%03d", i)
		for tier := 1; tier <= tiersPerSet; tier++ {
			catalog.Bonuses = append(catalog.Bonuses, Bonus{
				Set: set, Pieces: tier * 2, Attack: tier, Defense: tier,
			})
		}
	}
	inventory := NewInventory(catalog)
	for i := 0; i < itemCount; i++ {
		inventory.Equip(Item{Slot: fmt.Sprintf("slot%05d", i), Set: fmt.Sprintf("set%03d", i%setCount)})
	}
	const ops = 20000
	rng := rand.New(rand.NewSource(7))
	for i := 0; i < ops; i++ {
		slot := fmt.Sprintf("slot%05d", rng.Intn(itemCount))
		set := fmt.Sprintf("set%03d", rng.Intn(setCount))
		inventory.Equip(Item{Slot: slot, Set: set})
		inventory.Active()
	}
	if inventory.Recomputed() != 0 {
		t.Fatalf("Active 累计重扫件数必须为 0：%d", inventory.Recomputed())
	}
	attack, defense, active := inventory.Active()
	if attack < 0 || defense < 0 || len(active) > setCount {
		t.Fatalf("规模场景结果越界：%d %d %d", attack, defense, len(active))
	}
}
