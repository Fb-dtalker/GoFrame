# 框架名称：GoFrame

## -已实现的功能-

1>HTTP请求的处理

2>HTTP请求参数的封装【使用map占用较大】

3>Websocket请求的处理【发送数据长度限制至4294967295byte以内的文本类型】

4>静态html页面请求的处理

5>静态文件请求的处理

6>shtml模板页面的渲染



## -代码说明-

```go

```

```go
//go4web.go文件下

func CreateApp() (*GoFrame){}
//创建GoFrame框架实例

func (goFrame *GoFrame) StartFrame(port string){}
//启动GoFrame框架实例

func (goFrame *GoFrame) ServeHTTP(responeWriter http.ResponseWriter, request *http.Request){}
//GoFrame框架的路由请求分发器

func (goFrame *GoFrame) AddHttpUrl(method,path string, handle *HttpHandler) (bool){}
//添加Http请求处理

func (goFrame *GoFrame) AddWsUrl(path string, handler *WsHandler ) (bool){}
//添加Websocket请求处理

func (wsHandler *WsHandler) StartLink(params *WsParams){}
//启动一个websocket链接

func (wsHandler *WsHandler) EndLink(webSocketContext *WebSocketContext){}
//关闭一个websocket链接

func (goFrame *GoFrame) AddHtmlUrl(url string, staticPath string){}
//添加静态html文件请求处理

func (goFrame *GoFrame) RegisterShtmlView(templatePattern string){}
//添加shtml模板

func (goFrame *GoFrame) AddStaticUrl(url string, staticPath string){}
//添加静态文件的请求处理

type GoFrame struct{}
//框架主体

type Route struct {}
//路由管理器

type Params struct {}
//http参数

type WsParams struct {}
//websocket参数

type HttpHandler struct{}
//http处理函数handler的封装

func CreateHttpHandler(do doFunc) (*HttpHandler){}
//创建一个对应Http请求的handler，用于新建路由时使用
//当只需要一个处理函数时

func CreateHttpHandlerWithList(before []beforeFunc, do []doFunc, after []afterFunc) (*HttpHandler, error){}
//创建一个对应Http请求的handler，用于新建路由时使用
//当可能需要预处理函数时，before和after模仿springboot的责任链栈

func (httpHandler *HttpHandler) ExecuteHandler(params *Params){}
//Http请求分发后的执行者

```

```go
//使用案例
package main

import (
	"./go4web"
	"./go4web/utils"
	"fmt"
)

//程序入口
func main(){

	gf := go4web.CreateApp()

	h := go4web.CreateHttpHandler( //创建httpHandler
		func(params *go4web.Params) bool {
			params.ResponseWriter.Write(utils.StringToBytes("Hello,"+params.GetParam["name"]))
			return  true
		})
    
    //h2 := go4web.CreateHttpHandlerWithList(before []beforeFunc, do []doFunc, after []afterFunc)
    //可以给httpHandler添加多个处理函数但是要确保使用responseWrite时返回值为false
    //当函数返回值为false时将不会继续往下执行

	ok := gf.AddHttpUrl("GET", "/test", h) //绑定httpHandler至路由
	if !ok {
		print("此路径已存在")
	}

	w := go4web.CreateWsHandler( //创建一个WsHandler，需要4个处理函数
		func(context *go4web.WebSocketContext) bool {
			fmt.Print("打开链接!")
			context.Handler.SendMessage(context, "成功建立与后端的链接")
			return true
		},
		func(context *go4web.WebSocketContext, message string) {
			fmt.Print("收到消息:"+message+"\n")
			context.Handler.SendMessage(context, "告知前端：后端收到消息！")
		},
		func(context *go4web.WebSocketContext) {
			fmt.Print("关闭链接!")
		},
		func(context *go4web.WebSocketContext) {
			fmt.Print("出错!")
		},
		)
	gf.AddWsUrl("/wstest",w) //绑定wsHandler至路由

    gf.AddHtmlUrl("/static/","./view") //添加文件目录下/view文件夹内容可通过/static/文件名进行访问
    
    gf.AddStaticUrl("/static/file/", "./static/file") //添加文件目录下/static/file文件夹内容可通过/static/file/文件名进行访问
	
    gf.StartFrame(":8088") //以8088端口启动框架
}
```

