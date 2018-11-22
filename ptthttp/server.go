// Copyright 2018 The go-pttai Authors
// This file is part of the go-pttai library.
//
// The go-pttai library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-pttai library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-pttai library. If not, see <http://www.gnu.org/licenses/>.

package ptthttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/node"
	"github.com/ailabstw/go-pttai/rpc"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/gorilla/mux"
)

type Server struct {
	dir       string
	addr      string
	rpcServer *rpc.Server
	rpcClient *rpc.Client
	srv       *http.Server
}

type MyDir http.Dir

func pathNameToFullName(dir string, name string) (string, error) {
	if filepath.Separator != '/' && strings.ContainsRune(name, filepath.Separator) {
		return "", errors.New("http: invalid character in file path")
	}
	if dir == "" {
		dir = "."
	}
	fullName := filepath.Join(dir, filepath.FromSlash(path.Clean("/"+name)))

	return fullName, nil
}

func (d MyDir) Open(name string) (http.File, error) {
	dir := string(d)
	fullName, err := pathNameToFullName(dir, name)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(fullName)
	if err != nil {
		fullName = filepath.Join(dir, "index.html")
		f, err = os.Open(fullName)
		if err != nil {
			return nil, err
		}
	}
	return f, nil
}

func NewServer(dir string, addr string, rpcPort int, extAddr string, extPort int, node *node.Node) (*Server, error) {
	rpcPortStr := strconv.Itoa(extPort)
	extHTTPAddr = []byte(extAddr)
	extRPCPort = []byte("localhost:" + rpcPortStr)

	rpcServer, err := node.RPCHandler()
	if err != nil {
		return nil, err
	}
	client := rpc.DialInProc(rpcServer)

	srv := &http.Server{Addr: addr}

	s := &Server{
		dir:       dir,
		addr:      addr,
		rpcServer: rpcServer,
		srv:       srv,
		rpcClient: client,
	}

	fs := http.FileServer(MyDir(s.dir))
	r := mux.NewRouter()
	r.HandleFunc("/api/upload/{boardID}", s.uploadHandler).
		Methods("POST")
	r.HandleFunc("/api/upload/{boardID}", s.optionHandler).
		Methods("OPTIONS")
	r.HandleFunc("/api/uploadfile/{boardID}", s.uploadFileHandler).
		Methods("POST")
	r.HandleFunc("/api/uploadfile/{boardID}", s.optionHandler).
		Methods("OPTIONS")
	r.HandleFunc("/api/img/{boardID}/{imgID}", s.imgHandler).
		Methods("GET")
	r.HandleFunc("/api/img/{boardID}/{imgID}", s.optionHandler).
		Methods("OPTIONS")
	r.HandleFunc("/api/file/{boardID}/{mediaID}", s.fileHandler).
		Methods("GET")
	r.HandleFunc("/api/file/{boardID}/{mediaID}", s.optionHandler).
		Methods("OPTIONS")
	r.HandleFunc("/static/js/{path:main.*js}", func(w http.ResponseWriter, r *http.Request) {
		s.jsHandler(w, r, dir)
	}).Methods("Get")
	r.Handle("/{path:.*}", fs)
	srv.Handler = r

	return s, nil
}

func (s *Server) Start() {
	//http.Handle("/api", r)

	go s.srv.ListenAndServe()
}

func (s *Server) Stop() {
	log.Debug("Stop: start")
	s.rpcClient.Close()
	//s.srv.Shutdown(nil)
}

func (s *Server) SetRPCServer(n *node.Node) error {
	rpcServer, err := n.RPCHandler()
	if err != nil {
		return err
	}
	client := rpc.DialInProc(rpcServer)

	s.rpcServer = rpcServer
	s.rpcClient = client

	return nil
}

func (s *Server) optionHandler(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	log.Debug("optionHandler: start", "origin", origin, "method", r.Method)

	w.Header().Set("Accept", "*")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "OPTIONS,GET,POST")
	w.Header().Set("Access-Control-Allow-Headers", "X-CSRFToken")
}

func (s *Server) uploadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	boardIDStr := vars["boardID"]

	fileBytes, filetype, err := s.uploadPreprocess(w, r)
	if err != nil {
		s.renderError(w, "INVALID_FILE", http.StatusBadRequest)
		return
	}

	if filetype != "image/jpeg" && filetype != "image/jpg" &&
		filetype != "image/gif" && filetype != "image/png" {
		s.renderError(w, "INVALID_FILE_TYPE", http.StatusBadRequest)
		return
	}

	backendUploadImg := &content.BackendUploadImg{}
	err = s.rpcClient.Call(backendUploadImg, "content_uploadImage", boardIDStr, filetype, fileBytes)

	if err != nil {
		s.renderError(w, fmt.Sprintf(`{"success": false, "errorMsg": "%v"}`, err), http.StatusBadRequest)
		return
	}

	result := struct {
		Result interface{} `json:"result"`
	}{Result: backendUploadImg}
	resultBytes, err := json.Marshal(result)
	if err != nil {
		s.renderError(w, "UNABLE_TO_MARSHAL", http.StatusBadRequest)
		return
	}

	w.Write(resultBytes)
}

func (s *Server) uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	boardIDStr := vars["boardID"]
	filename := r.FormValue("filename")

	fileBytes, _, err := s.uploadPreprocess(w, r)
	if err != nil {
		s.renderError(w, "INVALID_FILE", http.StatusBadRequest)
		return
	}

	backendUploadFile := &content.BackendUploadFile{}
	err = s.rpcClient.Call(backendUploadFile, "content_uploadFile", boardIDStr, filename, fileBytes)
	if err != nil {
		s.renderError(w, fmt.Sprintf(`{"success": false, "errorMsg": "%v"}`, err), http.StatusBadRequest)
		return
	}

	result := struct {
		Result interface{} `json:"result"`
	}{Result: backendUploadFile}
	resultBytes, err := json.Marshal(result)
	if err != nil {
		s.renderError(w, "UNABLE_TO_MARSHAL", http.StatusBadRequest)
		return
	}

	w.Write(resultBytes)
}

func (s *Server) uploadPreprocess(w http.ResponseWriter, r *http.Request) ([]byte, string, error) {
	origin := r.Header.Get("Origin")

	w.Header().Set("Accept", "*")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "OPTIONS,GET,POST")
	w.Header().Set("Access-Control-Allow-Headers", "X-CSRFToken")
	w.Header().Set("Content-Type", "application/json")

	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)
	if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
		return nil, "", err
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, "", err
	}

	filetype := http.DetectContentType(fileBytes)

	return fileBytes, filetype, nil
}

func (s *Server) imgHandler(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")

	w.Header().Set("Accept", "*")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "OPTIONS,GET,POST")
	w.Header().Set("Access-Control-Allow-Headers", "X-CSRFToken")

	vars := mux.Vars(r)
	boardIDStr := vars["boardID"]
	imgIDStr := vars["imgID"]

	log.Debug("imgHandler: to backend", "boardIDStr", boardIDStr, "imgIDStr", imgIDStr)

	backendGetImg := &content.BackendGetImg{}
	err := s.rpcClient.Call(backendGetImg, "content_getImage", boardIDStr, imgIDStr)
	if err != nil {
		s.renderError(w, "UNABLE_TO_MARSHAL", http.StatusBadRequest)
		return
	}

	switch backendGetImg.Type {
	case pkgservice.MediaTypeJPEG:
		w.Header().Set("Content-Type", "image/jpg")
	case pkgservice.MediaTypePNG:
		w.Header().Set("Content-Type", "image/png")
	case pkgservice.MediaTypeGIF:
		w.Header().Set("Content-Type", "image/gif")
	}
	w.Write(backendGetImg.Buf)
}

func (s *Server) fileHandler(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")

	w.Header().Set("Accept", "*")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "OPTIONS,GET,POST")
	w.Header().Set("Access-Control-Allow-Headers", "X-CSRFToken")

	vars := mux.Vars(r)
	boardIDStr := vars["boardID"]
	mediaIDStr := vars["mediaID"]

	log.Debug("attachHandler: to backend", "boardIDStr", boardIDStr, "mediaIDStr", mediaIDStr)

	backendGetFile := &content.BackendGetFile{}
	err := s.rpcClient.Call(backendGetFile, "content_getFile", boardIDStr, mediaIDStr)
	if err != nil {
		s.renderError(w, "UNABLE_TO_MARSHAL", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(backendGetFile.Buf)
}

func (s *Server) origImgHandler(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")

	w.Header().Set("Accept", "*")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "OPTIONS,GET,POST")
	w.Header().Set("Access-Control-Allow-Headers", "X-CSRFToken")

	vars := mux.Vars(r)
	boardIDStr := vars["boardID"]
	imgIDStr := vars["imgID"]

	log.Debug("origImgHandler: to backend", "boardIDStr", boardIDStr, "imgIDStr", imgIDStr)

	backendGetImg := &content.BackendGetImg{}
	err := s.rpcClient.Call(backendGetImg, "content_getOrigImage", boardIDStr, imgIDStr)
	if err != nil {
		log.Error("origImgHandler: unable to getOrigImage", "e", err)
		s.renderError(w, "UNABLE_TO_MARSHAL", http.StatusBadRequest)
		return
	}

	switch backendGetImg.Type {
	case pkgservice.MediaTypeJPEG:
		w.Header().Set("Content-Type", "image/jpg")
	case pkgservice.MediaTypePNG:
		w.Header().Set("Content-Type", "image/png")
	case pkgservice.MediaTypeGIF:
		w.Header().Set("Content-Type", "image/gif")
	}
	w.Write(backendGetImg.Buf)
}

func (s *Server) jsHandler(w http.ResponseWriter, r *http.Request, dir string) {
	vars := mux.Vars(r)
	path := vars["path"]
	log.Debug("jsHandler: start", "path", path)
	fullName, err := pathNameToFullName(dir, "/static/js/"+path)
	if err != nil {
		s.renderError(w, "UNABLE_LOAD_FILE", http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadFile(fullName)
	if err != nil {
		s.renderError(w, "UNABLE_LOAD_FILE", http.StatusBadRequest)
		return
	}

	newData := reHTTPAddr.ReplaceAll(reRPCPort.ReplaceAll(data, extRPCPort), extHTTPAddr)
	w.Header().Set("Content-Type", "text/javascript")
	w.Write(newData)
}

func (s *Server) renderError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(message))
}
