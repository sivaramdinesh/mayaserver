 package cmd
 
 import (
 	"fmt"

 	"github.com/mitchellh/cli"
	
 )
 
 type StatusCommand struct {
	IP     string
	Status string	Port   string
	Ui     cli.Ui
	Ui cli.Ui
	// To command this CLI's display

 }
 
 func (c *StatusCommand) Help() string {
	return ""
	helpText := `
Usage: m-apiserver status

	Queries and Prints the current status ,
	port details and -bind ip address of  m-apiserver
`
return (helpText)
 }
 
 func (c *StatusCommand) Run(_ []string) int {
	var statusString bytes.Buffer
	// StatusCommand is a Command implementation prints the Status.

	// display the text in allignment
 
 	fmt.Printf("%-15s%-15s%-15s%-15s", "Service", "IP", "Port", "Status")
 	fmt.Println()
	fmt.Printf("%-15s%-15s%-15s%-1s", "m-apiserver", "17.28.12.04", "5656", "running")
	
 	return 0
 }
 
 func (c *StatusCommand) Synopsis() string
	return "Prints the m-apiserver Status"
 }
