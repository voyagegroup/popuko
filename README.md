# popuko

[![Build Status (master)](https://travis-ci.org/karen-irc/popuko.svg?branch=master)](https://travis-ci.org/karen-irc/popuko)

## What is this?

- This is an operation bot to do these things automatically.
    - GitHub
        - merge a pull request.
        - assign a pull request to a reviewer.
        - patrol a pull request which are newly unmergeable by others.
- Almost reimplementation of [homu](https://github.com/barosl/homu).


## Motivation

[homu](https://github.com/barosl/homu) is the super great operation bot for development on GitHub
and it supports a lot of valuable features: merge pull request into the latest upstream, try to testing on TravisCI,
and more. But its development is not in active now. And also Mozilla's servo team maintains
[their forked version of homu](https://github.com/servo/homu). But it is developed for their specific usecase.
Not for other projects.

This project intent to re-implement homu with minimum features to support our projects for work including non-public activities,
and to simplify to deploy this bot. This is why we have developed this project.


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
    - `Push`
3. Set your bot's account (or the team which it belonging to) as a collaborator for the repository (give __write__ priviledge.)
4. Create these labels to make the status visible.
    - `S-awaiting-review`: for a pull request assigned to some reviewer.
    - `S-awaiting-merge`: for a pull request queued to this bot.
    - `S-needs-rebase`: for an unmergeable pull request.
    - `S-fails-tests-with-upstream`: for a pull request which fails tests after try to merge into upstream.
5. Done!


## Why there is no released version?

- __This project always lives in canary__.
- We only support the latest revision.
- All the `HEAD` of `master` branch is equal to our released version.
- The base revision and build date are embedded to the exec binary. You can see them by checking stdout on start it.


## TODO

- Intelligent cooperation with TravisCI.
- [See more...](https://github.com/karen-irc/popuko/issues)
