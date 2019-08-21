package main

import (
	"./go4web"
)

//程序入口
func main(){

	gf := go4web.CreateApp()
	/*
	h := go4web.CreateHttpHandler(
		func(params *go4web.Params) bool {
			params.ResponseWriter.Write(utils.StringToBytes("Hello,"+params.GetParam["name"]))
			return  true
		})

	ok := gf.AddHttpUrl("GET", "/test", h)
	if !ok {
		print("此路径已存在")
	}

	w := go4web.CreateWsHandler(
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
	gf.AddWsUrl("/wstest",w)

	//http.Handle("/static/", http.StripPrefix("/static/",http.FileServer(http.Dir("./view"))))

	gf.AddHtmlUrl("/static/","./view")
	*/
	gf.StartFrame(":8088")
}
