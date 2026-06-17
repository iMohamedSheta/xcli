package xcli

import (
	"fmt"
	"sort"
)

func printBanner() {
	fmt.Println(Green + `
	██╗  ██╗ ██████╗██╗     ██╗
	╚██╗██╔╝██╔════╝██║     ██║
	 ╚███╔╝ ██║     ██║     ██║
	 ██╔██╗ ██║     ██║     ██║
	██╔╝ ██╗╚██████╗███████╗██║
	╚═╝  ╚═╝ ╚═════╝╚══════╝╚═╝
` + Yellow + "XCli - xApp Cli, Developer Tooling Made Simple" + Reset)
}

func (x *XCli) help() {
	fmt.Printf(Green + `
╔══════════════════════════════════════════╗
║              XCli Help                   ║
╚══════════════════════════════════════════╝
` + Reset)

	fmt.Println(Yellow + "Usage:" + Reset)
	fmt.Print("  XCli [command]\n\n")

	fmt.Println(Yellow + "Available Commands:" + Reset)

	// Sort group names
	groupNames := make([]string, 0, len(x.preRegisteredCmds))
	for group := range x.preRegisteredCmds {
		groupNames = append(groupNames, group)
	}
	sort.Strings(groupNames)

	for _, group := range groupNames {
		fmt.Printf("\n  "+Yellow+"%s:"+Reset+"\n", ucFirst(group))
		cmds := x.preRegisteredCmds[group]
		sort.Slice(cmds, func(i, j int) bool {
			return cmds[i].Use < cmds[j].Use
		})

		for _, cmd := range cmds {
			fmt.Printf("    %-24s %s\n", cmd.Use, cmd.Short)
		}
	}

	// External registered commands
	fmt.Println("\n" + Yellow + "External:" + Reset)
	for _, c := range x.externalCmds {
		fmt.Printf("    %-24s %s\n", c.Use, c.Short)
		if c.Long != "" && c.Long != c.Short {
			fmt.Printf("      %s\n", c.Long)
		}
	}

	fmt.Println("\nUse \"XCli [command]\" to run a specific command.")
}
