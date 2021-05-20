# Buildpacks GitHub Config

This repository contains config files common to buildpacks.

## How do I consume this common config

If you just wrote a new buildpack, run bootstrap.sh as follows:

```
./scripts/bootstrap.sh --target <path/to/your/buildpack>
```

This will copy the relevant config files to your buildpack. Git commit and Push.

Now, to wire up your buildpack repo to receive relevant updates as a pull requests:

* Configure secrets as required in all workflows

Submit your change to this repo as a PR. You should be all set when the PR is merged.
