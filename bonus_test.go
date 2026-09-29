package setbonus

import "testing"

func TestSingleItemGivesNothing(t *testing.T) {
	catalog := DefaultCatalog()
	attack, defense, active := catalog.Apply([]Item{{Slot: "head", Set: "lion"}})
	if attack != 0 || defense != 0 || len(active) != 0 {
		t.Fatalf("一两件散件不该有加成：%d %d %v", attack, defense, active)
	}
}

func TestEmptyInventory(t *testing.T) {
	catalog := DefaultCatalog()
	attack, defense, active := catalog.Apply(nil)
	if attack != 0 || defense != 0 || len(active) != 0 {
		t.Fatalf("空装备不该有加成：%d %d %v", attack, defense, active)
	}
	if Format(active) != "none" {
		t.Fatalf("没有档位时应该打成 none：%s", Format(active))
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
