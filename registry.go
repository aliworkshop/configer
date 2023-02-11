package configlib

import (
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"io"
)

type configRegistry struct {
	Registry
	rootViper *viper.Viper
	viper     *viper.Viper
	key       string
}

func New() Registry {
	reader := new(configRegistry)
	reader.viper = viper.New()
	reader.rootViper = reader.viper
	return reader
}

func (cr *configRegistry) SetConfig(k string, v interface{}) {
	cr.rootViper.SetDefault(k, v)
	cr.viper.SetDefault(k, v)
}

func (cr *configRegistry) Root() Registry {
	rootViper := cr.rootViper
	if rootViper == nil {
		rootViper = cr.viper
	}
	return &configRegistry{
		Registry: cr.Registry,
		viper:    rootViper,
	}
}

func (cr *configRegistry) ValueOf(key string) Registry {
	v := cr.viper.Sub(key)
	if v == nil {
		return nil
	}
	return &configRegistry{
		Registry:  cr.Registry,
		rootViper: cr.rootViper,
		viper:     v,
	}
}

func (cr *configRegistry) SetConfigType(in string) {
	cr.viper.SetConfigType(in)
}

func (cr *configRegistry) ReadConfig(opts ...interface{}) error {
	if opts == nil || len(opts) == 0 {
		return errors.New("parameter with type io.Reader required")
	}
	reader, ok := opts[0].(io.Reader)
	if !ok {
		return errors.New("type of given parameter is not io.Reader")
	}
	return cr.viper.ReadConfig(reader)
}

func (cr *configRegistry) Unmarshal(rawVal interface{}, opts ...interface{}) error {
	settings := cr.viper.AllSettings()
	var input interface{} = settings
	var options []viper.DecoderConfigOption
	if opts != nil && len(opts) > 0 {
		options := make([]viper.DecoderConfigOption, 0)
		for _, opt := range opts {
			options = append(options, opt.(viper.DecoderConfigOption))
		}
	}
	if cr.key != "" {
		var ok bool
		input, ok = settings[cr.key]
		if !ok {
			return errors.New(fmt.Sprintf("config with key %s not found", cr.key))
		}
	}
	config := defaultDecoderConfig(rawVal, options...)

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}

// defaultDecoderConfig returns default mapsstructure.DecoderConfig with suppot
// of time.Duration values & string slices
func defaultDecoderConfig(output interface{}, opts ...viper.DecoderConfigOption) *mapstructure.DecoderConfig {
	c := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           output,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}
