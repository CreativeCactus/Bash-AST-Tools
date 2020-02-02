# Bash AST Tools

For inspecting and improving bash and taskfile workflows.

## shcom

Shell comments and nesting inspector.

Will list named functions and any comments beginning with `##`.
Can be queried for nested functions.
Useful for generating docs and command completion.

Note that parameters and flags can be passed in any order: `./shcom [flags...] file [flags...] [query...] [flags...] [query...]`
The file to parse is the first parameter without a `-` prefix.
A query is any subsequent parameter without a `-` prefix.
The query is collected into a string and split on spaces.

`grep "case '" sh_comments/main.go -A1` for a list of flags.

### Examples

Using the [taskfile](./task) in this directory as an example:

- `./shcom -0 task "run" -p'' -c` and `./shcom -0 task "run " -p'' -c`, the common use cases:
With `0` depth, parse file `task` with query `run` or `run ` (respectively), no indentation `p`refix, `c`ompact output (no descriptions):
```
runtime 
run
```
and
```
noDescription 
docker 
default
```

- `./shcom task "run"` List functions matching a query
```
  runtime # Run with timestamp
  run # Run
```

- `./shcom task "run "` List functions in a given context
```
  noDescription 
  docker # Run inside docker
  default # Run the binary
```

- `./shcom task -3 -x -p"____"` Deeply list functions matching blank query, with extended comments, with a custom indent prefix
```
task # Top level comment
     # Second level comment
____build # Build the program
____runtime # Run with timestamp
____run # Run
        # Defaults to running the binary
________noDescription 
                      # Does nothing
________docker # Run inside docker
________default # Run the binary
```

- `./shcom -1 task "run" -R -c -p""` Deep query without showing the root nodes or descriptions, without indentation
```
noDescription 
docker # Run inside docker
default # Run the binary
```

- FUTURE: `./shcom -0 task "run* "` Globbing function paths 

