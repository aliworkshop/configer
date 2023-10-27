package configer

import "time"

type Registry interface {
	// get root config
	Root() Registry
	// get nested config of root
	SetConfig(k string, v interface{})
	ValueOf(key string) Registry
	SetConfigType(in string)
	ReadConfig(opts ...interface{}) error
	Unmarshal(rawVal interface{}, opts ...interface{}) error
	GetDuration(key string) time.Duration
}
