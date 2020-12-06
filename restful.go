package restfulx

import (
	"fmt"
	restfulSpec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/go-openapi/spec"
	"github.com/imdario/mergo"
	"log"
	"net/http"
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
func newPostBuildOpenAPIObjectFunc(config ServerConfig) restfulSpec.PostBuildSwaggerObjectFunc {
	return func(swo *spec.Swagger) {
		swo.Host = config.OpenAPI.Host
		swo.BasePath = config.OpenAPI.BasePath
		swo.Schemes = config.OpenAPI.Schemas

		var title, description string
		if config.Name != "" {
			title = config.Name
		} else {
			title = config.OpenAPI.Spec.Title
		}
		if config.Description != "" {
			description = config.Description
		} else {
			description = config.OpenAPI.Spec.Description
		}
		swo.Info = &spec.Info{
			InfoProps: spec.InfoProps{
				Title:       title,
				Description: description,
				Contact: &spec.ContactInfo{
					ContactInfoProps: spec.ContactInfoProps{
						Name:  config.OpenAPI.Spec.Contact.Name,
						Email: config.OpenAPI.Spec.Contact.Email,
						URL:   config.OpenAPI.Spec.Contact.URL,
					},
				},

				License: &spec.License{
					LicenseProps: spec.LicenseProps{
						Name: config.OpenAPI.Spec.License.Name,
						URL:  config.OpenAPI.Spec.License.URL,
					},
				},
				Version: config.OpenAPI.Spec.Version,
			},
		}

		var nTags []spec.Tag
		var tags []OpenapiTagConfig
		if len(config.OpenAPI.Tags) > 0 {
			tags = config.OpenAPI.Tags
		} else {
			tags = config.OpenAPI.Tags
		}
		for _, tag := range tags {
			nTag := spec.Tag{TagProps: spec.TagProps{Name: tag.Name, Description: tag.Description}}

			nTags = append(nTags, nTag)
		}
		swo.Tags = nTags
		// setup security definitions
		if config.OpenAPI.Auth == "basic" {
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

	var corsAllowedHost, redirectURL, schema string

	corsAllowedHost = serverCfg.OpenAPI.Host

	//if enable openapi setting. register swagger ui and apidocs json API.
	if serverCfg.OpenAPI.Enabled {

		if len(serverCfg.OpenAPI.Schemas) > 0 {
			schema = serverCfg.OpenAPI.Schemas[0]
		} else {
			schema = "http"
		}
		swaggerUICfg := serverCfg.OpenAPI.UI
		//定义swagger文档
		cfg := restfulSpec.Config{
			WebServices:                   container.RegisteredWebServices(), // you control what services are visible
			APIPath:                       swaggerUICfg.API,
			PostBuildSwaggerObjectHandler: newPostBuildOpenAPIObjectFunc(serverCfg)}
		container.Add(restfulSpec.NewOpenAPIService(cfg))
		//if setting swagger ui dist will handle swagger ui route
		if serverCfg.OpenAPI.Enabled && swaggerUICfg.External != "" {

			apiPath := schema + "://" + path.Join(serverCfg.OpenAPI.Host, serverCfg.OpenAPI.UI.API)
			redirectURL = fmt.Sprintf("%s?url=%s", swaggerUICfg.External, apiPath)

			log.Debugf("swagger ui: %s", redirectURL)
			container.ServeMux.HandleFunc("/open_apidocs", func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, redirectURL, 302)
			})
			//为OPENAPI添加cors跨域支持
			if corsAllowedHost != "" {
				log.Debugf("cors allowed host %s", corsAllowedHost)
				// Add container filter to enable CORS
				cors := restful.CrossOriginResourceSharing{

					AllowedDomains: []string{corsAllowedHost},
					//ExposeHeaders:  []string{"X-My-Header"},
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
				//uiURL, err := url.Parse(serverCfg.OpenAPI.Host)
				//copy
				//apiURL := *uiURL
				//if err != nil {
				//	log.Fatalln("swagger ui URL invalid\n", err)
				//}

				uiPath := schema + "://" + path.Join(serverCfg.OpenAPI.Host, swaggerUICfg.Entrypoint)
				apiPath := schema + "://" + path.Join(serverCfg.OpenAPI.Host, swaggerUICfg.API)
				redirectURL = fmt.Sprintf("%s?url=%s",
					uiPath,
					apiPath)
				log.Debugf("swagger ui: %s", redirectURL)
				container.ServeMux.HandleFunc("/open_apidocs", func(w http.ResponseWriter, r *http.Request) {
					http.Redirect(w, r, redirectURL, 302)
				})
			}
		}
	}
	//启动服务
	fmt.Printf("[%s] starting [http://%s]\tapidocs:[http://%s]\n", serverCfg.Name, serverCfg.Addr, serverCfg.Addr+"/open_apidocs")
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
