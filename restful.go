package restfulx

import (
	"c5x.io/logx"
	"net/http"
	"strconv"

	restfulSpec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/go-openapi/spec"

	"c5x.io/chassix"
)

func init() {
	chassix.Register(&chassix.Module{
		Name:      chassix.ModuleRestful,
		ConfigPtr: config,
	})
}

// KeyOpenAPITags is a Metadata key for a restful Route
const KeyOpenAPITags = restfulSpec.KeyOpenAPITags

//newPostBuildOpenAPIObjectFunc open api api docs data
func newPostBuildOpenAPIObjectFunc() restfulSpec.PostBuildSwaggerObjectFunc {
	return func(swo *spec.Swagger) {
		config := config.OpenAPI
		swo.Host = config.Host
		swo.BasePath = config.BasePath
		swo.Schemes = config.Schemas
		swo.Info = &spec.Info{
			InfoProps: spec.InfoProps{
				Title:       config.Spec.Title,
				Description: config.Spec.Description,
				Contact: &spec.ContactInfo{
					Name:  config.Spec.Contact.Name,
					Email: config.Spec.Contact.Email,
					URL:   config.Spec.Contact.URL,
				},

				License: &spec.License{
					Name: config.Spec.License.Name,
					URL:  config.Spec.License.URL,
				},
				Version: config.Spec.Version,
			},
		}

		var nTags []spec.Tag
		for _, tag := range config.Tags {
			nTag := spec.Tag{TagProps: spec.TagProps{Name: tag.Name, Description: tag.Description}}

			nTags = append(nTags, nTag)
		}
		swo.Tags = nTags
	}
}

//Serve rest webservice
func Serve(svc []*restful.WebService) {
	log := logx.New().Category("chassix").Component("restful")
	//restful.Filter(restFilters.RequestID)
	//restful.Filter(restFilters.MeasureTime)

	//if enable openapi setting. register swagger ui and apidocs json API.
	if config.OpenAPI.Enabled {
		swaggerUICfg := config.OpenAPI.UI
		//定义swagger文档
		cfg := restfulSpec.Config{
			WebServices:                   svc, // you control what services are visible
			APIPath:                       swaggerUICfg.API,
			PostBuildSwaggerObjectHandler: newPostBuildOpenAPIObjectFunc()}
		restful.DefaultContainer.Add(restfulSpec.NewOpenAPIService(cfg))
		http.Handle(swaggerUICfg.Entrypoint, http.StripPrefix(swaggerUICfg.Entrypoint, http.FileServer(http.Dir(swaggerUICfg.Dist))))
	}
	//启动服务
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.Server.Port), nil))
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
