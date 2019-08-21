package go4web

import (
	"./utils"
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"html/template"
	"math"
	"net"
	"net/http"
	"strings"
)

/*
* FB君
* 2019年8月22日
*/

type GoFrame struct {
	route *Route
}

//路由管理
type Route struct {
	httpRoute map[string]*HttpHandler
	websocketRoute map[string]*WsHandler
}

//http请求的参数封装
type Params struct {
	ResponseWriter http.ResponseWriter
	Request *http.Request
	GetParam map[string]string
	PostParam map[string]string
	PostParams map[string][]string
}

//websocket请求的参数封装
type WsParams struct {
	ResponseWriter http.ResponseWriter
	Request *http.Request
	FirstParam map[string]string
}

//http请求的处理函数类型
type beforeFunc = func(*Params) bool
type doFunc = func(*Params) bool
type afterFunc = func(*Params) bool

//http处理函数handler的封装
type HttpHandler struct{
	before []beforeFunc
	do []doFunc
	after []afterFunc
}

//创建一个对应Http请求的handler，用于新建路由时使用
//当只需要一个处理函数时
func CreateHttpHandler(do doFunc) (*HttpHandler){
	var handler *HttpHandler
	handler = new(HttpHandler)
	handler.before = nil
	handler.after = nil
	handler.do = []doFunc{do}
	return handler
}

//创建一个对应Http请求的handler，用于新建路由时使用
//当可能需要预处理函数时，before和after模仿springboot的责任链栈
func CreateHttpHandlerWithList(before []beforeFunc, do []doFunc, after []afterFunc) (*HttpHandler, error){
	handler := new(HttpHandler)
	if (len(do) < 1) {
		return nil, errors.New("lost the doFunc\n(it should have one function at least)")
	}else {
		handler.do = do
	}

	if(len(before) < 1){
		handler.before = nil
	}else{
		handler.before = before
	}

	if (len(after) < 1) {
		handler.after = nil
	}else{
		handler.after = after
	}
	handler.before = before
	handler.do = do
	handler.after = after
	return handler,nil
}

//Http请求分发后的执行者
func (httpHandler *HttpHandler) ExecuteHandler(params *Params){

	length := len(httpHandler.before)
	result := true
	for i := 0; i < length && result; i++ {
		result = httpHandler.before[i](params);
	}
	length = len(httpHandler.do)
	for i := 0; i < length && result; i++ {
		result = httpHandler.do[i](params);
	}
	length = len(httpHandler.after)
	for i := 0; i < length && result; i++ {
		result = httpHandler.after[i](params);
	}
}

//路由请求分发器
func (goFrame *GoFrame) ServeHTTP(responeWriter http.ResponseWriter, request *http.Request){
	url := request.RequestURI;
	path := strings.Split(url, "?")
	method := request.Method; //调用方法

	handle,ok := goFrame.route.GetHttpHandler(method, path[0])
	if !ok && request.Header.Get("Upgrade") == "websocket"{
		wsHandle,ok := goFrame.route.GetWebSocketHandler(path[0])
		if !ok {
			responeWriter.WriteHeader(404);
			responeWriter.Write(utils.StringToBytes("404 Not Found!"));
			return
		}
		wsParams := new(WsParams)
		wsParams.ResponseWriter = responeWriter
		wsParams.Request = request
		if len(path) > 1 {
			keyAndvalue := strings.Split(path[1], "&")
			wsParams.FirstParam = make(map[string]string, len(keyAndvalue))
			for i := 0; i < len(keyAndvalue); i++{
				kv := strings.Split(keyAndvalue[i], "=")
				wsParams.FirstParam[kv[0]] = kv[1]
			}
		}
		wsHandle.StartLink(wsParams)
		return
	}

	param := &Params{ResponseWriter:responeWriter, Request:request}
	if len(path) > 1 {
		keyAndvalue := strings.Split(path[1], "&")
		param.GetParam = make(map[string]string, len(keyAndvalue))
		for i := 0; i < len(keyAndvalue); i++{
			kv := strings.Split(keyAndvalue[i], "=")
			param.GetParam[kv[0]] = kv[1]
		}
	}
	if method == "POST" {
		contentType := request.Header.Get("Content-Type")
		switch contentType {
		case "application/x-www-form-urlencoded":
			request.ParseForm()
			param.PostParams = make(map[string][]string, len(request.PostForm))
			for key,value := range request.PostForm{
				param.PostParams[key] = value
			}
			break;
		case "application/x-www-form-urlencoded-1t1":
			request.ParseForm()
			param.PostParam = make(map[string]string, len(request.PostForm))
			for key,value := range request.PostForm{
				param.PostParam[key] = value[0]
			}
			break;
		}
	}
	handle.ExecuteHandler(param)
}

//go4web框架创建
func CreateApp() (*GoFrame){
	route := new(Route)
	route.httpRoute = map[string]*HttpHandler{}
	route.websocketRoute = map[string]*WsHandler{}
	
	goFrame := &GoFrame{route:route}

	//当favicon.ico为空时的预防
	goFrame.AddHttpUrl("GET","/favicon.ico",CreateHttpHandler(func(params *Params) bool {
		params.ResponseWriter.WriteHeader(200)
		return true
	}))
	return &GoFrame{route:route}
}

//启动框架运行
func (goFrame *GoFrame) StartFrame(port string){
	http.Handle("/", goFrame);
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Print("Fail to start the go4web Frame!")
	}
}

//添加http路由
func (goFrame *GoFrame) AddHttpUrl(method,path string, handle *HttpHandler) (bool){
	realPath := goFrame.route.JointPath(method,path)
	_,exist:=goFrame.route.httpRoute[realPath]
	if exist {
		return false
	}
	goFrame.route.httpRoute[realPath] = handle
	return true
}

//链接请求类型与请求url路径用
func (*Route) JointPath(method, path string) (string){
	//fmt.Print("\n"+method+":"+path+"\n")
	return method+":"+path
}

//分发路由后用于获得对应url的处理函数
func (route *Route) GetHttpHandler(method string, path string) (handler *HttpHandler, ok bool){
	handler, ok = route.httpRoute[route.JointPath(method,path)]
	return
}

//分发路由后用于获得对应websocket的处理函数
func (route *Route) GetWebSocketHandler(path string) (handler *WsHandler, ok bool){
	handler, ok = route.websocketRoute[path]
	return
}

/*--------------------websocket具体实现------------*/

//添加websocket路由
func (goFrame *GoFrame) AddWsUrl(path string, handler *WsHandler ) (bool){
	_,have := goFrame.route.websocketRoute[path]
	if have {
		return false
	}
	goFrame.route.websocketRoute[path] = handler
	return true
}

//创建一个对应websocket请求的handler，用于新建路由时使用
func CreateWsHandler(OnOpen func(context *WebSocketContext) bool, OnMessage func(context *WebSocketContext, message string), OnClose, OnError func(context *WebSocketContext)) (*WsHandler){
	handler := new(WsHandler)
	handler.OnOpen = OnOpen
	handler.OnMessage = OnMessage
	handler.OnClose = OnClose
	handler.OnError = OnError

	return handler
}

//websocket链接的上下文
type WebSocketContext struct {
	Connection net.Conn
	Buffer *bufio.ReadWriter
	Handler *WsHandler
}

//websocket链接的处理handler
type WsHandler struct{
	OnOpen func(context *WebSocketContext) bool
	OnMessage func(context *WebSocketContext, message string)
	OnClose func(context *WebSocketContext)
	OnError func(context *WebSocketContext)
}

//启动一个websocket链接
func (wsHandler *WsHandler) StartLink(params *WsParams){
	writer := params.ResponseWriter
	key := params.Request.Header.Get("Sec-WebSocket-Key")
	accept := upgradeSecWebSocketKey(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11");
	hj := writer.(http.Hijacker)
	connection, buffer, _ := hj.Hijack()

	webSocketContext := new(WebSocketContext)
	webSocketContext.Connection = connection
	webSocketContext.Buffer = buffer
	webSocketContext.Handler = wsHandler

	upgradeHeader := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + accept + "\r\n\r\n"

	buffer.Write(utils.StringToBytes(upgradeHeader))
	buffer.Flush()

	if !wsHandler.OnOpen(webSocketContext) {
		wsHandler.OnClose(webSocketContext)
		return
	}
	//fmt.Print("即将开始多线程")
	go func() {
		dataBuffer := new(bytes.Buffer)
		for{
			//fmt.Print("进度1\n")
			frameHead := make([]byte, 2)
			_, err := buffer.Read(frameHead)
			if err != nil {
				wsHandler.EndLink(webSocketContext)
				break
			}
			//fmt.Print("进度2\n")
			finRsvOpcode := parseByteToBin(uint8(frameHead[0]))
			if finRsvOpcode[1] || finRsvOpcode[2] || finRsvOpcode[3]{
				wsHandler.EndLink(webSocketContext)
				break
			}
			opcode := parseBinToInt(finRsvOpcode[4:])

			maskPayloadLen := parseByteToBin(uint8(frameHead[1]))
			payloadLen := parseBinToInt(maskPayloadLen[1:])
			//fmt.Print("进度3\n")
			//fmt.Print("当前opcode:",opcode,"\n")
			switch opcode {
			case 0:
			case 1:
				if !maskPayloadLen[0] {
					wsHandler.EndLink(webSocketContext)
					break
				}
				maskingKey := make([]byte, 4)
				buffer.Read(maskingKey)

				payload := make([]byte, payloadLen)
				data := make([]byte, payloadLen)
				buffer.Read(payload)

				for i:=0; i<payloadLen; i++ {
					data[i] = payload[i] ^ maskingKey[i%4]
				}
				dataBuffer.Write(data)
				if opcode==1 {

					wsHandler.OnMessage(webSocketContext, dataBuffer.String())
					dataBuffer.Reset()
				}
				break
			case 8:
				wsHandler.EndLink(webSocketContext)
				break
			default:
				wsHandler.EndLink(webSocketContext)
				break
			}
		}
	}()

}

//关闭一个websocket链接
func (wsHandler *WsHandler) EndLink(webSocketContext *WebSocketContext){
	wsHandler.OnClose(webSocketContext)
	webSocketContext.Connection.Close()
}

//处理http升级为websocket的加密问题
func upgradeSecWebSocketKey(key string) string {
	s := sha1.New()
	s.Write(utils.StringToBytes(key))
	return base64.StdEncoding.EncodeToString(s.Sum(nil))
}

//发送websocket消息
func (wsHandler *WsHandler) SendMessage(webSocketContext *WebSocketContext, message string){
	messageByte := utils.StringToBytes(message)
	length := len(messageByte)
	if length<126 {
		finRsvOpcodeMask := make([]bool, 8);
		initBoolListWithStr(finRsvOpcodeMask,"10000001")

		payloadLen := parseByteToBin(uint8(length));	//不需要payloadLen[0] = false

		frameHeader := make([]byte, 2)
		frameHeader[0] = byte(parseBinToInt(finRsvOpcodeMask))
		frameHeader[1] = byte(parseBinToInt(payloadLen))
		webSocketContext.Buffer.Write(frameHeader)
		webSocketContext.Buffer.Write(messageByte)
		webSocketContext.Buffer.Flush()
	}else if length < 65535 {
		finRsvOpcodeMask := make([]bool, 8);
		initBoolListWithStr(finRsvOpcodeMask,"10000001")

		payloadLen := parseByteToBin(uint8(126))
		frameHeader := make([]byte, 2)
		frameHeader[0] = byte(parseBinToInt(finRsvOpcodeMask))
		frameHeader[1] = byte(parseBinToInt(payloadLen))

		realPayloadLen := make([]byte, 2)
		binary.BigEndian.PutUint16(realPayloadLen, uint16(length))

		webSocketContext.Buffer.Write(frameHeader)
		webSocketContext.Buffer.Write(realPayloadLen)
		webSocketContext.Buffer.Write(messageByte)
		webSocketContext.Buffer.Flush()
	}else if length < 4294967295{
		finRsvOpcodeMask := make([]bool, 8);
		initBoolListWithStr(finRsvOpcodeMask,"10000001")

		payloadLen := parseByteToBin(uint8(127))
		frameHeader := make([]byte, 2)
		frameHeader[0] = byte(parseBinToInt(finRsvOpcodeMask))
		frameHeader[1] = byte(parseBinToInt(payloadLen))

		realPayloadLen := make([]byte, 4)
		binary.BigEndian.PutUint32(realPayloadLen, uint32(length))

		webSocketContext.Buffer.Write(frameHeader)
		webSocketContext.Buffer.Write(realPayloadLen)
		webSocketContext.Buffer.Write(messageByte)
		webSocketContext.Buffer.Flush()
	}else{
		//需要分块发送
	}
}

//转byte数组至代表二进制的bool数组【单字节】
func parseByteToBin(byteCode uint8) []bool{
	binCode := make([]bool, 8)
	//fmt.Print("byteCode:",byteCode,"\n")
	for pos:=0; pos < 8; pos++{
		binCode[pos] = uint8(( byteCode << uint(pos)) >> 7 ) == 1
	}
	//fmt.Print(binCode,"\n")
	return binCode
}

//转byte数组至代表二进制的bool数组【多字节】
func parseBytesToBins(byteCode uint32, byteSize int) []bool{
	binCode := make([]bool, byteSize*8)
	//fmt.Print("byteCode:",byteCode,"\n")
	if byteSize == 1 {
		return parseByteToBin(uint8(byteCode))
	}else if byteSize == 2{
		for pos:=0; pos < 16; pos++{
			binCode[pos] = uint16(( byteCode << uint(pos)) >> 15 ) == 1
		}
	}else if byteSize==4 {
		for pos:=0; pos < 32; pos++{
			binCode[pos] = uint32(( byteCode << uint(pos)) >> 31 ) == 1
		}
	}
	//fmt.Print(binCode,"\n")
	return binCode
}

//转二进制表示的bool数组至Int类型
func parseBinToInt(binCode []bool) int{
	length := len(binCode)
	result := 0
	for i:=0; i<length; i++{
		if(binCode[i]){
			result += int(math.Pow(2, float64(length-1-i)))
		}
	}
	return result
}

//初始化一个bool数组
func initBoolList(boolList []bool){
	length := len(boolList)
	for i:=0; i< length; i++{
		boolList[i] = false
	}
}

//用字符串初始化一个bool数组
func initBoolListWithStr(boolList []bool, target string){
	length := len(boolList)
	if length > len(target) {
		length = len(target)
	}

	for i:=0; i<length ; i++  {
		boolList[i] = target[i] == '1'
	}
}

/*--------------------html具体实现------------*/

//添加静态html页面支持
func (goFrame *GoFrame) AddHtmlUrl(url string, staticPath string){
	//http.Handle("/f/", http.StripPrefix("/f/",http.FileServer(http.Dir("./view")))) 调用:http://127.0.0.1:8088/f/view文件夹下文件
	http.Handle(url, http.StripPrefix(url, http.FileServer(http.Dir(staticPath))))
}

//添加管理的shtml模板
func (goFrame *GoFrame) RegisterShtmlView(templatePattern string){
											//"view/**/*"表示与main.go同级下view文件夹下**文件夹下的*文件
	tpl, err := template.ParseGlob(templatePattern)
	if err != nil {
		fmt.Print(err.Error())
	}
	for _,view := range tpl.Templates(){
		tqlName := view.Name()

		http.HandleFunc(tqlName, func(writer http.ResponseWriter, request *http.Request) {
			tpl.ExecuteTemplate(writer, tqlName, nil)
		})
	}
}

/*--------------------静态文件具体实现------------*/

func (goFrame *GoFrame) AddStaticUrl(url string, staticPath string){
	//http.Handle("/f/", http.StripPrefix("/f/",http.FileServer(http.Dir("./view")))) 调用:http://127.0.0.1:8088/f/view文件夹下文件
	http.Handle(url, http.StripPrefix(url, http.FileServer(http.Dir(staticPath))))
}

