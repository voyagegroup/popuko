# popuko

[![Build Status (master)](https://travis-ci.org/nekoya/popuko.svg?branch=master)](https://travis-ci.org/nekoya/popuko)

## What is this?

- This is an operation bot to do these things automatically.
  - GitHub
    - merge a pull request.
    - assign a pull request to a reviewer.
- Almost reimplementation of [homu](https://github.com/barosl/homu).


## Command

You can use these command as the comment for pull request.

- `@<botname> r+`
  - Merge this pull request by `<botname>` with labeling `S-awaiting-merge`.
  - `@<botname> r=<reviewer>` means the same thing.
- `@<reviewer> r?`
  - Assign the pull request to the reviewer with labeling `S-awaiting-review`.


## Setup Instructions


### Build

This tools does not work with `go get` currently.
So you need to do these things.

0. Clone this source file
1. [`gom install`](https://github.com/mattn/gom)
2. `make new_config`
3. Fill `config.go` with your preference.
  - For building a single binary which contains all configure at the compile time.
4. `make build`
5. You can get the exec binary as named `popuko` into the current directory.


### Setup

#### GitHub

1. Start the exec binary in your server.
2. Set `http://<your_server_with_port>/github` for the webhook to your repository with these events
  - `Issue comment`
3. Create these labels to make the status visible.
  - `S-awaiting-review`: for a pull request assigned to some reviewer.
  - `S-awaiting-merge`: for a pull request queued to this bot.
  - `S-needs-rebase`: for an unmergeable pull request.
4. Done!


## TODO

- Intelligent cooperation with TravisCI.
- Intelligent parse your command.
- Emojify bot's comment.
- [See more...](https://github.com/nekoya/popuko/issues)
