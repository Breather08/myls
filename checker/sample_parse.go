package checker

import (
	"fmt"
	"os"
	"strings"
)

/*Option map consist of options*/
// var Option = map[rune]bool{}
var Option = map[rune]bool{}

var Input Data

type Data struct {
	Path   []string
	Errors []string
}

/*Parser parse and checks CLI input*/
func Parser(input []string) {
	var c = &Input
	flag := ""
	comError := "Command not specified"
	lsError := "Only 'ls' command is acceptable"
	// dirError := "No such file or directory"
	invError := "Invalid option detected"
	// input := os.Args[1:]

	/*Checks if data input and main command is acceptabe*/
	switch {
	case len(input) < 1:
		Input.Errors = append(Input.Errors, comError)
		c.SortPath()
		// c.CallErrors()
		// c.CallPath()
		return
	case input[0] != "ls":
		Input.Errors = append(Input.Errors, lsError)
		c.SortPath()
		// c.CallErrors()
		// c.CallPath()
		return
	}
	for _, value := range input[1:] {
		_, err := os.Open(value)
		if err == nil {
			Input.Path = append(Input.Path, value)
			continue
		}
		for index, char := range value {
			switch {
			case value[0] == '-' && len(value) != 1:
				switch {
				case !LatChecker(char) && index != 0:
					Input.Errors = append(Input.Errors, invError)
					continue
				case LatChecker(char):
					Option[char] = true
				}
			default:
				flag += string(char)
			}
		}
		if flag == value {
			Input.Errors = append(Input.Errors, "cannot access '"+value+"': No such file or directory")
		}
		flag = ""
	}
	if len(Input.Path) < 1 && len(Input.Errors) < 1 {
		Input.Path = append(Input.Path, ".")
	}
	c.SortPath()
	// c.CallErrors()
	// c.CallPath()
}

/*debug section*/
func (c *Data) CallErrors() {
	for _, val := range c.Errors {
		fmt.Println(val)
	}
}
func (c *Data) CallPath() {
	for _, path := range c.Path {
		fmt.Println(path)
	}
}

/*end of dedug section*/

func (c *Data) SortPath() {
	var a, b string
	for i := 0; i < len(c.Path); i++ {
		for j := i; j < len(c.Path); j++ {
			a = strings.ToUpper(c.Path[i])
			b = strings.ToUpper(c.Path[j])
			switch {
			case Option['r']:
				if a <= b {
					c.Path[i], c.Path[j] = c.Path[j], c.Path[i]
				}
			case a > b:
				c.Path[i], c.Path[j] = c.Path[j], c.Path[i]

			}
		}
	}
}

/*LatChecker Just checks if char is Latin or not*/
func LatChecker(r rune) bool {
	if (r >= 65 && r <= 90) || (r >= 97 && r <= 122) {
		return true
	}
	return false
}
