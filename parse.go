package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/fatih/color"
)

// Task 结构体
type Task struct {
	Cmds   []string
	Desc   string
	Silent bool
}
type Config struct {
	Vars  map[string]any
	Tasks map[string]Task
}

func parse() (*Config, error) {
	wd, err := os.Getwd()
	if err != nil {
		msg := color.RedString("%s", err)
		return nil, errors.New(msg)
	}

	file_path := filepath.Join(wd, TEMPLATE_NAME)

	if _, err := os.Stat(file_path); errors.Is(err, fs.ErrNotExist) {
		msg := color.RedString("the config file Taskfile.toml do not exist in current directory")
		return nil, errors.New(msg)
	}
	var config Config
	var raw map[string]map[string]any

	if _, err := toml.DecodeFile(file_path, &raw); err != nil {
		color.Red("[parse faild] %s\n", err)
	}
	// 解析 [var]
	if vars, ok := raw["var"]; ok {
		config.Vars = make(map[string]any)
		for field_name, value := range vars {
			switch v := value.(type) {
			case string:
				config.Vars[field_name] = v
			case int64:
				config.Vars[field_name] = strconv.FormatInt(v, 10)
			case []any:
				for _, item := range v {
					switch it := item.(type) {
					case string:
						config.Vars[field_name] = it
					case int64:
						config.Vars[field_name] = append(config.Vars[field_name].([]string), strconv.FormatInt(it, 10))
					default:
						msg := color.RedString("[parse failed] the array of %s field in var contain unknown type, it is one of string, int", field_name)
						return nil, errors.New(msg)
					}

				}
			default:
				msg := color.RedString("[parse failed] unknown type in %s field of var, it is one of string, string array, int.", field_name)
				return nil, errors.New(msg)
			}
		}
	}

	// 解析 [task]
	if tasks, ok := raw["task"]; ok {
		config.Tasks = make(map[string]Task)
		for task_name, task_field := range tasks {
			var task Task
			for field, value := range task_field.(map[string]any) {
				switch field {
				case "cmds":
					if v, ok := value.(string); ok {
						task.Cmds = append(task.Cmds, v)
					} else if v, ok := value.([]any); ok { // 数组必须断言为 []any，在逐元素断言，因为无法确定数组中元素类型， toml 数组元素类型任意
						for _, item := range v {
							if va, ok := item.(string); !ok {
								msg := color.RedString("[parse failed] unknown type in cmds field of %s task, it should be string or string array.", task_name)
								return nil, errors.New(msg)
							} else {
								task.Cmds = append(task.Cmds, va)
							}
						}
					} else {
						msg := color.RedString("[parse failed] unknown type in cmds field of %s task, it should be string or string array.", task_name)
						return nil, errors.New(msg)
					}
				case "desc":
					if v, ok := value.(string); ok {
						task.Desc = v
					} else {
						msg := color.RedString("[parse failed] unknown type in desc field of %s task, it should be string.", task_name)
						return nil, errors.New(msg)
					}
				case "silent":
					if v, ok := value.(bool); ok {
						task.Silent = v
					} else {
						msg := color.RedString("[parse failed] unknown type in silent field of %s task, it should be bool.", task_name)
						return nil, errors.New(msg)
					}
				default:
					msg := color.RedString("[parse failed] unknown %s field of %s task", field, task_name)
					return nil, errors.New(msg)
				}
			}
			config.Tasks[task_name] = task
		}
	}

	// check
	if config.Tasks == nil {
		msg := color.RedString("[parse failed] the top-level task field is required.")
		return nil, errors.New(msg)
	}
	for task_name, task := range config.Tasks {
		if task.Cmds == nil {
			msg := color.RedString("[parse failed] the task of %s require cmds field.", task_name)
			return nil, errors.New(msg)
		}
	}

	return &config, nil
}

// "echo {{dir}}" ==> {{dir}} 替换成字符串
func replace_variables(config *Config) error {
	for _, task := range config.Tasks {
		for i, command := range task.Cmds {
			replaced_command, err := rv(command, get_varposition(command), config)

			if err != nil {
				return err
			}
			task.Cmds[i] = replaced_command
		}
	}
	return nil
}

func rv(command string, var_pos []VarPostion, config *Config) (string, error) {
	if var_pos == nil {
		return command, nil
	}
	var replaced_command strings.Builder // 避免在 loop 中迭代字符串
	replaced_command.Grow(len(command))  // 替换后长度通常接近原长，预估一下
	last_pos := 0
	for _, var_pos := range var_pos {
		wrap_symbol_start := var_pos.warp_symbol_pos[0]
		wrap_symbol_end := var_pos.warp_symbol_pos[1]
		identifier := command[var_pos.unwrap_identifiler_pos[0]:var_pos.unwrap_identifiler_pos[1]]
		if value, ok := config.Vars[identifier]; ok {
			switch v := value.(type) {
			case string:
				replaced_command.WriteString(command[last_pos:wrap_symbol_start])
				replaced_command.WriteString(v)
			case []string:
				replaced_command.WriteString(command[last_pos:wrap_symbol_start])
				replaced_command.WriteString(strings.Join(v, " "))
			default:
				msg := color.RedString("[inner error] config.Vars exist unknown type, not string or string array.")
				return "", errors.New(msg)
			}
		} else {
			msg := color.RedString("[parse failed] undefined identifier %s", identifier)
			return "", errors.New(msg)
		}
		last_pos = wrap_symbol_end
	}

	replaced_command.WriteString(command[last_pos:])

	return replaced_command.String(), nil

}

type VarPostion struct {
	warp_symbol_pos        [2]int
	unwrap_identifiler_pos [2]int
}

func new_position(wrap_symbol_pos []int, unwrap_identifiler_pos []int) VarPostion {
	return VarPostion{
		[2]int(wrap_symbol_pos),
		[2]int(unwrap_identifiler_pos),
	}
}

var var_symbol = regexp.MustCompile(`{{\s*(\w*)\s*}}`)

// 返回一条命令中所有 {{ }} 变量替换的位置
func get_varposition(command string) []VarPostion {
	var var_positions []VarPostion
	for _, wrap_loc := range var_symbol.FindAllStringSubmatchIndex(command, -1) {
		var_position := new_position(wrap_loc[0:2], wrap_loc[2:4])
		var_positions = append(var_positions, var_position)
	}
	return var_positions
}

func debug(config Config) {
	if config.Vars != nil {
		fmt.Println("[var]")

		for variable, value := range config.Vars {
			fmt.Println(variable, ": ", value)
		}
	}

	if config.Tasks != nil {
		for task_name, task := range config.Tasks {
			fmt.Printf("[task.%s]\n", task_name)
			fmt.Println("cmds: ", task.Cmds)
			fmt.Println("desc: ", task.Desc)
			fmt.Println("silent: ", task.Silent)
		}
	}
}
