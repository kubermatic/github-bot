# github-bot

Based on [adtac/cherry-pick-bot](https://github.com/adtac/cherry-pick-bot)

A bot to improve the GH experience, currently it can:

* Cherry-Pick PRs

## Setup

* Create a Github user
* Subscribe the user to the repositories you want the bot to work on
* Configure the Github user to get notifications via Web instead of Mail (`Personal Settings` -> `Notifications`)
* Create an access token for the Github user via `Personal Settings` -> `Developer Settings` -> `Personal access tokens`
* Export the token: `export GITHUB_ACCESS_TOKEN=my-token`
* Run the bot locally via `./hack/run-bot.sh`
* Deploy the bot into Kubernetes by entering adding the required secrets into `manifests/kubernetes.yaml` and then
  just execute `kubectl apply -f manifests/kubernetes.yaml`

### Usage

* Comment on a pr with `/cherry-pick target-branch`, this will make the bot add a `cherry-pick/target-branch` label
* Merge the pr. This will make the bot create a cherry-pick PR and report the status of that as comment onto the
  original PR

Caveats:

* This currently only works with PRs from the same repo, not with forks

### License

```
Copyright 2017 Adhityaa Chandrasekar

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
```
