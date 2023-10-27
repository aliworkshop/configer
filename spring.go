package configer

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type spring struct {
	Address  string
	ApiToken string
	Repo     string

	key   string
	viper *viper.Viper
	root  *viper.Viper
}

func NewSpring(Address, token, repo string) Registry {
	s := &spring{
		Address:  Address,
		ApiToken: token,
		Repo:     strings.Trim(repo, "/"),
		viper:    viper.New(),
	}
	s.root = s.viper
	return s
}

func (s *spring) ReadConfig(opts ...interface{}) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", s.Address, s.Repo), nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Token", s.ApiToken)
	cl := &http.Client{Timeout: time.Second * 5}
	resp, err := cl.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := ioutil.ReadAll(resp.Body)
		return errors.New("config server returned non 200 status. err: " + string(b))
	}
	s.viper.SetConfigType("yaml")
	err = s.viper.ReadConfig(resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func (s *spring) Root() Registry {
	return &spring{
		Address:  s.Address,
		ApiToken: s.ApiToken,
		Repo:     s.Repo,
		viper:    s.root,
	}
}

func (s *spring) ValueOf(key string) Registry {
	v := s.viper.Sub(key)
	if v == nil {
		return nil
	}
	return &spring{
		Address:  s.Address,
		ApiToken: s.ApiToken,
		Repo:     s.Repo,
		viper:    v,
		root:     s.root,
	}
}

func (s *spring) SetConfigType(in string) {
	s.viper.SetConfigType(in)
}

func (s *spring) Unmarshal(rawVal interface{}, opts ...interface{}) error {
	settings := s.viper.AllSettings()
	var input interface{} = settings
	var options []viper.DecoderConfigOption
	if opts != nil && len(opts) > 0 {
		options := make([]viper.DecoderConfigOption, 0)
		for _, opt := range opts {
			options = append(options, opt.(viper.DecoderConfigOption))
		}
	}
	if s.key != "" {
		var ok bool
		input, ok = settings[s.key]
		if !ok {
			return errors.New(fmt.Sprintf("config with key %s not found", s.key))
		}
	}
	config := defaultDecoderConfig(rawVal, options...)

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}

func (s *spring) SetConfig(k string, v interface{}) {
	s.root.SetDefault(k, v)
	s.viper.SetDefault(k, v)
}

func (s *spring) GetDuration(key string) time.Duration {
	return s.viper.GetDuration(key)
}
