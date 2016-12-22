goprompt is a program to display version control information on the command
line.

To install, run `go install` in the repo directory, then add the following to `$HOME/.bash_aliases`:

    PROMPT_COMMAND="PS1=\"\$(goprompt)$ \""
