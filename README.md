# cc: a small C compiler written in Go

It's just a toy, and it will still be a toy for the foreseeable future. 
But the plan is to achieve most C11 features, so it can be used to compile real-world programs, such as SQLite, Git and so on.
After achieving this goal, I also want to try some optimization techniques.

## Intro

cc compiles C programs into GNU assembly, currently only tested under macOS 10.15.3 and clang 11.0.

The complication consists of the following stages:
- Scanner
- Recursive descendent parser
- Codegen

## Why

There are already many toy C compilers, why another one? 

The main purpose is to learn. C is a compiled language with simple and compact syntax, which is very suitable for learning compiler techniques.

## Status

The project is still in its early stages, you can check the [tests](tests) folder for supported features.


## Contributing

The project currently does not accept pull requests.
If you find bugs, please file an issue.

## References

I learned a lot from the following resources:

- [Crafting Interpreters](https://craftinginterpreters.com/)
- [chibicc](https://github.com/rui314/chibicc)

## License

The source code is in the public-domain, you can do what the fuck you want without any limitations.