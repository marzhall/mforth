# mforth: It's turing-complete!

A small forth-like.

Operations and values are added to a stack. When evaluated, values pop the previous entries off of the stack to use as their arguments.

E.g., adding two plus two operation looks like:

    mforth: 2 2 +
    > 4

This places the values "2" and "2" on the stack, then a `+` operator. The `+` operator pops the `2`s off the stack, then adds them, placing the result back on the stack.

Stack values are kept after each evaluation. E.g.:

	mforth: 2
	> 2
	mforth: 2
	> 2 2
	mforth: +
	> 4

Is analagous to our earlier `2 2 +` example; we just added the `2`s and `+` one command at a time instead of all at once.

## Operators and control structures
The operations currently supported are:

### Simple Statment Operators
- The arithmetic operators `+, -, /, *`. All numbers placed on the stack are interpereted as doubles.
- The `.` operator, which prints the entire stack before it.
- The `drop` operator, which drops the value previous to it on the stack.

		mforth: test 4
		> test 4
		mforth: drop
		> test
- The `dup` operator, which copies the value before it and duplicates it on the stack

		mforth: 4 dup
		> 4 4
- the `swap` command, which swaps the two values on the stack before it. E.g.,

		mforth: 4 5
		> 4 5
		mforth: swap
		> 5 4

### Conditional Operations
- `if` statements of the form `<condition> if <thing to do> then`. The boolean values are 'true' and 'false'.

		mforth: true if "yay!" then
		> "yay!"
		mforth: false if "yay!" then
		>
- The comparison operators `>, <,`  and `==`, which operate on numbers; `==` performs a simple string comparison.
- The `!` operator, which will change the value `true` into `false`, and any string != `true` to `true`.

		mforth: true !
		> false
		mforth: !
		> true
### Function Definitions

`dec <statements> <name> as` is the format for functions, e.g.:

Aliasing the plus operator:

	mforth: dec + plus as
	>
	mforth: 2 2 plus
	> 4


Writing a simple factorial:

	mforth: dec dup 1 == ! if dup 1 - fact * then fact as
	> 
	mforth: 4 fact
	> 24
	mforth: 3 fact
	> 24 6
	mforth: 5 fact
	> 24 6 120
