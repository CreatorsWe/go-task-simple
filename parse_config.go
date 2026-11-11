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

	yaml "github.com/goccy/go-yaml"
)

// Task 结构体
type Task struct {
	Cmds     []string `yaml:"cmds"`
	Desc     string   `yaml:"desc"`
	Silent   bool     `yaml:"silent"`
	Parallel bool     `yaml:"parallel"`
	Continue bool     `yaml:"continue"`
	Timeout  uint64   `yaml:"timeout"`
	Shell    bool     `yaml:"shell"`
}
type Config struct {
	Vars  map[string]any  `yaml:"vars"`
	Tasks map[string]Task `yaml:"tasks"`
}

func parse_config(flag_vars map[string]string) *Config {
	wd, err := os.Getwd()
	if err != nil {
		msg := error_msg_without_run(err.Error())
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}

	file_path := filepath.Join(wd, TEMPLATE_NAME)

	if _, err := os.Stat(file_path); errors.Is(err, fs.ErrNotExist) {
		msg := error_msg_without_run("the config file Taskfile.yml do not exist in current directory")
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}

	data, err := os.ReadFile(file_path)
	if err != nil {
		msg := error_msg_without_run("read configure file fail: %s", err)
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}

	var config Config

	if err := yaml.Unmarshal(data, &config); err != nil {
		msg := error_msg_without_run(err.Error())
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}

	// flags 改变 vars
	for vari, value := range flag_vars {
		if _, ok := config.Vars[vari]; ok {
			config.Vars[vari] = value
		} else {
			msg := error_msg_without_run("unknown flag variable: %s", vari)
			fmt.Println(msg)
			os.Exit(1)
		}
	}

	// repalce variable
	if err := _replace_variables(&config); err != nil {
		msg := error_msg_without_run(err.Error())
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}

	return &config
}

func (c *Config) UnmarshalYAML(data []byte) error {
	var raw struct {
		Vars  map[string]any  `yaml:"vars"`
		Tasks map[string]Task `yam:"tasks"`
	}

	if err := yaml.Unmarshal(data, &raw); err != nil {
		return err
	}

	c.Vars = make(map[string]any, len(raw.Vars))

	for k, v := range raw.Vars {
		if err := _assert_vars(k, v, c.Vars); err != nil {
			return err
		}
	}

	c.Tasks = raw.Tasks
	return nil
}

// vars 值类型有 string 和 []string 因此需要 any 类型
func _assert_vars(key string, value any, vars map[string]any) error {
	switch x := value.(type) {
	case string:
		vars[key] = x
	case int:
		vars[key] = strconv.Itoa(x)
	case int64:
		vars[key] = strconv.FormatInt(x, 10)
	case uint64:
		vars[key] = strconv.FormatUint(x, 10)
	case float64:
		vars[key] = strconv.FormatFloat(x, 'f', 3, 64)
	case []any:
		var items []string
		for _, item := range x {
			switch i := item.(type) {
			case string:
				items = append(items, i)
			case int:
				items = append(items, strconv.Itoa(i))
			case int64:
				items = append(items, strconv.FormatInt(i, 10))
			case uint64:
				items = append(items, strconv.FormatUint(i, 10))
			case float64:
				items = append(items, strconv.FormatFloat(i, 'f', 3, 64))
			default:
				return fmt.Errorf("vars.%s: only support string, int, and array consisting of them ", key)
			}
		}
		vars[key] = items
	default:
		return fmt.Errorf("vars.%s: only support string, int, and array consisting of them", key)
	}
	return nil
}

// "echo {{dir}}" ==> {{dir}} 替换成字符串
func _replace_variables(config *Config) error {
	for _, task := range config.Tasks {
		for i, command := range task.Cmds {
			replaced_command, err := _rv(command, _get_varposition(command), config)

			if err != nil {
				return err
			}
			task.Cmds[i] = replaced_command
		}
	}
	return nil
}

func _rv(command string, var_pos []VarPostion, config *Config) (string, error) {
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
				panic("var value should be  string or []string type in logic")
			}
		} else {
			msg := fmt.Sprintf("undefined variable %s", identifier)
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

func _new_position(wrap_symbol_pos []int, unwrap_identifiler_pos []int) VarPostion {
	return VarPostion{
		[2]int(wrap_symbol_pos),
		[2]int(unwrap_identifiler_pos),
	}
}

var var_symbol = regexp.MustCompile(`{{\s*(\w*)\s*}}`)

// 返回一条命令中所有 {{ }} 变量替换的位置
func _get_varposition(command string) []VarPostion {
	var var_positions []VarPostion
	for _, wrap_loc := range var_symbol.FindAllStringSubmatchIndex(command, -1) {
		var_position := _new_position(wrap_loc[0:2], wrap_loc[2:4])
		var_positions = append(var_positions, var_position)
	}
	return var_positions
}

func _debug(config Config) {
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
			fmt.Println("timeout:", task.Timeout)
			fmt.Println("parallel:", task.Parallel)
		}
	}
}
