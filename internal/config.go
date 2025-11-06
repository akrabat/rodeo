package internal

import (
	"fmt"
	"github.com/spf13/viper"
)

var config *Config

type Command struct {
	Convert  string
	Exiftool string
}

type Flickr struct {
	ApiKey      string `mapstructure:"api_key"`
	ApiSecret   string `mapstructure:"api_secret"`
	Fullname    string `mapstructure:"full_name"`
	OauthToken  string `mapstructure:"oauth_token"`
	OauthSecret string `mapstructure:"oauth_token_secret"`
	UserId      string `mapstructure:"user_nsid"`
	Username    string `mapstructure:"username"`
}

type Upload struct {
	SetDatePosted             bool `mapstructure:"set_date_posted"`
	StoreUploadListInImageDir bool `mapstructure:"store_uploaded_list_in_image_dir"`
}

type Resize struct {
	Method  string
	Quality string
	Scale   string
}

type Condition struct {
	ExcludesAll []string `mapstructure:"excludes_all"` // list of keywords that must all not exist on image
	ExcludesAny []string `mapstructure:"excludes_any"` // list of keywords where any one must not exist on image
	IncludesAll []string `mapstructure:"includes_all"` // list of keywords that must all exist on image
	IncludesAny []string `mapstructure:"includes_any"` // list of keywords where any one must exist on image
}

type Album struct {
	Id   string
	Name string
}

func (a Album) String() string {
	return fmt.Sprintf("%s (%s)", a.Name, a.Id)
}

type Permissions struct {
	Family  bool
	Friends bool
	Public  bool
}

func (p *Permissions) SetDefaults() {
	p.Family = true
	p.Friends = true
	p.Public = true
}

type Action struct {
	Delete  bool
	Privacy *Permissions
	Albums  []Album
}
type Rules struct {
	Name      string
	Condition Condition
	Action    Action
}

type Config struct {
	Cmd    Command
	Flickr Flickr
	Upload Upload
	Resize Resize
	Rules  []Rules
}

func GetConfig() *Config {
	if config != nil {
		return config
	}

	// set defaults
	setDefaults()

	err := viper.Unmarshal(&config)
	if err != nil {
		fmt.Printf("Unable to decode into struct, %v\n", err)
		return nil
	}

	return config
}

func setDefaults() {

	if viper.IsSet("upload.set_date_posted") == false {
		viper.Set("upload.set_date_posted", false)
	}

	if viper.IsSet("upload.store_uploaded_list_in_image_dir") == false {
		viper.Set("upload.store_uploaded_list_in_image_dir", false)
	}

	if viper.IsSet("cmd.convert") == false {
		viper.Set("cmd.convert", "/usr/local/bin/convert")
	}

	if viper.IsSet("resize.scale") == false {
		viper.Set("resize.scale", "2000x2000")
	}
	if viper.IsSet("resize.method") == false {
		viper.Set("resize.method", "catrom")
	}
	if viper.IsSet("resize.quality") == false {
		viper.Set("resize.quality", "75")
	}

	if err := viper.WriteConfig(); err != nil {
		fmt.Println("Error writing config: ", err)
	}
}
