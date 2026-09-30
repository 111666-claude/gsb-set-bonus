package setbonus

import "testing"

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
