package banner

import (
	"fmt"
)

// ANSI color codes
const (
	Reset   = "\033[0m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	Bold    = "\033[1m"
)

// PrintBanner prints the startup banner with colors
func PrintBanner(appName, env, port, localIP string) {
	bc := Bold + Cyan // border color

	fmt.Println()
	fmt.Println(bc + "+================================================================+" + Reset)
	fmt.Println(bc + "|" + Reset + "  " + Bold + Green + "[*] " + appName + " started successfully!" + Reset + "                   " + bc + "|" + Reset)
	fmt.Println(bc + "+================================================================+" + Reset)
	fmt.Println(bc + "|" + Reset + "  " + Yellow + "> Environment:" + Reset + "  " + env + "                                      " + bc + "|" + Reset)
	fmt.Println(bc + "+----------------------------------------------------------------+" + Reset)
	fmt.Println(bc + "|" + Reset + "  " + Green + "> Local:" + Reset + "        http://localhost:" + port + "                         " + bc + "|" + Reset)
	fmt.Println(bc + "|" + Reset + "  " + Green + "> Network:" + Reset + "      http://" + localIP + ":" + port + "                      " + bc + "|" + Reset)
	fmt.Println(bc + "+----------------------------------------------------------------+" + Reset)
	fmt.Println(bc + "|" + Reset + "  " + Magenta + "> API Base:" + Reset + "     http://localhost:" + port + "/api/v1                  " + bc + "|" + Reset)
	fmt.Println(bc + "|" + Reset + "  " + Magenta + "> API Docs:" + Reset + "     http://localhost:" + port + "/docs                     " + bc + "|" + Reset)
	fmt.Println(bc + "|" + Reset + "  " + Magenta + "> Swagger:" + Reset + "      http://localhost:" + port + "/swagger/index.html      " + bc + "|" + Reset)
	fmt.Println(bc + "|" + Reset + "  " + Magenta + "> OpenAPI:" + Reset + "      http://localhost:" + port + "/swagger/doc.json        " + bc + "|" + Reset)
	fmt.Println(bc + "+================================================================+" + Reset)
	fmt.Println()
}
