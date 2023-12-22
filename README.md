# 玲珑命令行工具

一个[玲珑](https://linglong.dev)的命令行辅助工具。

## 命令帮助

### 提取玲珑 layer 文件信息

```bash
Get info of linglong layer file

Usage:
  linglong-tools info [flags]

Examples:
linglong-tools info -f ./test.layer -p
linglong-tools info -f ./test.layer --format '{{ .Raw }}'
linglong-tools info -f ./test.layer --format '{{ .Info.Appid }}'
linglong-tools info -f ./test.layer --format '{{ index .Info.Arch 0 }}'

Flags:
  -f, --file string     layer file
      --format string   Format output using a custom template
  -h, --help            help for info
  -p, --prettier        output pretty JSON
```

### 推送玲珑 layer 文件

```bash
Push linglong layer file to remote repository

Usage:
  linglong-tools push [flags]

Examples:
# use environment variables: $LINGLONG_USERNAME and $LINGLONG_PASSOWRD (Recommend)
linglong-tools push -f ./test.layer -r https://repo.linglong.dev
# pass username and password
linglong-tools push -f ./test.layer -r https://user:pass@repo.linglong.dev
# pass repo name
linglong-tools push -f ./test.layer -r https://repo.linglong.dev -n develop-snipe

Flags:
  -c, --channel string   remote repo channel (default "linglong")
  -f, --file string      layer file
  -h, --help             help for push
  -n, --name string      remote repo name (default "repo")
  -p, --print            print all status
  -r, --repo string      remote repo url
```
