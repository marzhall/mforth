package main

import (
    "bufio"
    "fmt"
    "os"
    "io"
    "strings"
    "strconv"
)

type ValueType int

const(
	Unset = iota
	Num
	BuiltinOp
	FlowControl
	FuncCall
	String
)

func (v ValueType) String() string {
    return [...]string{"Unset", "Num", "BuiltinOp", "FlowControl", "FuncCall", "String"}[v]
}

type StackEntry interface {
	Previous() StackEntry
	Pop() (StackEntry, StackEntry)
	Append(newEntry StackEntry)
	ValueType() ValueType
	Value() string
}


// StackStatement ------------------------------

type StackStatement struct {
	value string
	valueType ValueType
	previous StackEntry
}

func(s *StackStatement) Append(newEntry StackEntry) {
	if (newEntry == nil) {
		return
	}

	if (s.previous == nil) {
		s.previous = newEntry
	} else {
		s.previous.Append(newEntry)
	}
}

func(s *StackStatement) Previous() StackEntry {
	// fmt.Println("s previous is", s.previous.Value())
	return s.previous
}

func (s *StackStatement) Value() string {
	return s.value
}

func (s *StackStatement) ValueType() ValueType {
	return s.valueType
}

func (s *StackStatement) Pop() (StackEntry, StackEntry) {
	return s, s.Previous()
}

func (s *StackStatement) String() string {
	str := ""
	var temp StackEntry = s
	for temp != nil {
		str = temp.Value() + " " + str
		temp = temp.Previous()
	}

	return str
}

// StackFlowcontrol ----------------------------

type StackFlowControl struct {
	*StackStatement
	// The 'branch' is the stack of logic to be executed should the flowcontrol be true
	branch StackEntry
}

func(s *StackFlowControl) Previous() StackEntry {
	flowcontrol_value, rest := EvaluateStack(s.previous).Pop()
	if (flowcontrol_value.Value() == "true") {
		s.branch.Append(rest)
		return s.branch
	}

	return rest
}

func (s *StackFlowControl) Pop() (StackEntry, StackEntry) {
	return s, s.Previous()
}


func (s *StackFlowControl) String() string {
	str := ""

	str = s.Value() + " " + str
	var temp StackEntry
	temp = s
	for temp != nil {
		str = temp.Value() + " " + str
	}

	for temp != nil {
		str = temp.Value() + " " + str
		temp = temp.Previous()
	}

	return str
}


//---------------------------------------------

type parseContext struct {
	PreviousContext *parseContext
	PreviousStackEntry StackEntry
}

// Generate a syntax tree from the parsed tokens
func parse(tokens chan string, stack StackEntry, resultChan chan StackEntry) {
	for value := range tokens {
		value = strings.TrimSpace(value)
		switch value {
		case "if":
			childBranchResult := make(chan StackEntry, 1)
			parse(tokens, nil, childBranchResult)
			childBranch := <-childBranchResult
			newEntry := &StackFlowControl{&StackStatement{value, FlowControl, stack}, childBranch}
			stack = newEntry
			closingThen := &StackStatement{"then", BuiltinOp, stack}
			stack = closingThen
		case "then":
			break
		case "else":
			fallthrough
		case "swap":
			fallthrough
		case "drop":
			fallthrough
		case "dup":
			fallthrough
		case ".":
			fallthrough
		case "-":
			fallthrough
		case "/":
			fallthrough
		case "*":
			fallthrough
		case "+":
			newEntry := &StackStatement{value, BuiltinOp, stack}
			stack = newEntry
		default:
			//fmt.Println("tokenizing", value)
			if (stack != nil) {
				//fmt.Println("stack head is currently ", stack.Value())
			}
			newEntry := &StackStatement{value, Num, stack}
			stack = newEntry
			//fmt.Println("stack head is now ", stack.Value())
			if (stack.Previous() != nil) {
				//fmt.Println("stack previous is ", stack.Previous().Value())
			}
		}
	}

	resultChan <- stack
}

func tokenize(text string, stack StackEntry) StackEntry {
	if (text == "\n") {
		return stack
	}

	values := strings.Split(text, " ")
	tokChan := make(chan string, 10)
	go func(){
		for _, value := range(values) {
			tokChan <- value
		}

		close(tokChan)
	}()

	resultChan := make(chan StackEntry)
	defer close(resultChan)
	go parse(tokChan, stack, resultChan)
	result := <-resultChan
	return result
}

func EvaluateStack(s StackEntry) StackEntry {
	if s == nil {
		return nil
	}

	tempstack := s
	currentEntry, tempstack := tempstack.Pop()
	switch currentEntry.ValueType() {
	case FlowControl:
		switch currentEntry.Value() {
		case "if":
			return EvaluateStack(tempstack)
		default:
			fmt.Println("Interpreter internal error: %s has been identified as a flow control statement, but isn't in our hardcoded list of operators.", currentEntry.Value())
			return currentEntry
		}
	case BuiltinOp:
		switch currentEntry.Value() {
		case "then":
			return EvaluateStack(tempstack)
		case "swap":
			val1, tempstack := EvaluateStack(tempstack).Pop()
			val2, tempstack := EvaluateStack(tempstack).Pop()
			tempstack = &StackStatement{val1.Value(), val1.ValueType(), tempstack}
			tempstack = &StackStatement{val2.Value(), val2.ValueType(), tempstack}
			return tempstack
		case "drop":
			_, tempstack := EvaluateStack(tempstack).Pop()
			return tempstack
		case "dup":
			firstVar, tempstack := EvaluateStack(tempstack).Pop()
			tempstack = &StackStatement{firstVar.Value(), firstVar.ValueType(), tempstack}
			tempstack = &StackStatement{firstVar.Value(), firstVar.ValueType(), tempstack}
			return tempstack
		case "*":
			firstVar, tempstack := EvaluateStack(tempstack).Pop()
			firstValue := firstVar.Value()
			valOneFloat, _ := strconv.ParseFloat(firstValue, 64)
			secondVar, tempstack := EvaluateStack(tempstack).Pop()
			secondValue := secondVar.Value()
			valTwoFloat, _ := strconv.ParseFloat(secondValue, 64)
			tempstack = &StackStatement{strconv.FormatFloat(valOneFloat * valTwoFloat, 'f', -1, 64), Num, tempstack}
			return tempstack
		case "/":
			firstVar, tempstack := EvaluateStack(tempstack).Pop()
			firstValue := firstVar.Value()
			valOneFloat, _ := strconv.ParseFloat(firstValue, 64)
			secondVar, tempstack := EvaluateStack(tempstack).Pop()
			secondValue := secondVar.Value()
			valTwoFloat, _ := strconv.ParseFloat(secondValue, 64)
			tempstack = &StackStatement{strconv.FormatFloat(valOneFloat / valTwoFloat, 'f', -1, 64), Num, tempstack}
			return tempstack
		case "+":
			firstVar, tempstack := EvaluateStack(tempstack).Pop()
			firstValue := firstVar.Value()
			valOneFloat, _ := strconv.ParseFloat(firstValue, 64)
			secondVar, tempstack := EvaluateStack(tempstack).Pop()
			secondValue := secondVar.Value()
			valTwoFloat, _ := strconv.ParseFloat(secondValue, 64)
			tempstack = &StackStatement{strconv.FormatFloat(valOneFloat + valTwoFloat, 'f', -1, 64), Num, tempstack}
			return tempstack
		case "-":
			firstVar, tempstack := EvaluateStack(tempstack).Pop()
			firstValue := firstVar.Value()
			valOneFloat, _ := strconv.ParseFloat(firstValue, 64)
			secondVar, tempstack := EvaluateStack(tempstack).Pop()
			secondValue := secondVar.Value()
			valTwoFloat, _ := strconv.ParseFloat(secondValue, 64)
			tempstack = &StackStatement{strconv.FormatFloat(valOneFloat - valTwoFloat, 'f', -1, 64), Num, tempstack}
			return tempstack
		default:
			fmt.Println("Interpreter internal error: %s has been identified as an operator, but isn't in our hardcoded list of operators.", currentEntry.Value())
			return currentEntry
		}
	default:
		return currentEntry
	}

	return tempstack
}

func main () {
	reader := bufio.NewReader(os.Stdin)
	var stack StackEntry = nil
	for ;; {
		fmt.Print("CMD: ")
		text, err := reader.ReadString('\n')
		if (err == io.EOF) {
			os.Exit(0)
		}

		// fmt.Println("About to tokenize")
		stack = tokenize(text, stack)
		// fmt.Println("Done tokenizing. About to evaluateStack")
		stack = EvaluateStack(stack)
		// fmt.Println("Done eval the stack. About to print the stack")
		if (stack != nil) {
			fmt.Println(stack)
		} else {
			fmt.Println("")
		}
	}
}
