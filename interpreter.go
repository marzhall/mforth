package main

import (
    "bufio"
    "fmt"
    "os"
    "io"
    "strings"
    "strconv"
    //"runtime/debug"
    // "reflect"
)

type ValueType int

const(
	Unset = iota
	Num
	BuiltinOp
	FlowControl
	FuncCall
	String
	Bool
)

func (v ValueType) String() string {
    return [...]string{"Unset", "Num", "BuiltinOp", "FlowControl", "FuncCall", "String", "Bool"}[v]
}

type StackEntry interface {
	String() string
	Previous() StackEntry
	Peek() StackEntry
	Pop() (StackEntry, StackEntry)
	Append(newEntry StackEntry)
	ValueType() ValueType
	Value() string
	Copy() StackEntry
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

func(s *StackStatement) Peek() StackEntry {
	// fmt.Println("s previous is", s.previous.Value())
	return s.previous
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
	prevStr := ""
	if (s.previous != nil) {
		prevStr = s.previous.String() + " "
	}

	return prevStr + s.Value()
}

func (s *StackStatement) Copy() StackEntry{
	var childStack StackEntry = nil
	if s.previous != nil {
		childStack = s.previous.Copy()
	}

	return &StackStatement{s.value, s.ValueType(), childStack}
}

// IfStatement -------------------------------

type IfStatement struct {
	*StackStatement
	// The 'branch' is the stack of logic to be executed should the flowcontrol be true
	Branch StackEntry
}

func (s *IfStatement) Copy() StackEntry{
	branchStack := s.Branch.Copy()
	previousStack := s.previous.Copy()
	return &IfStatement{&StackStatement{s.value, s.valueType, previousStack}, branchStack}
}

func (s *IfStatement) String() string {
	branchStr := ""
	prevStr := ""
	if (s.Branch != nil) {
		branchStr = s.Branch.String()
	}

	if (s.previous != nil) {
		prevStr = s.previous.String() + " "
	}

	return prevStr + s.Value() + " " + branchStr
}

func (s *IfStatement) Pop() (StackEntry, StackEntry) {
	return s, s.Previous()
}

//---------------------------------------------

// DecStatement -------------------------------

type DecStatement struct {
	*StackStatement
	FuncBody StackEntry
}

func (s *DecStatement) Copy() StackEntry {
	funcBodyStack := s.FuncBody.Copy()
	previousStack := s.previous.Copy()
	return &DecStatement{&StackStatement{s.value, s.valueType, previousStack}, funcBodyStack}
}

func (s *DecStatement) String() string {
	branchStr := ""
	prevStr := ""
	if (s.FuncBody != nil) {
		branchStr = s.FuncBody.String()
	}

	if (s.previous != nil) {
		prevStr = s.previous.String() + " "
	}

	return prevStr + s.Value() + " " + branchStr
}

func (s *DecStatement) Pop() (StackEntry, StackEntry) {
	return s, s.Previous()
}


//---------------------------------------------

func goBoolToMforthBool(val bool) string {
	if val {
		return "true"
	}

	return "false"
}

// Generate a syntax tree from the parsed tokens
func parse(tokens chan string, stack StackEntry, resultChan chan StackEntry) {
	tempstack := stack

	for value := range tokens {
		value = strings.TrimSpace(value)
		switch value {
		case "if":
			childBranchResult := make(chan StackEntry, 1)
			parse(tokens, nil, childBranchResult)
			childBranch := <-childBranchResult
			newEntry := &IfStatement{&StackStatement{value, FlowControl, tempstack}, childBranch}
			tempstack = newEntry
			closingThen := &StackStatement{"then", BuiltinOp, tempstack}
			tempstack = closingThen
		case "then":
			resultChan <- tempstack
			return
		case "dec":
			decBodyResult := make(chan StackEntry, 1)
			parse(tokens, nil, decBodyResult)
			decBody := <-decBodyResult
			newEntry := &DecStatement{&StackStatement{value, FlowControl, tempstack}, decBody}
			tempstack = newEntry
			closingDec := &StackStatement{"as", BuiltinOp, tempstack}
			tempstack = closingDec
		case "as":
			resultChan <- tempstack
			return
		case "false":
			newEntry := &StackStatement{value, Bool, tempstack}
			tempstack = newEntry
		case "true":
			newEntry := &StackStatement{value, Bool, tempstack}
			tempstack = newEntry
		case "else":
			fallthrough
		case "swap":
			fallthrough
		case "drop":
			fallthrough
		case "dup":
			fallthrough
		case "<":
			fallthrough
		case ">":
			fallthrough
		case "!":
			fallthrough
		case ".":
			fallthrough
		case "==":
			fallthrough
		case "-":
			fallthrough
		case "/":
			fallthrough
		case "*":
			fallthrough
		case "+":
			newEntry := &StackStatement{value, BuiltinOp, tempstack}
			tempstack = newEntry
		default:
			var newEntry StackEntry = nil
			_, err := strconv.ParseFloat(value, 64)
			if (err != nil) {
				newEntry = &StackStatement{value, FuncCall, tempstack}
			} else {
				newEntry = &StackStatement{value, Num, tempstack}
			}

			tempstack = newEntry
		}
	}

	resultChan <- tempstack
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

type Namespace struct {
	previousNamespace *Namespace
	funcs map[string]StackEntry
}

func(c *Namespace) AddFunctionDefinition(name string, s StackEntry) {
	c.funcs[name] = s
}

func(n *Namespace) GetFunctionCopy(name string) StackEntry {
	funcDef, ok := n.funcs[name]
	if !ok {
		if n.previousNamespace != nil {
			return n.previousNamespace.GetFunctionCopy(name)
		}

		return nil
	}

	return funcDef.Copy()
}

func MakeChildNamespace(oldNamespace *Namespace) *Namespace {
	newFuncMap := make(map[string]StackEntry)
	return &Namespace{oldNamespace, newFuncMap}
}


func EvaluateStack(s StackEntry, namespace *Namespace) StackEntry {
	if s == nil {
		return nil
	}

	tempstack := s
	currentEntry, tempstack := tempstack.Pop()
	switch currentEntry.ValueType() {
	case FuncCall:
		funcName := currentEntry.Value()
		funcCopy := namespace.GetFunctionCopy(funcName)
		if funcCopy == nil {
			fmt.Println("Couldn't find a function named", funcName)
		} else {
			funcCopy.Append(tempstack)
			tempstack = funcCopy
		}

		return EvaluateStack(tempstack, namespace)
	case FlowControl:
		switch currentEntry.Value() {
		case "dec":
			decStatement, ok := currentEntry.(*DecStatement)
			if !ok {
				fmt.Println("ERROR: We got an 'as' statement that isn't a DecStatement type. What's up with that?")
				return nil
			}

			funcName, funcDef := decStatement.FuncBody.Pop()
			namespace.AddFunctionDefinition(funcName.Value(), funcDef)
			return EvaluateStack(tempstack, namespace)
		case "if":
			ifStatement, ok := currentEntry.(*IfStatement)
			if !ok {
				fmt.Println("We weren't able to cast to an IfStatement Pointer for some reason.")
			}

			flowcontrol_value, tempstack := EvaluateStack(tempstack, namespace).Pop()
			if (flowcontrol_value.Value() == "true") {
				ifStatementBranch := ifStatement.Branch.Copy()
				ifStatementBranch.Append(tempstack)
				tempstack = ifStatementBranch
			}

			return EvaluateStack(tempstack, MakeChildNamespace(namespace))
		default:
			fmt.Println("Interpreter internal error: %s has been identified as a flow control statement, but isn't in our hardcoded list of operators.", currentEntry.Value())
			return currentEntry
		}
	case BuiltinOp:
		switch currentEntry.Value() {
		case "as":
			return EvaluateStack(tempstack, namespace)
		case "then":
			return EvaluateStack(tempstack, namespace)
		case "swap":
			val1, tempstack := EvaluateStack(tempstack, namespace).Pop()
			val2, tempstack := EvaluateStack(tempstack, namespace).Pop()
			tempstack = &StackStatement{val1.Value(), val1.ValueType(), tempstack}
			tempstack = &StackStatement{val2.Value(), val2.ValueType(), tempstack}
			return tempstack
		case "drop":
			_, tempstack := EvaluateStack(tempstack, namespace).Pop()
			return tempstack
		case "dup":
			firstVar, tempstack := EvaluateStack(tempstack, namespace).Pop()
			tempstack = &StackStatement{firstVar.Value(), firstVar.ValueType(), tempstack}
			tempstack = &StackStatement{firstVar.Value(), firstVar.ValueType(), tempstack}
			return tempstack
		case "*":
			firstVar, tempstack := EvaluateStack(tempstack, namespace).Pop()
			firstValue := firstVar.Value()
			valOneFloat, _ := strconv.ParseFloat(firstValue, 64)
			secondVar, tempstack := EvaluateStack(tempstack, namespace).Pop()
			secondValue := secondVar.Value()
			valTwoFloat, _ := strconv.ParseFloat(secondValue, 64)
			tempstack = &StackStatement{strconv.FormatFloat(valOneFloat * valTwoFloat, 'f', -1, 64), Num, tempstack}
			return tempstack
		case "/":
			firstVar, tempstack := EvaluateStack(tempstack, namespace).Pop()
			firstValue := firstVar.Value()
			valOneFloat, _ := strconv.ParseFloat(firstValue, 64)
			secondVar, tempstack := EvaluateStack(tempstack, namespace).Pop()
			secondValue := secondVar.Value()
			valTwoFloat, _ := strconv.ParseFloat(secondValue, 64)
			tempstack = &StackStatement{strconv.FormatFloat(valTwoFloat / valOneFloat, 'f', -1, 64), Num, tempstack}
			return tempstack
		case "==":
			firstVar, tempstack := EvaluateStack(tempstack, namespace).Pop()
			firstValue := firstVar.Value()
			secondVar, tempstack := EvaluateStack(tempstack, namespace).Pop()
			secondValue := secondVar.Value()
			tempstack = &StackStatement{goBoolToMforthBool(firstValue == secondValue), Num, tempstack}
			return tempstack
		case ".":
			fmt.Println("The stack before evaluating the . operator:\n", currentEntry)
			evalResults := EvaluateStack(tempstack, namespace)
			fmt.Println("The stack after evaluating the . operator:\n", evalResults)
			return evalResults
		case ">":
			firstVar, tempstack := EvaluateStack(tempstack, namespace).Pop()
			firstValue := firstVar.Value()
			valOneFloat, _ := strconv.ParseFloat(firstValue, 64)
			secondVar, tempstack := EvaluateStack(tempstack, namespace).Pop()
			secondValue := secondVar.Value()
			valTwoFloat, _ := strconv.ParseFloat(secondValue, 64)
			tempstack = &StackStatement{goBoolToMforthBool(valOneFloat > valTwoFloat), Num, tempstack}
			return tempstack
		case "!":
			firstVar, tempstack := EvaluateStack(tempstack, namespace).Pop()
			firstValue := firstVar.Value()
			result := "true"
			if (firstValue == "true") {
				result = "false"
			}

			tempstack = &StackStatement{result, Bool, tempstack}
			return tempstack
		case "<":
			firstVar, tempstack := EvaluateStack(tempstack, namespace).Pop()
			firstValue := firstVar.Value()
			valOneFloat, _ := strconv.ParseFloat(firstValue, 64)
			secondVar, tempstack := EvaluateStack(tempstack, namespace).Pop()
			secondValue := secondVar.Value()
			valTwoFloat, _ := strconv.ParseFloat(secondValue, 64)
			tempstack = &StackStatement{goBoolToMforthBool(valOneFloat < valTwoFloat), Num, tempstack}
			return tempstack
		case "+":
			firstVar, tempstack := EvaluateStack(tempstack, namespace).Pop()
			firstValue := firstVar.Value()
			valOneFloat, _ := strconv.ParseFloat(firstValue, 64)
			secondVar, tempstack := EvaluateStack(tempstack, namespace).Pop()
			secondValue := secondVar.Value()
			valTwoFloat, _ := strconv.ParseFloat(secondValue, 64)
			tempstack = &StackStatement{strconv.FormatFloat(valOneFloat + valTwoFloat, 'f', -1, 64), Num, tempstack}
			return tempstack
		case "-":
			firstVar, tempstack := EvaluateStack(tempstack, namespace).Pop()
			firstValue := firstVar.Value()
			valOneFloat, _ := strconv.ParseFloat(firstValue, 64)
			secondVar, tempstack := EvaluateStack(tempstack, namespace).Pop()
			secondValue := secondVar.Value()
			valTwoFloat, _ := strconv.ParseFloat(secondValue, 64)
			tempstack = &StackStatement{strconv.FormatFloat(valTwoFloat - valOneFloat, 'f', -1, 64), Num, tempstack}
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
	toplevel_namespace := MakeChildNamespace(nil)
	for ;; {
		fmt.Print("mforth: ")
		text, err := reader.ReadString('\n')
		if (err == io.EOF) {
			os.Exit(0)
		}

		// fmt.Println("About to tokenize")
		stack = tokenize(text, stack)
		// fmt.Println("Done tokenizing. About to evaluateStack")
		stack = EvaluateStack(stack, toplevel_namespace)
		// fmt.Println("Done eval the stack. About to print the stack")
		print("> ")
		if (stack != nil) {
			fmt.Println(stack)
		} else {
			fmt.Println("")
		}
	}
}
