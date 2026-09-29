// Command set-bonus 跑套装加成样例。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"example.com/setbonus"
)

func items(pairs ...string) []setbonus.Item {
	out := make([]setbonus.Item, 0, len(pairs)/2)
	for index := 0; index+1 < len(pairs); index += 2 {
		out = append(out, setbonus.Item{Slot: pairs[index], Set: pairs[index+1]})
	}
	return out
}

// Run 执行一次命令行调用，返回退出码。
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("set-bonus", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sample := flags.String("sample", "mixed", "mixed / pair / full")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	var equipped []setbonus.Item
	switch *sample {
	case "mixed":
		// 四个不同套装各一件。
		equipped = items("head", "lion", "chest", "moon", "legs", "star", "feet", "wind")
	case "pair":
		// 同一套装正好两件。
		equipped = items("head", "lion", "chest", "lion")
	case "full":
		// 同一套装六件。
		equipped = items("head", "lion", "chest", "lion", "legs", "lion",
			"feet", "lion", "hands", "lion", "back", "lion")
	default:
		fmt.Fprintln(stderr, "需要 --sample mixed|pair|full")
		return 2
	}
	catalog := setbonus.DefaultCatalog()
	attack, defense, active := catalog.Apply(equipped)
	fmt.Fprintf(stdout, "attack=%d defense=%d active=%s\n", attack, defense, setbonus.Format(active))
	return 0
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
