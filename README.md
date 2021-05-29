# gitlab-tools

A collection of small tools for various GitLab automation.

## gitlab-runner-janitor

Removes rotten GitLab Runners registered in the provided group.

Example:

```shell
GITLAB_TOKEN=abcdef gitlab-runner-janitor -group-id 123456
```
