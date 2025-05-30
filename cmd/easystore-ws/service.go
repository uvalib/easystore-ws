package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/uvalib/easystore/uvaeasystore"
	"log"
	"net/http"
	"strings"
)

// this is our service implementation
type serviceImpl struct {
	es  uvaeasystore.EasyStore
	cfg *ServiceConfig
}

func NewService(cfg *ServiceConfig) *serviceImpl {

	es, err := uvaeasystore.NewEasyStore(cfg.esCfg)
	if err != nil {
		log.Fatalf("create easystore failed: %s", err.Error())
	}

	return &serviceImpl{es: es, cfg: cfg}
}

// IgnoreFavicon is a dummy to handle browser favicon requests without warnings
func (s *serviceImpl) IgnoreFavicon(c *gin.Context) {
}

// GetVersion reports the version of the service
func (s *serviceImpl) GetVersion(c *gin.Context) {

	vMap := make(map[string]string)
	vMap["build"] = Version()
	c.JSON(http.StatusOK, vMap)
}

// HealthCheck reports the health of the service
func (s *serviceImpl) HealthCheck(c *gin.Context) {

	type hcResp struct {
		Healthy bool   `json:"healthy"`
		Message string `json:"message,omitempty"`
	}

	msg := ""
	err := s.es.Check()
	if err != nil {
		msg = err.Error()
	}
	hcMap := make(map[string]hcResp)
	hcMap["easystore"] = hcResp{
		Healthy: err == nil,
		Message: msg,
	}

	c.JSON(http.StatusOK, hcMap)
}

// GetObject gets a single object
func (s *serviceImpl) GetObject(c *gin.Context) {

	ns := c.Param("ns")
	id := c.Param("id")

	// which components are being requested?
	attribs := c.DefaultQuery("attribs", "none")
	components := decodeComponents(attribs)

	log.Printf("INFO: request [%s/%s]", ns, id)

	o, err := s.es.GetByKey(ns, id, components)
	if err != nil {
		if errors.Is(err, uvaeasystore.ErrNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		log.Printf("ERROR: %s", err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, o)
}

// GetObjects gets a list of objects
func (s *serviceImpl) GetObjects(c *gin.Context) {

	ns := c.Param("ns")

	var req getObjectsRequest
	if jsonErr := c.BindJSON(&req); jsonErr != nil {
		log.Printf("ERROR: Unable to parse request: %s", jsonErr.Error())
		err := requestError{Message: "Request is malformed or unsupported", Details: jsonErr.Error()}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	log.Printf("INFO: request [%s/%s]", ns, strings.Join(req.Ids, ","))

	results, err := s.es.GetByKeys(ns, req.Ids, uvaeasystore.AllComponents)
	if err != nil {
		if errors.Is(err, uvaeasystore.ErrNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		log.Printf("ERROR: %s", err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var resp getObjectsResponse
	resp.Results = make([]uvaeasystore.EasyStoreObject, 0)

	// process results as appropriate
	if results.Count() != 0 {
		total := results.Count()
		log.Printf("INFO: located %d object(s)...", total)
		var obj uvaeasystore.EasyStoreObject
		obj, err = results.Next()
		for err == nil {
			resp.Results = append(resp.Results, obj)
			obj, err = results.Next()
		}
	}

	c.JSON(http.StatusOK, resp)
}

func (s *serviceImpl) SearchObjects(c *gin.Context) {

	ns := c.Param("ns")

	var req uvaeasystore.EasyStoreObjectFields
	if jsonErr := c.BindJSON(&req); jsonErr != nil {
		log.Printf("ERROR: Unable to parse request: %s", jsonErr.Error())
		err := requestError{Message: "Request is malformed or unsupported", Details: jsonErr.Error()}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	//log.Printf("INFO: request [%s/%s]", ns, strings.Join(req.Ids, ","))

	results, err := s.es.GetByFields(ns, req, uvaeasystore.AllComponents)
	if err != nil {
		if errors.Is(err, uvaeasystore.ErrNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		log.Printf("ERROR: %s", err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var resp searchObjectsResponse
	resp.Results = make([]uvaeasystore.EasyStoreObject, 0)

	// process results as appropriate
	if results.Count() != 0 {
		total := results.Count()
		log.Printf("INFO: located %d object(s)...", total)
		var obj uvaeasystore.EasyStoreObject
		obj, err = results.Next()
		for err == nil {
			resp.Results = append(resp.Results, obj)
			obj, err = results.Next()
		}
	}

	c.JSON(http.StatusOK, resp)
}

func (s *serviceImpl) CreateObject(c *gin.Context) {

	ns := c.Param("ns")

	var req uvaeasystore.EasyStoreObject
	if jsonErr := c.BindJSON(&req); jsonErr != nil {
		log.Printf("ERROR: Unable to parse request: %s", jsonErr.Error())
		err := requestError{Message: "Request is malformed or unsupported", Details: jsonErr.Error()}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	// validate that the namespace is consistent
	if req.Namespace() != ns {
		log.Printf("ERROR: inconsistent namespaces in request %s/%s", req.Namespace(), ns)
		err := requestError{Message: "Inconsistent namespaces", Details: fmt.Sprintf("%s/%s", req.Namespace(), ns)}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	o, err := s.es.Create(req)
	if err != nil {
		log.Printf("ERROR: %s", err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// cleanup the return object
	o.SetFiles(nil)
	o.SetMetadata(nil)
	o.SetFields(uvaeasystore.DefaultEasyStoreFields())

	c.JSON(http.StatusCreated, o)
}

func (s *serviceImpl) UpdateObject(c *gin.Context) {

	ns := c.Param("ns")
	id := c.Param("id")

	var req uvaeasystore.EasyStoreObject
	if jsonErr := c.BindJSON(&req); jsonErr != nil {
		log.Printf("ERROR: Unable to parse request: %s", jsonErr.Error())
		err := requestError{Message: "Request is malformed or unsupported", Details: jsonErr.Error()}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	// validate that the namespace is consistent
	if req.Namespace() != ns {
		log.Printf("ERROR: inconsistent namespaces in request %s/%s", req.Namespace(), ns)
		err := requestError{Message: "Inconsistent namespaces", Details: fmt.Sprintf("%s/%s", req.Namespace(), ns)}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	// validate that the id is consistent
	if req.Namespace() != ns {
		log.Printf("ERROR: inconsistent id in request %s/%s", req.Id(), id)
		err := requestError{Message: "Inconsistent id", Details: fmt.Sprintf("%s/%s", req.Id(), id)}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	// FIXME
	o, err := s.es.Update(req, uvaeasystore.AllComponents)
	if err != nil {
		log.Printf("ERROR: %s", err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// cleanup the return object
	o.SetFiles(nil)
	o.SetMetadata(nil)
	o.SetFields(uvaeasystore.DefaultEasyStoreFields())

	c.JSON(http.StatusOK, o)
}

// DeleteObject deletes a single object
func (s *serviceImpl) DeleteObject(c *gin.Context) {

	ns := c.Param("ns")
	id := c.Param("id")

	// need to include the vtag
	vtag := c.DefaultQuery("vtag", "unknown")

	o := uvaeasystore.ProxyEasyStoreObject(ns, id, vtag)
	_, err := s.es.Delete(o, uvaeasystore.AllComponents)
	if err != nil {
		if errors.Is(err, uvaeasystore.ErrNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		log.Printf("ERROR: %s", err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// standard delete response
	r := emptyStruct{}
	c.JSON(http.StatusNoContent, r)
}

func decodeComponents(attribs string) uvaeasystore.EasyStoreComponents {
	// short circuit special case
	if attribs == "all" {
		return uvaeasystore.AllComponents
	}

	// the default, no components requested
	components := uvaeasystore.BaseComponent

	if strings.Contains(attribs, "fields") {
		components += uvaeasystore.Fields
	}
	if strings.Contains(attribs, "files") {
		components += uvaeasystore.Files
	}
	if strings.Contains(attribs, "metadata") {
		components += uvaeasystore.Metadata
	}
	return components
}

//
// end of file
//
