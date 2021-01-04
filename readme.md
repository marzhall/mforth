# Mforth

A small forth-like.

Operations and values are added to a stack. When evaluated, values pop the previous entries off of the stack to use as their arguments.

E.g., an add operations would look like:

    CMD: 2 2 +
    4

This places the values "2" and "2" on the stack, then a `+` operator. The `+` operator pops the `2`s off the stack, then adds them, placing the result back on the stack.

Stack values are kept after each evaluation. E.g.:

	CMD: 2
	2
	CMD: 2
	2 2
	CMD: +
	4

Is analagous to our earlier `2 2 +` example; we just added the `2`s and `+` one command at a time instead of all at once.

# Operators and control structures
The operations currently supported are:

- The arithmetic operators `+, -, /, *`. All numbers placed on the stack  are interpereted as doubles.
- The `drop` operator, which drops the value previous to it on the stack.

		CMD: test 4
		test 4
		CMD: drop
		test
- The `dup` operator, which copies the value before it and duplicates it on the stack

		CMD: 4 dup
		4 4
- the `swap` command, which swaps the two values on the stack before it. E.g.,

		CMD: 4 5
		4 5
		CMD: swap
		5 4

- `if` statements of the form `<condition> if <thing to do> then`. 'true' is the only accepted boolean value; all other values evaluate to false.

		CMD: true if "yay!" then
		"yay!"
		CMD: hjsdklfhjdk if "yay!" then
