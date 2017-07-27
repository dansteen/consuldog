package datadog

// contains primitives for working with objects that datadog expects

type CheckConf struct {
	InitConfig []interface{} `yaml:"init_config"`
	Instances  []interface{} `yaml:"instances"`
}
