# gin-swagger

# install swag 
go get github.com/swaggo/swag/cmd/swag

# add gin-swagger annotations such as:
## project swagger config
// @title gin-swagger测试
// @version 1.0
func main() {
	r := gin.Default()
	r.GET("/hi", controller.Record)
	r.GET("/hello", hello)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Run()
}

## http  swagger config
// @Tags 第二组
// @Summary 测试接口
// @Description 描述信息
// @Param name query string true "名称"   //参数依次为 参数名称name query传值 参数string类型 参数必填   参数描述
// @Success 200 {string} string    "ok"
// @Router /hello [get]
func hello(c *gin.Context)  {
	name,_:=c.GetQuery("name")
	fmt.Println("--------------"+name)
	c.String(200, name)
}
#generate swagger docs
swag init

# import docs in main package 
import _ "project_name/docs"

# doc url
http://ip:port/swagger/index.html

# more doc
https://github.com/swaggo/gin-swagger
https://swaggo.github.io/swaggo.io/declarative_comments_format/