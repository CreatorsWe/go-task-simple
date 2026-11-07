### `task`

task 是类似 make 这类的自动化执行脚本，配置文件名为 `Taskfile.yml`，仅提供核心且必要的功能。

### `Taskfile.yml`

只有一些简单的配置项：
```yml
  vars:
    greet: hello, world!
    count: 3
  tasks:
    default:
      desc: echo commands
      cmds:
        - echo {{greet}}      # {{ variable }} 使用变量 
        - echo \{{ greet \}}  # 转移字符，输出原字符
        - echo 3
      slient: false   # 不打印输出
      timeout: 4      # 超时自动退出
      parallel: true  # 并行执行命令
      continue: true  # 执行失败继续执行下一条任务
    build:
      desc: build project
      cmds:
        - cargo build
      silent: true
```

### `sub-command`

```bash
  task <task-name> 
  task --help
  task --version
  task --list   # 列出任务信息
  task --init
```

### `output-explain`

+ 非命令执行错误：task 程序本身执行错误，输出错误信息到 stderr ，终止程序。
+ 命令正常执行： output（标准输出） + success prompt([task]step %n: execute success.)；
  + silent 时静默；
  + parallel 时, 打印信息要加锁输出（收集后输出，避免阻塞命令执行阶段）
+ 命令执行错误：output( 标准输出 ) + failed prompt([task]step %n: execute failed: %err: exitcode status: %code.)；
  + 错误信息不会静默，一定会打印
  + continue 时，打印错误信息，继续执行下一条任务而不退出