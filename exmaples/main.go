package main

import (
	"c5x.io/chassix"
	"c5x.io/restfulx"
	"github.com/emicklei/go-restful/v3"
)

type HelloResource struct {
}

func (h HelloResource) hello(_ *restful.Request, res *restful.Response) {
	res.WriteEntity([]string{"world"})
}

func (h HelloResource) Webservice() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/").Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)

	ws.Route(ws.GET("hello").To(h.hello).Doc("hello").Returns(200, "success", []string{}))
	return ws
}

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}
type UserResource struct {
	users map[string]User
}

func (u UserResource) getUser(req *restful.Request, res *restful.Response) {
	uid := req.PathParameter("user_id")
	if user, ok := u.users[uid]; ok {
		res.WriteEntity(user)
		return
	}
	res.WriteHeader(404)
}
func (u UserResource) Webservice() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/api/v1").Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	ws.Route(ws.GET("users/{user_id}").To(u.getUser).Doc("获取用户列表").Param(
		ws.PathParameter("user_id", "用户id").
			DefaultValue("test").
			DataType("string")).Returns(200, "success", []User{}))
	restfulx.AddMetaDataTags(ws, []string{"User"})
	return ws
}

func main() {
	chassix.Init()

	go func() {
		container := restful.NewContainer()
		users := make(map[string]User)
		users["test"] = User{
			Name: "Test",
			Age:  10,
		}
		container.Add(UserResource{users: users}.Webservice())

		restfulx.Serve(container, 2)
	}()

	restful.Add(HelloResource{}.Webservice())
	// Add container filter to enable CORS
	//cors := restful.CrossOriginResourceSharing{
	//
	//	//ExposeHeaders:  []string{"X-My-Header"},
	//	AllowedHeaders: []string{"Content-Type", "Accept"},
	//	AllowedMethods: []string{"GET", "POST"},
	//	CookiesAllowed: false,
	//	Container:      restful.DefaultContainer}
	//restful.DefaultContainer.Filter(cors.Filter)
	//
	//// Add container filter to respond to OPTIONS
	//restful.DefaultContainer.Filter(restful.DefaultContainer.OPTIONSFilter)
	restfulx.ServeDefault()
}
