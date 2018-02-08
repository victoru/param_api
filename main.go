package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const DEFAULT_ENVIRONMENT = "dev"

var (
	currentEnvironment = DEFAULT_ENVIRONMENT
	DebugMode          = false
	region             = os.Getenv("AWS_REGION")
	debug              = os.Getenv("DEBUG")
	cliVersion         = "0.0.2"
)

type Response struct {
	status  int
	Message string
	Data    interface{}
}

type paramRequest map[string]string

func (p paramRequest) valid() bool {
	fmt.Println(p)
	return true
}

func main() {
	if env, ok := os.LookupEnv("ENVIRONMENT"); ok {
		currentEnvironment = env
	}
	api()
}

func api() {
	router := mux.NewRouter().StrictSlash(true)
	registerHandlers(router)
	loggedRouter := handlers.LoggingHandler(os.Stdout, router)
	log.Println("Validating Config") //todo, validate config
	if region == "" {
		log.Fatal("Environment variable AWS_REGION undefined")
		//todo, check against list of known regions
	}
	//in debug mode no caching takes place
	//logs are produced in greater detail
	if debug != "" {
		log.Printf("DEBUG flag set to %+v - attempting to parse to boolean", debug)
		debugenabled, err := strconv.ParseBool(debug)
		if err != nil {
			log.Printf("Warning: Could not parse debug flag, value provided was %s\n %s", DebugMode, err.Error())
			log.Println("debug mode: false")
			DebugMode = false
		} else {
			DebugMode = debugenabled
			log.Printf("debug mode set to %+v", DebugMode)
		}
	}
	log.Println("Started: Ready to serve")
	log.Fatal(http.ListenAndServe(":8080", loggedRouter)) //todo, refactor to make port dynamic
}

func registerHandlers(r *mux.Router) {
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	r.HandleFunc("/params", envHandler).Methods("POST")
}

func parseParamRequestBody(b io.ReadCloser) paramRequest {
	decoder := json.NewDecoder(b)
	var p paramRequest
	err := decoder.Decode(&p)
	if err != nil {
		log.Printf("encountered issue decoding request body; %s", err.Error())
		return paramRequest{}
	}
	return p
}

func (p paramRequest) getData() map[string]string {
	c := ssmClient{NewClient(region)}
	var params []string

	for _, v := range p {
		params = append(params, fmt.Sprintf("/%s/%s", currentEnvironment, v))
	}
	o, err := c.ParamList(params...)
	if err != nil {
		log.Fatal(err)
	}

	// param keys cant be used more than once this way...
	data := map[string]string{}
	for k, v := range p {
		for _, par := range o.Parameters {
			if fmt.Sprintf("/%s/%s", currentEnvironment, v) == *par.Name {
				// base64 encode
				data[k] = base64.StdEncoding.EncodeToString([]byte(*par.Value))
			}
		}
	}
	return data
}

func envHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("1")
	p := parseParamRequestBody(r.Body)
	if !p.valid() {
		badRequest(w, p)
		return
	}
	data := p.getData()
	resp := Response{status: http.StatusOK, Data: data} //todo, check length of list before returning
	JSONResponseHandler(w, resp)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var m = make(map[string]string)
	m["error"] = fmt.Sprintf("Route %s not found with method %s, please check request and try again",
		r.URL.Path, r.Method)
	resp := Response{Data: m, status: http.StatusNotFound}
	JSONResponseHandler(w, resp)
}

func badRequest(w http.ResponseWriter, p paramRequest) {
	w.Header().Set("Content-Type", "application/json")
	var m = make(map[string]string)
	m["error"] = fmt.Sprintf("bad request: %s", p)
	resp := Response{Data: m, status: http.StatusBadRequest}
	JSONResponseHandler(w, resp)
}

func JSONResponseHandler(w http.ResponseWriter, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.status)
	json.NewEncoder(w).Encode(resp.Data)
}
