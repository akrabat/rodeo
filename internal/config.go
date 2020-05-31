package internal

import (
	"fmt"
	"github.com/spf13/viper"
)

var config *Config

type Command struct {
	Convert string
	Exiftool string
}

type Flickr struct {
	ApiKey string `mapstructure:"api_key"`
	ApiSecret string `mapstructure:"api_secret"`
	Fullname string `mapstructure:"full_name"`
	OauthToken string `mapstructure:"oauth_token"`
	OauthSecret string `mapstructure:"oauth_token_secret"`
	UserId string `mapstructure:"user_nsid"`
	Username string `mapstructure:"username"`
}

type Resize struct {
	Method  string
	Quality string
	Scale string
}

type Condition struct {
	ExcludesAll []string  `mapstructure:"excludes_all"` // list of keywords that must all not exist on image
	ExcludesAny []string  `mapstructure:"excludes_any"` // list of keywords where any one must not exist on image
	IncludesAll []string  `mapstructure:"includes_all"` // list of keywords that must all exist on image
	IncludesAny []string  `mapstructure:"includes_any"` // list of keywords where any one must exist on image
}

type Album struct {
	Id   string
	Name string
}
func (a Album) String() string {
	return fmt.Sprintf("%s (%s)", a.Name, a.Id)
}

type Action struct {
	Delete bool
	Albums []Album
}
type Rules struct {
	Condition Condition
	Action     Action
}

type Config struct {
	Cmd Command
	Flickr Flickr
	Resize Resize
	Rules []Rules
}

func GetConfig() *Config {
	if config != nil {
		return config
	}

	err := viper.Unmarshal(&config)
	if err != nil {
		fmt.Printf("Unable to decode into struct, %v\n", err)
		return nil
	}

	return config
}