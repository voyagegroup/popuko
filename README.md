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
- `@<botname> r-`
    - Cancel the approved by `@<botname> r+`.
    - This set back the label to `S-awaiting-review`
- `@<reviewer> r?`
    - Assign the pull request to the reviewer with labeling `S-awaiting-review`.


## Setup Instructions

### Build & Launch the Application

1. Build from source file
    - You can do `go get`.
2. Create the config directory.
    - By default, this app uses `$XDG_CONFIG_HOME/popuko/` as the config dir.
      (If you don't set `$XDG_CONFIG_HOME` environment variable, this use `~/.config/popuko/`.)
    - You can configure the config directory by `-config-base-dir`
3. Set `config.toml` to the config directory.
    - Let's copy from [`./example.config.toml`](./example.config.toml)
4. Start the exec binary.
    - This app dumps all logs into stdout & stderr.

#### Set up for your repository in GitHub.

1. Set the account (or the team which it belonging to) which this app uses as a collaborator
   for your repository (requires __write__ priviledge).
2. Add `OWNERS.json` file to the root of your repository.
    - Please see [`OwnersFile`](./setting/ownersfile.go) about the detail.
    - The example is [here](./OWNERS.json).
3. Set `http://<your_server_with_port>/github` for the webhook to your repository with these events:
    - `Issue comment`
    - `Push`
    - `Status` (required to use Auto-Merge feature).
4. Create these labels to make the status visible.
    - `S-awaiting-review`
        - for a pull request assigned to some reviewer.
    - `S-awaiting-merge`
        - for a pull request queued to this bot.
    - `S-needs-rebase`
        - for an unmergeable pull request.
    - `S-fails-tests-with-upstream`
        - for a pull request which fails tests after try to merge into upstream (used by Auto-Merge feature).
6. Enable to start the build on creating the branch named `auto` for your CI service (e.g. TravisCI).
7. Done!


## Why there is no released version?

- __This project always lives in canary__.
- We only support the latest revision.
- All of `master` branch is equal to our released version.
- The base revision and build date are embedded to the exec binary. You can see them by checking stdout on start it.


## TODO

- Intelligent cooperation with TravisCI.
- [See more...](https://github.com/karen-irc/popuko/issues)
