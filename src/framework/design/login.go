package design

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"framework/design/tools/http_api"
	"html/template"
	"fmt"
)


//登陆
func (s *httpServer) loginHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
	reqParams, err := http_api.NewReqParams(req)

	if err != nil {
		return nil, http_api.Err{400, "INVALID_REQUEST"}
	}

	username, err := reqParams.Get("username")
	if err != nil {
		return nil, http_api.Err{400, "MISSING_ARG_USERNAME"}
	}

	password, err := reqParams.Get("password")

	if err != nil {
		return nil, http_api.Err{400, "MISSING_ARG_PASSWORD"}
	}

	if username != "admin" || password != "123456" {
		//登陆失败, 跳到登陆界面, 给出登陆提示

		http.Redirect(w, req, "/loginForm", http.StatusFound)
	}



	fmt.Println("登陆成功!")
	session, err := s.ctx.design.store.Get(req, "admin")
	if err != nil {
		return nil, http_api.Err{500, "GET_SESSION_ERROR"}
	}
	session.Values["password"] = "123456"

	err = session.Save(req, w)

	if err != nil {
		return nil, http_api.Err{500, "SAVE_SESSION_ERROR"}
	}

	//保存帐号信息成功
	http.Redirect(w, req, "/", http.StatusFound)


	return nil, nil
}

func (s *httpServer) saveAccountSession(w http.ResponseWriter, req *http.Request) (interface{}, error) {
	session, err := s.ctx.design.store.Get(req, "admin")

	if err != nil {
		return false, http_api.Err{500, "CANOT_GET_SESSION"}
	}

	session.Values["password"] = "123456"

	err = session.Save(req, w)

	if err != nil {
		return nil, http_api.Err{500, "SAVE_SESSION_ERROR"}
	}

	return "OK", nil
}

//初始化帐号信息
func (s *httpServer) initAccountInfoHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
	session, err := s.ctx.design.store.Get(req, "admin")
	if err != nil {
		return nil, http_api.Err{500, "GET_SESSION_ERROR"}
	}

	session.Values["password"] = "123456"

	err = session.Save(req, w)

	if err != nil {
		return nil, http_api.Err{500, "SAVE_SESSION_ERROR"}
	}

	return "OK", nil
}


func (s *httpServer) checkLogin(req *http.Request) (interface{}, error) {
	session, err := s.ctx.design.store.Get(req, "admin")

	if err != nil {
		return nil, http_api.Err{500, "CANOT_GET_SESSION"}
	}

	if session.Values["password"] != "123456" {
		return nil, http_api.Err{500, "USER_CANOT_LOGIN"}
	}

	return "OK", nil
}

func (s *httpServer) loginForm(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
	asset, _ := Asset("resources/html/loginForm.html")
	t, _ := template.New("index").Parse(string(asset))

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, nil)

	return nil, nil
}

