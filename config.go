package restfulx

type RestfulConfig struct {
	OpenAPI OpenAPIConfig  `yaml:"openapi"`
	Servers []ServerConfig `yaml:"servers"`
}

//ServerConfig
type ServerConfig struct {
	Name        string
	Addr        string
	Description string
	OpenAPI     OpenAPIConfig `yaml:"openapi"`
}

//OpenAPIConfig open api config
type OpenAPIConfig struct {
	Enabled  bool     `yaml:"enabled"`
	Host     string   `yaml:"host"`
	BasePath string   `yaml:"basePath"`
	Schemas  []string `yaml:"schemas"`
	Auth     string
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
	URL        string `yaml:"url"`
	API        string `yaml:"api"`
	Dist       string `yaml:"dist"`
	Entrypoint string `yaml:"entrypoint"`
	External   string `yaml:"external"`
}

//OpenapiTagConfig openapi tag
type OpenapiTagConfig struct {
	Name        string
	Description string `yaml:"desc"`
}

var config = new(RestfulConfig)
