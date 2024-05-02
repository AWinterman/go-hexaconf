# hexconf

When you configure process, there are basically three (possibly four) options

1. Load environment variables
2. Read a file from disk
3. Command line arguments
4. Make a request across a network (lol kubernetes)

Of these, service authors do not much like 3, since in kubernetes objects it's
not often clear what command is being run in a container, and it simply adds an
extra potentially error prone step between the config map and the invocation
that starts the service.

While kubernetes does 4., almost no one else will without additional support.
Standing up an API server is much more complicated than setting a few values in
a configmap, for most devs.

This project encodes a convention:

1. Load configuration from one or more yaml files
2. Override configuration from env variables.

It is inspired by 12 factor apps, which store configuration in environment
variables, while nodding to the reality that environment variables are
unsuitable to complex configs that some apps may required.

## Usage

See ./example_test.go for usage.



