package datadog

// contains primitives for working with objects that datadog expects

type CheckConf struct {
	InitConfig map[string]interface{} `yaml:"init_config"`
	Instances  []interface{}          `yaml:"instances"`
}
