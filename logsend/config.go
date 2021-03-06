package logsend

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"study2016/logshot/logger"
)

//配置结构
type Configuration struct {
	WatchDir          string
	ReadWholeLog      bool
	ReadAlway         bool
	SenderName        string
	registeredSenders map[string]*SenderRegister
	IsPoll            bool
}

var Conf = &Configuration{
	WatchDir:          "",
	registeredSenders: make(map[string]*SenderRegister),
}

//默认配置结构
var (
	rawConfig = make(map[string]interface{}, 0)
)

//载入默认配置
func LoadRawConfig(f *flag.Flag) {
	rawConfig[f.Name] = f.Value
}

//载入自定义配置文件
func LoadConfigFromFile(fileName string) (rule *Rule, err error) {
	config := ReadConfig(fileName)
	var mysender Sender
	conifg_sender_name := Conf.SenderName
	for sender_name, register := range Conf.registeredSenders {
		//使用指定的sender
		if sender_name != conifg_sender_name {
			continue
		}
		if val, ok := config[sender_name]; ok {
			sender, err := register.InitSender(val)
			if err != nil {
				panic(err)
			}
			mysender = sender
			break
		}
	}
	watch_dir, _ := config["agent"]["watchDir"]
	regexp, _ := config["agent"]["regexp"]
	rule, err = NewRule(regexp, watch_dir)
	if err != nil {
		panic(err)
	}
	//判断sender是否存在
	if mysender == nil {
		panic(errors.New("No sender found !"))
	}
	rule.sender = mysender
	return
}

//读取ini格式的配置文件
func ReadConfig(cfgFile string) map[string]map[string]string {
	fin, err := os.OpenFile(cfgFile, os.O_RDWR, 0644)
	defer fin.Close()
	if err != nil {
		fmt.Println(err)
		logger.GetLogger().Errorln(err.Error())
	}
	config := make(map[string]map[string]string)
	config[""] = make(map[string]string)
	var section = ""
	scanner := bufio.NewScanner(fin)
	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), " ")
		if line == "" || line[0] == ';' || line[0] == '#' {
			//这行是注释，跳过
			continue
		}
		lSqr := strings.Index(line, "[")
		rSqr := strings.Index(line, "]")
		if lSqr == 0 && rSqr == len(line)-1 {
			section = line[lSqr+1 : rSqr]
			_, ok := config[section]
			if !ok {
				config[section] = make(map[string]string)
			}
			continue
		}

		parts := strings.Split(line, "=")
		if len(parts) == 2 {
			key := strings.Trim(parts[0], " ")
			val := strings.Trim(parts[1], " ")
			config[section][key] = val
		}
	}
	return config
}
