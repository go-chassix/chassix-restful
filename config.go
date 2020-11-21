package restfulx

type RestfulConfig struct {
	OpenAPI OpenAPIConfig `yaml:"openapi"`
	Server  ServerConfig  `yaml:"server"`
}

//ServerConfig
type ServerConfig struct {
	Port int
	Addr string
}

//OpenAPIConfig open api config
type OpenAPIConfig struct {
	Enabled  bool     `yaml:"enabled"`
	Host     string   `yaml:"host"`
	BasePath string   `yaml:"basePath"`
	Schemas  []string `yaml:"schemas"`
	Spec     struct {
		Title       string
		Description string `yaml:"desc"`
		Contact     struct {
			Name  string
			Email string
			URL   string
		} `yaml:"contact"`
		License struct {
			Name string
			URL  string
		} `yaml:"license"`
		Version string
	}
	Tags []OpenapiTagConfig `yaml:",flow"`
	UI   OpenapiUIConfig    `yaml:"ui"`
}

//OpenapiUIConfig swagger ui config
type OpenapiUIConfig struct {
	API        string `yaml:"api"`
	Dist       string `yaml:"dist"`
	Entrypoint string `yaml:"entrypoint"`
}

//OpenapiTagConfig openapi tag
type OpenapiTagConfig struct {
	Name        string
	Description string `yaml:"desc"`
}

var config = new(RestfulConfig)
