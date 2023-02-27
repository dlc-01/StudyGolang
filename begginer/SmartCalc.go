package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode"
)

type Stack[T any] []T

func (stack *Stack[T]) add(s T) {
	*stack = append(*stack, s)
}

func (stack *Stack[T]) pop() T {
	lastElementPosition := len(*stack) - 1
	lastElement := (*stack)[lastElementPosition]
	*stack = (*stack)[:lastElementPosition]
	return lastElement
}

func (stack *Stack[T]) last() *T {
	if len(*stack) == 0 {
		return nil
	}

	lastElementPosition := len(*stack) - 1
	return &(*stack)[lastElementPosition]
}

type Operator string

func (operator *Operator) isOperator() bool {
	if len(*operator) == 1 && strings.ContainsAny(string(*operator), "*/()^") {
		return true
	}

	var count int
	for _, r := range *operator {
		if r == '-' || r == '+' {
			count++
		}
	}

	return len(*operator) == count
}

func (operator *Operator) convertAdmissibleOperator() {
	if !operator.isOperator() || !strings.ContainsAny(string(*operator), "+-") {
		return
	}

	if len(*operator)%2 == 0 || strings.Contains(string(*operator), "+") {
		*operator = "+"
		return
	}

	*operator = "-"
}

func (operator *Operator) getPriority() int {
	switch *operator {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	case "^":
		return 3
	}
	return 0
}

type Variables map[string]int

func (variables *Variables) getVariable(name string) (int, bool) {
	for k, v := range *variables {
		if k == name {
			return v, true
		}
	}
	return 0, false
}

func (variables *Variables) setVariable(assignment []string) error {
	name := strings.TrimSpace(assignment[0])
	if !variables.isValidIdentifier(name) {
		return errors.New("Invalid identifier")
	}

	rawValue := strings.TrimSpace(assignment[1])
	var value int
	if !variables.isValidIdentifier(rawValue) && !isNumber(rawValue) {
		return errors.New("Invalid assignment")
	}

	if val, ok := variables.getVariable(rawValue); ok && !isNumber(rawValue) {
		value = val
	} else if isNumber(rawValue) {
		value = getNumber(rawValue)
	} else {
		return errors.New("Unknown variable")
	}

	(*variables)[name] = value
	return nil
}

func (variables *Variables) isValidIdentifier(identifier string) bool {
	if len(identifier) == 0 {
		return false
	}

	var count int
	for _, r := range identifier {
		if unicode.IsLetter(r) {
			count++
		}
	}
	return count == len(identifier)
}

type Postfix []any

func (postfix *Postfix) append(element any) {
	if e, ok := element.(Operator); ok && (e == "(" || e == ")") {
		return
	}

	*postfix = append(*postfix, element)
}

func (postfix *Postfix) convert(infix Infix) {
	var operators Stack[Operator]

	infixElements := strings.Fields(string(infix))

	for _, element := range infixElements {
		if operator := Operator(element); operator.isOperator() {
			if len(operators) == 0 || *operators.last() == "(" {
				operators.add(operator)
				continue
			}

			if operator == "(" {
				operators.add(operator)
				continue
			}

			if operator == ")" {
				for operators.last() != nil && *operators.last() != "(" {
					postfix.append(operators.pop())
				}
				operators.pop()
				continue
			}

			if operator.getPriority() > operators.last().getPriority() {
				operators.add(operator)
				continue
			}

			if operator.getPriority() <= operators.last().getPriority() {
				for operators.last() != nil && (operators.last().getPriority() >= operator.getPriority() || *operators.last() != "(") {
					postfix.append(operators.pop())
				}
				operators.add(operator)
				continue
			}
		}
		postfix.append(element)
	}

	for operators.last() != nil {
		postfix.append(operators.pop())
	}
}

func isNumber(s string) bool {
	if len(s) == 0 {
		return false
	}

	if strings.HasPrefix(s, "+") || strings.HasPrefix(s, "-") {
		s = s[1:]
	}

	var count int
	for _, r := range s {
		if unicode.IsNumber(r) {
			count++
		}
	}
	return len(s) == count && count != 0
}

func getNumber(s string) int {
	number, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal(err)
	}
	return number
}

type Infix string

func (infix *Infix) validBrackets() bool {
	var leftBrackets, rightBrackets int
	for _, r := range *infix {
		if r == '(' {
			leftBrackets++
			continue
		}

		if r == ')' {
			rightBrackets++
		}

		if rightBrackets > leftBrackets {
			return false
		}
	}

	return leftBrackets == rightBrackets
}

func (infix *Infix) validOperators() bool {
	var lastOperator rune
	for _, r := range *infix {
		if r == lastOperator && (lastOperator == '*' || lastOperator == '/' || lastOperator == '^') {
			return false
		}
		lastOperator = r
	}
	return true
}

func (infix *Infix) validate() error {
	if !infix.validBrackets() || !infix.validOperators() {
		return errors.New("Invalid Expression")
	}
	return nil
}

func (infix *Infix) isSpaceLessForm() bool {
	if isNumber(string(*infix)) {
		return false
	}

	return !strings.Contains(string(*infix), " ")
}

func (infix *Infix) addSpaces() {
	*infix = Infix(strings.TrimSpace(string(*infix)))

	if !infix.isSpaceLessForm() {
		*infix = Infix(strings.Replace(string(*infix), "(", "( ", -1))
		*infix = Infix(strings.Replace(string(*infix), ")", " )", -1))
		return
	}

	// I think the space less form guaranties no admissible operators like '---' or '++'.
	var infixWithSpaces strings.Builder
	var lastElement rune
	for i, r := range *infix {

		if (isNumber(string(r)) && (isNumber(string(lastElement)) || i == 0)) ||
			(strings.ContainsAny(string(lastElement), "+-/*^") && strings.ContainsAny(string(r), "+-/*^")) ||
			(unicode.IsLetter(r) && (unicode.IsLetter(lastElement) || i == 0)) {
			infixWithSpaces.WriteString(string(r))
			lastElement = r
			continue
		}

		infixWithSpaces.WriteString(" " + string(r))
		lastElement = r
	}
	*infix = Infix(infixWithSpaces.String())
}

func calculate(postfix *Postfix, variables Variables) (int, error) {
	var result Stack[int]

	for _, e := range *postfix {
		if operator, ok := e.(Operator); ok {
			operator.convertAdmissibleOperator()

			b := result.pop()
			a := result.pop()

			switch operator {
			case "+":
				result.add(a + b)
			case "-":
				result.add(a - b)
			case "*":
				result.add(a * b)
			case "/":
				result.add(a / b)
			case "^":
				result.add(int(math.Pow(float64(a), float64(b))))
			}
			continue
		}

		if isNumber(e.(string)) {
			result.add(getNumber(e.(string)))
			continue
		}

		if !variables.isValidIdentifier(e.(string)) {
			return 0, errors.New("Invalid identifier")
		} else if _, ok := variables.getVariable(e.(string)); !ok {
			return 0, errors.New("Unknown variable")
		}

		val, _ := variables.getVariable(e.(string))
		result.add(val)
	}

	return *result.last(), nil
}

func handleCommands(s string) error {
	s = strings.TrimPrefix(s, "/")

	switch s {
	case "help":
		fmt.Println("The program calculates the sum of numbers")
	case "exit":
		fmt.Println("Bye!")
		os.Exit(0)
	default:
		return errors.New("unknown command")
	}
	return nil
}

func main() {
	var variables = make(Variables)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var infix Infix
		var postfix Postfix

		input := scanner.Text()

		if len(input) == 0 {
			continue
		}

		if strings.HasPrefix(input, "/") {
			err := handleCommands(input)
			if err != nil {
				fmt.Println("Unknown command")
			}
			continue
		}

		if strings.Contains(input, "=") {
			err := variables.setVariable(strings.SplitN(input, "=", 2))
			if err != nil {
				fmt.Println(err)
				continue
			}
			continue
		}

		if val, ok := variables.getVariable(strings.TrimSpace(input)); ok {
			fmt.Println(val)
			continue
		}

		infix = Infix(input)
		infix.addSpaces()
		err := infix.validate()
		if err != nil {
			fmt.Println(err)
			continue
		}

		postfix.convert(infix)

		result, err := calculate(&postfix, variables)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(result)
	}
}
