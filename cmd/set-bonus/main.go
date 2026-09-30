// Command set-bonus 跑套装加成样例。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"example.com/setbonus"
)

func run(sample string, stdout, stderr io.Writer) int {
	inventory := setbonus.NewInventory(setbonus.DefaultCatalog())
	switch sample {
	case "mixed":
		// 四个不同套装各一件。
		equipped := [][2]string{{"head", "lion"}, {"chest", "moon"}, {"legs", "star"}, {"feet", "wind"}}
		for _, pair := range equipped {
			inventory.Equip(setbonus.Item{Slot: pair[0], Set: pair[1]})
		}
		attack, defense, active := inventory.Active()
		fmt.Fprintf(stdout, "attack=%d defense=%d active=%s\n", attack, defense, setbonus.Format(active))
	case "pair":
		// 同一套装正好两件。
		inventory.Equip(setbonus.Item{Slot: "head", Set: "lion"})
		inventory.Equip(setbonus.Item{Slot: "chest", Set: "lion"})
		attack, defense, active := inventory.Active()
		fmt.Fprintf(stdout, "attack=%d defense=%d active=%s\n", attack, defense, setbonus.Format(active))
	case "full":
		// 同一套装六件。
		slots := []string{"head", "chest", "legs", "feet", "hands", "back"}
		for _, slot := range slots {
			inventory.Equip(setbonus.Item{Slot: slot, Set: "lion"})
		}
		attack, defense, active := inventory.Active()
		fmt.Fprintf(stdout, "attack=%d defense=%d active=%s\n", attack, defense, setbonus.Format(active))
	case "reslot":
		// head 先穿 lion 再换成 moon，然后又补了一件 lion。
		inventory.Equip(setbonus.Item{Slot: "head", Set: "lion"})
		inventory.Equip(setbonus.Item{Slot: "chest", Set: "lion"})
		inventory.Equip(setbonus.Item{Slot: "head", Set: "moon"})
		inventory.Equip(setbonus.Item{Slot: "legs", Set: "lion"})
		attack, defense, active := inventory.Active()
		fmt.Fprintf(stdout, "lion_pieces=%d attack=%d defense=%d active=%s\n",
			inventory.Pieces("lion"), attack, defense, setbonus.Format(active))
	case "cost":
		// 两千件装备，连续查五千次加成。
		for index := 0; index < 2000; index++ {
			inventory.Equip(setbonus.Item{Slot: fmt.Sprintf("slot-%d", index), Set: "lion"})
		}
		for query := 0; query < 5000; query++ {
			inventory.Active()
		}
		fmt.Fprintf(stdout, "recomputed=%d\n", inventory.Recomputed())
	default:
		fmt.Fprintln(stderr, "需要 --sample mixed|pair|full|reslot|cost")
		return 2
	}
	return 0
}

// Run 执行一次命令行调用，返回退出码。
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("set-bonus", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sample := flags.String("sample", "mixed", "mixed / pair / full / reslot / cost")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	return run(*sample, stdout, stderr)
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
