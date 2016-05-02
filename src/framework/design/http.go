package design

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"framework/design/tools/http_api"
	"html/template"
	"framework/design/tools/version"
	"path"
	"mime"
	"fmt"
	"net/url"
	"regexp"
	"mime/multipart"
	"log"
	"bytes"
	"hash/crc32"
	"io"
)

type httpServer struct {
	ctx *context
	router http.Handler
}

func newHTTPServer(ctx *context) *httpServer {
	log := http_api.Log(ctx.design.getOpts().Logger)

	router := httprouter.New()
	router.HandleMethodNotAllowed = true
	router.PanicHandler = http_api.LogPanicHandler(ctx.design.getOpts().Logger)
	router.NotFound = http_api.LogNotFoundHandler(ctx.design.getOpts().Logger)
	router.MethodNotAllowed = http_api.LogMethodNotAllowedHandler(ctx.design.getOpts().Logger)
	s := &httpServer{
		ctx: ctx,
		router: router,
	}

	router.Handle("GET", "/ping", http_api.Decorate(s.pingHandler, log, http_api.PlainText))

	router.Handle("GET", "/", http_api.Decorate(s.indexHandler, log))
	router.Handle("GET", "/initAccountInfo", http_api.Decorate(s.initAccountInfoHandler, log))
	router.Handle("GET", "/loginForm", http_api.Decorate(s.loginForm, log))
	router.Handle("GET", "/login", http_api.Decorate(s.loginHandler, log))

	router.Handle("GET", "/static/*filepath", http_api.Decorate(s.staticAssetHandler, log, http_api.PlainText))
	router.Handle("POST", "/upload/go", http_api.Decorate(s.uploadHandler,log))
	return s
}

func (s *httpServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.router.ServeHTTP(w, req)
}

func (s *httpServer) pingHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
	return "OK", nil
}

func (s *httpServer) indexHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
	_, err := s.checkLogin(req)
	if err != nil {
		fmt.Println("用户未登陆")
		//跳到登陆界面
		http.Redirect(w, req, "/loginForm", http.StatusFound)
	}


	fmt.Println("登陆校验成功")

	asset, _ := Asset("resources/html/index.html")
	t, _ := template.New("index").Parse(string(asset))

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, struct{
		Version string
	}{
		Version: version.Binary,
	})

	return nil, nil
}

func (s *httpServer) staticAssetHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
	assetName := ps.ByName("filepath")
	asset, err := Asset(assetName[1:])
	if err != nil {
		return nil, http_api.Err{404, "NOT_FOUND"}
	}

	ext := path.Ext(assetName)
	ct := mime.TypeByExtension(ext)
	if ct == "" {
		switch ext {
		case ".svg":
			ct = "image/svg+xml"
		case ".woff":
			ct = "application/font-woff"
		case ".ttf":
			ct = "application/vnd.ms-fontobject"
		case ".woff2":
			ct = "application/font-woff2"
		}

	}

	if ct != "" {
		w.Header().Set("Content-Type", ct)
	}

	return string(asset), nil
}

const (
	EXCEL_TYPES = "(xls)"
	ACCEPT_FILE_TYPES = EXCEL_TYPES
)

var (
	MIN_FILE_SIZE = 1
	MAX_FILE_SIZE = 999000
	excelTypes = regexp.MustCompile(EXCEL_TYPES)
	acceptFileTypes = regexp.MustCompile(ACCEPT_FILE_TYPES)
)

type FileInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Size string `json:"size"`
	Error string `json:"error, omitempty"`
}

func (fi *FileInfo) ValidateType() (valid bool) {
	if acceptFileTypes.MatchString(fi.Type) {
		return true
	}
	fi.Error = "Filetype not allowed"
	return false
}

func (fi *FileInfo) ValidateSize() (valid bool) {
	if fi.Size < MIN_FILE_SIZE {
		fi.Error = "File is too small"
	} else if fi.Size > MAX_FILE_SIZE {
		fi.Error = "File is too big"
	} else {
		return true
	}
	return false
}

func handleUpload(r *http.Request, p *multipart.Part) (fi *FileInfo) {
	fi = &FileInfo{
		Name: p.FileName(),
		Type: p.Header.Get("Content-Type"),
	}

	if !fi.ValidateType() {
		return
	}

	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			fi.Error = rec.(error).Error()
		}
	}()
	var buffer bytes.Buffer
	hash := crc32.NewIEEE()
	mw := io.MultiWriter(&buffer, hash)
	lr := &io.LimitedReader{R: p, N: MAX_FILE_SIZE + 1}
	_, err := io.Copy(mw, lr)
	if err != nil {
		panic(err)
	}
	fi.Size = MAX_FILE_SIZE + 1 - lr.N
	if !fi.ValidateSize() {
		return
	}

	
}

func (s *httpServer) uploadHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
	fmt.Println("接收到上传的文件.....")
	mr, err := req.MultipartReader()
	if err != nil {
		return nil, http_api.Err{500, "READ_UPLOADFILE_ERROR"}
	}

	req.Form, err = url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		return nil, http_api.Err{500, "PARSE_UPLOADFILE_ERROR"}
	}

	part, err := mr.NextPart()
	for err == nil {
		if name := part.FormName(); name != "" {
			if part.FormName() != "" {
				fmt.Println("fileInfos1: ", name)
				fmt.Println("fileInfos1: ", part.FormName())
				fmt.Println("fileInfos1 filename: ", part.FileName())
				fmt.Println("fileInfos1 type: ", part.Header.Get("Content-Type"))


			} else {
				fmt.Println("fileInfos2: ", name)
			}
		}
		part, err = mr.NextPart()
	}


	if err != nil {
		return nil, http_api.Err{400, "INVALID_REQUEST"}
	}

	return nil, nil
}
