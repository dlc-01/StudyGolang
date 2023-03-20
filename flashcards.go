package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

type (
	flashcards struct {
		reader    *bufio.Reader
		TermDef map[string]string `json:"termToDef"`
		defTerm   map[string]string
		err       map[string]int
		hiCounter int
		log       []string
	}
)

func main() {
	c := initCardsStruct()
	expPath := c.handleFlags()
	c.menu(expPath)
}

func initCardsStruct() *flashcards {
	return &flashcards{
		reader:  bufio.NewReader(os.Stdin),
		TermDef: make(map[string]string),
		defTerm: make(map[string]string),
		err:     make(map[string]int),
	}
}

func (c *flashcards) handleFlags() string {
	var imp, exp string

	flag.StringVar(&imp, "import_from", "", "")
	flag.StringVar(&exp, "export_to", "", "")
	flag.Parse()

	if imp != "" {
		c.importCards(imp)
	}

	return exp
}

func (c *flashcards) menu(expPath string) {
	for {
		msg := "Input the action (add, remove, import, export, ask, exit, log, hardest card, reset stats):"
		c.writeLogEntry(msg, true)

		line, _ := c.reader.ReadString('\n')
		line = strings.ToLower(strings.TrimSpace(line))
		c.writeLogEntry(line, false)

		switch line {
		case "add":
			c.addCard()
		case "remove":
			c.removeCard()
		case "import":
			c.importCards("")
		case "export":
			c.exportCards("")
		case "ask":
			c.ask()
		case "log":
			c.saveLog()
		case "hardest card":
			c.printHardestCards()
		case "reset stats":
			c.resetStats()
		case "exit":
			msg = "Bye bye!"
			if expPath != "" {
				c.exportCards(expPath)
			}
			c.writeLogEntry(msg, true)

			return
		default:
		}
	}
}

func (c *flashcards) addCard() {
	msg := "Input the number of flashcards:"
	c.writeLogEntry(msg, true)

	term := c.getTerm()
	def := c.getDefinition()

	c.TermDef[term] = def
	c.defTerm[def] = term

	msg = fmt.Sprintf("The pair (\"%s\":\"%s\") has been added.", term, def)
	c.writeLogEntry(msg, true)
}

func (c *flashcards) getTerm() string {
	msg := "The card:"
	c.writeLogEntry(msg, true)

	for {
		line, _ := c.reader.ReadString('\n')
		line = strings.TrimSpace(line)
		c.writeLogEntry(line, false)

		if _, ok := c.TermDef[line]; !ok {
			return line
		} else {
			msg = fmt.Sprintf("The term \"%s\" already exists. Try again:", line)
			c.writeLogEntry(msg, true)
		}
	}
}

func (c *flashcards) getDefinition() string {
	msg := "The definition of the card:"
	c.writeLogEntry(msg, true)

	for {
		line, _ := c.reader.ReadString('\n')
		line = strings.TrimSpace(line)
		c.writeLogEntry(line, false)

		if _, ok := c.defTerm[line]; !ok {
			return line
		} else {
			msg = fmt.Sprintf("The definition \"%s\" already exists. Try again:", line)
			c.writeLogEntry(msg, true)
		}
	}
}

func (c *flashcards) removeCard() {
	msg := "Which card?"
	c.writeLogEntry(msg, true)

	line, _ := c.reader.ReadString('\n')
	line = strings.TrimSpace(line)
	c.writeLogEntry(line, false)

	def, ok := c.TermDef[line]
	if ok {
		delete(c.TermDef, line)
		delete(c.defTerm, def)

		msg = "The card has been removed."
		c.writeLogEntry(msg, true)
	} else {
		msg = fmt.Sprintf("Can't remove \"%s\": there is no such card.", line)
		c.writeLogEntry(msg, true)
	}

}

func (c *flashcards) importCards(path string) {
	var line, msg string

	if path != "" {
		line = path
	} else {
		msg = "File name:"
		c.writeLogEntry(msg, true)

		line, _ = c.reader.ReadString('\n')
		line = strings.TrimSpace(line)
		c.writeLogEntry(line, false)
	}

	file, err := os.Open(line)
	if err != nil {
		msg = "File not found."
		c.writeLogEntry(msg, true)

		return
	}
	defer file.Close()

	byteValue, _ := io.ReadAll(file)

	fileCards := make(map[string]string)

	err = json.Unmarshal(byteValue, &fileCards)
	if err != nil {
		msg = fmt.Sprintf("Error at Unmarshal: %s", err.Error())
		c.writeLogEntry(msg, true)

		return
	}

	for term, def := range fileCards {
		c.TermDef[term] = def
		c.defTerm[def] = term
	}

	msg = fmt.Sprintf("%d flashcards have been loaded.", len(fileCards))
	c.writeLogEntry(msg, true)
}

func (c *flashcards) exportCards(expPath string) {
	var line, msg string
	if expPath != "" {
		line = expPath
	} else {
		msg = "File name:"
		c.writeLogEntry(msg, true)

		line, _ = c.reader.ReadString('\n')
		line = strings.TrimSpace(line)
		c.writeLogEntry(line, false)
	}

	cardsJSON, err := json.MarshalIndent(c.TermDef, "", "  ")
	if err != nil {
		msg = fmt.Sprintf("Error at MarshalIndent: %s", err.Error())
		c.writeLogEntry(msg, true)

		return
	}

	err = os.WriteFile(line, cardsJSON, 0644)
	if err != nil {
		msg = fmt.Sprintf("Error at WriteFile: %s", err.Error())
		c.writeLogEntry(msg, true)

		return
	}

	msg = fmt.Sprintf("%d flashcards have been saved.", len(c.TermDef))
	c.writeLogEntry(msg, true)
}

func (c *flashcards) ask() {
	msg := "How many times to ask?"
	c.writeLogEntry(msg, true)

	line, _ := c.reader.ReadString('\n')
	line = strings.TrimSpace(line)
	c.writeLogEntry(line, false)

	num, _ := strconv.Atoi(line)

	for i := 0; i < num; i++ {
		term, def := pickRandomEntry(c.TermDef)

		userDef := c.getAnswer(term)
		if userDef == def {
			msg = "Correct!"
			c.writeLogEntry(msg, true)
		} else {
			c.err[term] += 1
			if c.err[term] > c.hiCounter {
				c.hiCounter = c.err[term]
			}

			cDef, ok := c.defTerm[userDef]
			if ok {
				msg = fmt.Sprintf("Wrong. The right answer is \"%s\", but your definition is correct for \"%s\".", def, cDef)
				c.writeLogEntry(msg, true)
			} else {
				msg = fmt.Sprintf("Wrong. The right answer is \"%s\".", def)
				c.writeLogEntry(msg, true)
			}
		}
	}
}

func pickRandomEntry(m map[string]string) (string, string) {
	i := rand.Intn(len(m))
	for term, def := range m {
		if i == 0 {
			return term, def
		}
		i--
	}
	return "", ""
}

func (c *flashcards) getAnswer(term string) string {
	msg := fmt.Sprintf("Print the definition of \"%s\":", term)
	c.writeLogEntry(msg, true)

	line, _ := c.reader.ReadString('\n')
	line = strings.TrimSpace(line)
	c.writeLogEntry(line, false)

	return line
}

func (c *flashcards) writeLogEntry(str string, doPrint bool) {
	if doPrint {
		fmt.Println(str)
	}

	c.log = append(c.log, str)
}

func (c *flashcards) saveLog() {
	msg := "File name:"
	c.writeLogEntry(msg, true)

	var log string
	for _, entry := range c.log {
		log += entry + "\n"
	}

	line, _ := c.reader.ReadString('\n')
	line = strings.TrimSpace(line)
	c.writeLogEntry(line, false)

	err := os.WriteFile(line, []byte(log), 0644)
	if err != nil {
		msg = fmt.Sprintf("Error at WriteFile: %s", err.Error())
		c.writeLogEntry(msg, true)

		return
	}

	msg = "The log has been saved."
	c.writeLogEntry(msg, true)
}

func (c *flashcards) getHardestCards() map[int][]string {
	cardsMap := make(map[int][]string)
	for term, counter := range c.err {
		cardsMap[counter] = append(cardsMap[counter], term)
	}

	return cardsMap
}

func (c *flashcards) printHardestCards() {
	if len(c.err) == 0 {
		msg := "There are no flashcards with err."
		c.writeLogEntry(msg, true)
	} else {
		cardsMap := c.getHardestCards()

		msg := fmt.Sprintf("The hardest card %s \"%s\". You have %d err answering it.",
			func() string {
				if len(cardsMap[c.hiCounter]) > 1 {
					return "are"
				} else {
					return "is"
				}
			}(),
			strings.Join(cardsMap[c.hiCounter], ", "), c.hiCounter)

		c.writeLogEntry(msg, true)
	}
}

func (c *flashcards) resetStats() {
	c.err = make(map[string]int)
	c.hiCounter = 0

	msg := "Card statistics have been reset."
	c.writeLogEntry(msg, true)
}
