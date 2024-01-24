# 玲珑命令行工具

一个[玲珑](https://linglong.dev)的命令行辅助工具。

## 命令帮助

注意事项

1. 以下所有命令在出错时进程会退出并打印错误信息，可通过进程的状态码是否为 0 来判断是否出错
2. 以下所有命令的参数可使用简化名或正式名，简化名以`-`开头，正式名以`--`开头，两者作用相同
3. 以下所有命令都可使用 -h 或 --help 打印命令的帮助

### 提取玲珑 layer 文件信息

linglong-tools info 命令用于获取 linglong Layer 文件的信息。

示例

```bash
# 输出 JSON 格式的详细信息
linglong-tools info -f ./test.layer -p

# 使用自定义模板输出详细信息中的 Appid
linglong-tools info -f ./test.layer --format '{{ .Info.Appid }}'

# 使用自定义模板输出详细信息中的第一个 Arch
linglong-tools info -f ./test.layer --format '{{ index .Info.Arch 0 }}'
```

参数

    -f, --file string：指定 Layer 文件的路径。
    --format string：使用自定义模板格式化输出。
    -h, --help：获取命令帮助信息。
    -p, --prettier：以漂亮的 JSON 格式输出。

### 推送玲珑 layer 文件

linglong-tools push 命令用于将 Linglong Layer 文件推送到远程仓库。

示例

```bash
# 使用环境变量 $LINGLONG_USERNAME 和 $LINGLONG_PASSWORD 推送
linglong-tools push -f ./test.layer -r https://repo.linglong.dev

# 通过url指定的用户名和密码推送
linglong-tools push -f ./test.layer -r https://user:pass@repo.linglong.dev

# 通过指定仓库名称推送
linglong-tools push -f ./test.layer -r https://repo.linglong.dev -n develop-snipe
```

参数

    -c, --channel string：指定远程仓库的通道（默认为 "linglong"）。
    -f, --file string：指定 Layer 文件的路径。
    -h, --help：获取命令帮助信息。
    -n, --name string：指定远程仓库的名称（默认为 "repo"）。
    -p, --print：打印所有状态。
    -r, --repo string：指定远程仓库的 URL。

### 搜索仓库的软件包

linglong-tools search 命令用于从远程仓库搜索应用信息。

~~目前仅支持精准搜索，需要传递应用五元组(id,module,channel,version,arch)，所以只能用来判断应用是否存在仓库。~~

如果不传递 version，返回应用的所有版本。

如果应用不存在会返回空数组`[]`，如果应用存在会返回具体的应用信息`[{...}]`

示例

```bash
# 搜索具有以下信息的应用：id: org.deepin.home, version: 1.5.4 其他信息使用默认值
linglong-tools search -r https://repo.linglong.dev -i org.deepin.home -c main -v 1.5.4 -p
```

参数

    -a, --app_arch string：指定应用的架构（默认为 "x86_64"）。
    -i, --app_id string：指定应用的 ID。
    -m, --app_module string：指定应用的模块（默认为 "runtime"）。
    -v, --app_version string：指定应用的版本。
    -c, --channel string：指定远程仓库的通道（默认为 "linglong"）。
    -h, --help：获取命令帮助信息。
    -n, --name string：指定远程仓库的名称（默认为 "repo"）。
    -p, --prettier：以漂亮的 JSON 格式输出。
    -r, --repo string：指定远程仓库的 URL（默认为 "https://repo.linglong.dev"）。

### 删除仓库的软件包

linglong-tools delete 命令用于从远程仓库删除应用。

示例

```bash
# 删除具有以下信息的应用：id: org.deepin.home, version: 1.5.4
linglong-tools delete -r https://repo.linglong.dev -i org.deepin.home -c main -v 1.5.4
```

标志

    -a, --app_arch string：指定应用的架构（默认为 "x86_64"）。
    -i, --app_id string：指定应用的 ID。
    -m, --app_module string：指定应用的模块（默认为 "runtime"）。
    -v, --app_version string：指定应用的版本。
    -c, --channel string：指定远程仓库的通道（默认为 "linglong"）。
    -h, --help：获取命令帮助信息。
    -n, --name string：指定远程仓库的名称（默认为 "repo"）。
    -r, --repo string：指定远程仓库的 URL（默认为 "https://repo.linglong.dev"）。
