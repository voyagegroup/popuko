# popuko

[![CI Status](https://github.com/voyagegroup/popuko/workflows/CI/badge.svg)](https://github.com/voyagegroup/popuko/actions?query=workflow%3ACI)

## What is this?

- This is an operation bot to do these things automatically for your project on GitHub.
    - merge a pull request automatically.
    - assign a pull request to a reviewer.
    - patrol a pull request which are newly unmergeable by others.
- Almost reimplementation of [homu][homu].


## Motivation

A development on GitHub with many developers requires many operations for users.
You need looking for who can review your pull request for the repository,
assigning your pull request to some reviewers,
checking why your pull request is not able to merge into the upstream branch,
checking that it will not cause any failures before merging your pull request,
merging your pull request actually, and etc.
However, basically, we have to operate these actions by hand. It's stressful.

As a developer, we must automate them by creating some bots for GitHub.
We should achieve hassle-free development.

In the area of automating GitHub operation, [homu][homu] is the super great operation bot
and it supports a lot of valuable features: merge pull request into the latest upstream, try to testing on TravisCI, and more.
It realizes the principle: _The Not Rocket Science Rule Of Software Engineering:
automatically maintain a repository of code that always passes all the tests_ [by Graydon Hoare][graydon's-entry].
Homu is used in [Rust language][github-rust-repo] and [Servo][github-servo]. It works well in there.
But its development is not in active now. And also Mozilla's servo team maintains
[their forked version of homu][servo-homu]. But it is developed for their specific usecase.
Not for other projects.

And, to use without host homu by yourself, you can use `homu.io` or other similar service (e.g. [bors.tech][bors.tech]).
But it's shared by other third repositories. It would not suite to use it for your internal repository.

Some features (e.g. assigning reviewers to the pull request) are provided by [highfive][highfive], not by homu.
Thus you also have to setup it to use their features which are used in code review frequently.

And furthermore, homu's reviewer configuration need to configure the central configuration file.
But we'd like to place the configuration for each repositries as decentralization.
This decentraization is important if you manage many repositories and
each of them has contibutors and reviewers individually.

By these things, this project intent to re-implement homu and highfive with minimum features
which can support a review process, and the primary targets are an internal repository on GitHub for work
or a public repository which want to host some merge bots by themselves.
And also this aims to simpify deploying this bot. We challenge to make it easier than the original's one.

These are why we have developed this project.


## Features

These features are inspired by [homu][homu] and [highfive][highfive].

- __Change the labels, the assignees of the pull request by comments__
    - By a reviewer's comment, this bot changes the labels, the assignees of the pull request.
- __Patrol pull requests which cannot merge into it after the upstream has been updated__
    - This bot patrols automatically by hooking GitHub's push events.
    - Change the label for the unmergeable pull request and comment about it.
- __Try the pull request with the latest default branch, and merge into it automatically__
    - We call this feature as "Auto-Merging".
- __Specify a reviewer by a file committed to the repository__
    - This feature is not implemented by homu.
    - You can manage a reviewer by normal pull request process for open governance.

### Command

You can use these command as the comment for pull request.

#### `r? @<reviewer>`

- Assign the reviewer to the pull request with labeling `S-awaiting-review`.
- You can call `r? @<reviewer1> @<reviewer2>` to assign multiple reviewers.
- You can also call `@<reviewer> r?` (But this is deprecated syntax).
- All user can call this command.

#### `@<botname> r+` or `@<botname> r=<reviewer>`

- Mark this pull request as `S-awaiting-merge` by labeling.
- If you enable Auto-Merging, this bot queues the pull request into the approved queue.
- Require _reviewer_ privilege to call this command.

#### `@<botname> r-`

- Cancel the approved by `@<botname> r+`.
    - If Auto-Merging is enabled, this removes the pull request from the approved queue.
- This set back the label to `S-awaiting-review`
- Require _reviewer_ privilege to call this command.


### Auto-Merging

This bot provides a powerful feature we called as _Auto-Merging_.
Auto-Merging behaves like this:

1. Accept the pull request by the review's approved comment (e.g. `@<bot> r+`)
2. This bot queues its pull request into the approved queue.
3. If there is no active item, try to merge it into the latest upstream and run CI on the special branch used for auto testing.
4. If the result of step 3 is success, this bot merge its pull request into the upstream actually.
   Otherwise, this bot marks it as failed.
5. This bot redo step 3 until the approved queue will be empty.


### Reviewer

- A _reviewer_ is managed by `OWNERS.json` places to the root of your repository.
- You can provide _reviewer_ privilege for all users that can comment to the repository.
    - This is useful for an internal repository.


## Setup Instructions

### Build & Launch the Application

1. Build from the source.
    - Run these steps:
        1. `make build` or `make build_linux_x64`.
    - Run `make help` to see more details.
2. Create the config directory.
    - By default, this app uses `$XDG_CONFIG_HOME/popuko/` as the config dir.
      (If you don't set `$XDG_CONFIG_HOME` environment variable, this use `~/.config/popuko/`.)
    - You can configure the config directory by `--config-base-dir`
3. Set `config.toml` to the config directory.
    - Let's copy from [`./example.config.toml`](./example.config.toml)
4. Start the exec binary.
    - This app dumps all logs into stdout & stderr.
    - If you'd like to use TLS, then provide `--tls`, `--cert`, and `--key` options.

#### Set up for your repository in GitHub.

1. Set the account (or the team which it belonging to) which this app uses as a collaborator
   for your repository (requires __write__ priviledge).
2. Add `OWNERS.json` file to the root of your repository.
    - Please see [`OwnersFile`](./setting/ownersfile.go) about the detail.
    - The example is [here](./OWNERS.json).
3. Set `http://<your_server_with_port>/github` for the webhook to your repository with these events:
    - `Issue comment`
    - `Push`
    - `Status` (required to use Auto-Merging feature (non GitHub App CI services)).
    - `Check Suite` (required to use Auto-Merging feature (GitHub App CI Services)).
    - `Pull Request` (required to remove all status (`S-` prefixed) labels after a pull request is closed).
4. Create these labels to make the status visible.
    - `S-awaiting-review`
        - for a pull request assigned to some reviewer.
    - `S-awaiting-merge`
        - for a pull request queued to this bot.
    - `S-needs-rebase`
        - for an unmergeable pull request.
    - `S-fails-tests-with-upstream`
        - for a pull request which fails tests after try to merge into upstream (used by Auto-Merging feature).
6. Enable to start the build on creating the branch named `auto` for your CI service (e.g. TravisCI).
    - You can configure this branch's name by `OWNERS.json`.
7. Done!


## FAQ

### Why there is no released version?

- __This project always lives in canary__.
- We only support the latest revision.
- All of `master` branch is equal to our released version.
- The base revision and build date are embedded to the exec binary. You can see them by checking stdout on start it.


### Out of scope of this project

- Full-replace homu.
- This project does not have any plan to re-implement all features of homu.
- No plans to create any alternatives of `homu.io` or [bors.tech][bors.tech].


### The current limitations

- If your pull request which try to be merged into non-default branch, this bot does not detect the conflict
  even if the upstream has been changed.
    - [TODO: #197](https://github.com/voyagegroup/popuko/issues/197)


### Can I reuse this package as a library?

- Yes... But I don't recomment to do it.
- Sorry. We don't think to maintain this package as a library.
    - We don't care the breaking change for library APIs.
- This repository is developed for the application, not to reuse from others.


### Do you have any plan to support GitLab or GitHub Enterprise?

- GitLab: see [#152](https://github.com/voyagegroup/popuko/issues/152).
- GitHub Enterprise: [#173](https://github.com/voyagegroup/popuko/issues/173).


### Does this bot recommend [Trunk Based Development](https://trunkbaseddevelopment.com/) style?

Yes. This bot is designed heavily for the development style which has "trunk" branch in a repository.

Of course, you can create a some branch on your upstream repository
and you can open your pull request which would be merged into their non-trunk branch.
However, we don't support the feature to conflict detection for them at this time
([#197](https://github.com/voyagegroup/popuko/issues/197)), and its priority is very low for us.


### Why didn't you fork homu?

It was notion.


### Why didn't you fork [bors-ng][bors-ng]?

When we had started this project, we could not find it.
And then we thought it's more better as an internal toolchain to implement a merging operation bot which is fully customized for our purpose.


## License

[The MIT License](./LICENSE.MIT)


## How to Contribute

- [TODO: Write CONTRUBUTING.md](https://github.com/voyagegroup/popuko/issues/97)
- If you have a problem, please find [existing issues](https://github.com/voyagegroup/popuko/issues) at first.
    - If there is no one which is similar to yours, please [file it as a new issue](https://github.com/voyagegroup/popuko/issues/new).
- We welcome your pull request. But we may not be able to accept your pull request if it is against to our project direction. Sorry.


## TODO

- Intelligent cooperation with TravisCI.
- [See more...](https://github.com/voyagegroup/popuko/issues)



[homu]: https://github.com/barosl/homu
[servo-homu]: https://github.com/servo/homu
[highfive]: https://github.com/servo/highfive
[bors.tech]: https://bors.tech/
[github-rust-repo]: https://github.com/rust-lang/
[github-servo]: https://github.com/servo
[graydon's-entry]: http://graydon2.dreamwidth.org/1597.html
[bors-ng]: https://github.com/bors-ng/bors-ng
