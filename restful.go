package restfulx

import (
	"fmt"
	restfulSpec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/go-openapi/spec"
	"github.com/imdario/mergo"
	"log"
	"net/http"
	"net/url"
	"path"

	"c5x.io/chassix"
	"c5x.io/logx"
)

func init() {
	chassix.Register(&chassix.Module{
		Name:      chassix.ModuleRestful,
		ConfigPtr: config,
	})
}

const (
	// KeyOpenAPITags is a Metadata key for a restful Route

	KeyOpenAPITags = restfulSpec.KeyOpenAPITags

	SecurityDefinitionKey = "OAPI_SECURITY_DEFINITION"
)

//newPostBuildOpenAPIObjectFunc open api api docs data
func newPostBuildOpenAPIObjectFunc(serverIndex int) restfulSpec.PostBuildSwaggerObjectFunc {
	return func(swo *spec.Swagger) {
		serverCfg := config.Servers[serverIndex-1]
		config := config.OpenAPI
		swo.Host = serverCfg.Addr
		swo.BasePath = config.BasePath
		swo.Schemes = config.Schemas

		var title, description string
		if serverCfg.Name != "" {
			title = serverCfg.Name
		} else {
			title = config.Spec.Title
		}
		if serverCfg.Description != "" {
			description = serverCfg.Description
		} else {
			description = config.Spec.Description
		}
		swo.Info = &spec.Info{
			InfoProps: spec.InfoProps{
				Title:       title,
				Description: description,
				Contact: &spec.ContactInfo{
					ContactInfoProps: spec.ContactInfoProps{
						Name:  config.Spec.Contact.Name,
						Email: config.Spec.Contact.Email,
						URL:   config.Spec.Contact.URL,
					},
				},

				License: &spec.License{
					LicenseProps: spec.LicenseProps{
						Name: config.Spec.License.Name,
						URL:  config.Spec.License.URL,
					},
				},
				Version: config.Spec.Version,
			},
		}

		var nTags []spec.Tag
		var tags []OpenapiTagConfig
		if len(serverCfg.OpenAPI.Tags) > 0 {
			tags = serverCfg.OpenAPI.Tags
		} else {
			tags = config.Tags
		}
		for _, tag := range tags {
			nTag := spec.Tag{TagProps: spec.TagProps{Name: tag.Name, Description: tag.Description}}

			nTags = append(nTags, nTag)
		}
		swo.Tags = nTags
		// setup security definitions
		if serverCfg.OpenAPI.Auth == "basic" {
			swo.SecurityDefinitions = map[string]*spec.SecurityScheme{
				"basicAuth": spec.BasicAuth(),
			}
			auth := make(map[string][]string)
			auth["basicAuth"] = []string{}
			swo.Security = append(swo.Security, auth)
		}

	}
}

//Serve rest webservice index start from 1
//func Serve(svc []*restful.WebService) {
func Serve(container *restful.Container, servIndex int) {
	if servIndex < 1 || servIndex > len(config.Servers) {
		log.Fatal("server config error, pls check your servers config")
	}
	log := logx.New().Category("chassix").Component("restful")

	serverCfg := config.Servers[servIndex-1]

	//merge config
	//apiCfg := &(serverCfg.OpenAPI)
	mergo.Merge(&(serverCfg.OpenAPI), config.OpenAPI)

	log.Debugf("server [%d] config merged: %+v", servIndex, serverCfg.OpenAPI)

	var corsAllowedHost string

	//if enable openapi setting. register swagger ui and apidocs json API.
	if serverCfg.OpenAPI.Enabled {
		swaggerUICfg := serverCfg.OpenAPI.UI
		//定义swagger文档
		cfg := restfulSpec.Config{
			WebServices:                   container.RegisteredWebServices(), // you control what services are visible
			APIPath:                       swaggerUICfg.API,
			PostBuildSwaggerObjectHandler: newPostBuildOpenAPIObjectFunc(servIndex)}
		container.Add(restfulSpec.NewOpenAPIService(cfg))
		//if setting swagger ui dist will handle swagger ui route
		if serverCfg.OpenAPI.Enabled && swaggerUICfg.External != "" {
			apiUrl, err := url.Parse(serverCfg.OpenAPI.UI.URL)
			if err != nil {
				log.Fatalln("openapi ui url invalid\n", err)
			}
			apiUrl.Path = path.Join(apiUrl.Path, serverCfg.OpenAPI.UI.API)
			corsAllowedHost = apiUrl.Host
			fmt.Printf("server [%s] apidocs addr [%s?url=%s]\n",
				serverCfg.Name,
				swaggerUICfg.External,
				apiUrl.String())
			//为OPENAPI添加cors跨域支持
			if corsAllowedHost != "" {
				// Add container filter to enable CORS
				cors := restful.CrossOriginResourceSharing{

					ExposeHeaders:  []string{"X-My-Header"},
					AllowedHeaders: []string{"Content-Type", "Accept"},
					AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "BATCH"},

					CookiesAllowed: false,
					Container:      container}
				container.Filter(cors.Filter)

				// Add container filter to respond to OPTIONS
				container.Filter(container.OPTIONSFilter)
			}
		} else if serverCfg.OpenAPI.Enabled && swaggerUICfg.Entrypoint != "" && swaggerUICfg.Dist != "" {
			container.Handle(swaggerUICfg.Entrypoint, http.StripPrefix(swaggerUICfg.Entrypoint, http.FileServer(http.Dir(swaggerUICfg.Dist))))
			if serverCfg.OpenAPI.Enabled && config.OpenAPI.UI.Entrypoint != "" {
				uiURL, err := url.Parse(swaggerUICfg.URL)
				//copy
				apiURL := *uiURL
				if err != nil {
					log.Fatalln("swagger ui URL invalid\n", err)
				}
				uiURL.Path = path.Join(uiURL.Path, swaggerUICfg.Entrypoint)
				apiURL.Path = path.Join(apiURL.Path, swaggerUICfg.API)
				fmt.Printf("server [%s] apidocs addr [%s?url=%s]\n",
					serverCfg.Name,
					uiURL.String(),
					apiURL.String())
			}
		}
	}
	//启动服务
	fmt.Printf("server [%s] starting [http://%s]\n", serverCfg.Name, serverCfg.Addr)

	log.Fatal(http.ListenAndServe(serverCfg.Addr, container.ServeMux))
}

//ServeDefault serve with default container and first server config
func ServeDefault() {
	Serve(restful.DefaultContainer, 1)
}

//AddMetaDataTags add metadata tags to Webservice all routes
func AddMetaDataTags(ws *restful.WebService, tags []string) {
	routes := ws.Routes()
	for i, route := range routes {
		if route.Metadata == nil {
			routes[i].Metadata = map[string]interface{}{}
		}
		routeTags := routes[i].Metadata[KeyOpenAPITags]
		if routeTags != nil {
			existedTags, ok := routeTags.([]string)
			if ok {
				existedTags = append(existedTags, tags...)
				routes[i].Metadata[KeyOpenAPITags] = existedTags
			}
			continue
		}
		routes[i].Metadata[KeyOpenAPITags] = tags
	}
}
