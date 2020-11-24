package restfulx

import (
	"fmt"
	restfulSpec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/go-openapi/spec"
	"log"
	"net/http"

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
	//if enable openapi setting. register swagger ui and apidocs json API.
	if serverCfg.OpenAPI.Enabled {
		swaggerUICfg := config.OpenAPI.UI
		//定义swagger文档
		cfg := restfulSpec.Config{
			WebServices:                   container.RegisteredWebServices(), // you control what services are visible
			APIPath:                       swaggerUICfg.API,
			PostBuildSwaggerObjectHandler: newPostBuildOpenAPIObjectFunc(servIndex)}
		container.Add(restfulSpec.NewOpenAPIService(cfg))
		container.Handle(swaggerUICfg.Entrypoint, http.StripPrefix(swaggerUICfg.Entrypoint, http.FileServer(http.Dir(swaggerUICfg.Dist))))
	}
	//启动服务
	fmt.Printf("server [%s] starting [http://%s]\n", serverCfg.Name, serverCfg.Addr)
	if serverCfg.OpenAPI.Enabled && config.OpenAPI.UI.Entrypoint != "" {
		fmt.Printf("server [%s] apidocs addr [http://%s]\n", serverCfg.Name, serverCfg.Addr+config.OpenAPI.UI.Entrypoint)
	}
	log.Fatal(http.ListenAndServe(serverCfg.Addr, container.ServeMux))
}

//ServeDefault serve with default container and first server config
//func Serve(svc []*restful.WebService) {
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
