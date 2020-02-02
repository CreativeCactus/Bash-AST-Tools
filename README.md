# Bash AST Tools

For inspecting and improving bash and taskfile workflows.

## shcom

Shell comments and nesting inspector.

Will list named functions and any comments beginning with `##`.
Can be queried for nested functions.
Useful for generating docs and command completion.

`grep "case '" sh_comments/main.go -A1` for a list of flags.

### Examples

Using the [taskfile](./task) in this directory as an example:

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

- `./shcom -1 task "run" -R -p""` Deep query without showing the root nodes, without indentation
```
noDescription 
docker # Run inside docker
default # Run the binary
```

- FUTURE: `./shcom -0 task "run* "` Globbing function paths 

