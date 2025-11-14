package main

import (
	"errors"
	"github.com/davecgh/go-spew/spew"
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

// ObjectGet gets a single object
func (s *serviceImpl) ObjectGet(c *gin.Context) {

	ns := c.Param("ns")
	id := c.Param("id")

	// which components from the object are being requested?
	attribs := c.DefaultQuery("attribs", "none")
	components := decodeComponents(attribs)

	// log request info
	if s.cfg.Debug == true {
		log.Printf("INFO: get object request [%s/%s] (attribs %s)", ns, id, attribs)
	}

	// grab the access lock
	accessLock(id)
	defer accessUnlock(id)

	o, err := s.es.ObjectGetByKey(ns, id, components)
	if err != nil {
		if errors.Is(err, uvaeasystore.ErrNotFound) {
			c.String(http.StatusNotFound, uvaeasystore.ErrNotFound.Error())
			return
		}
		c.String(mapEsErrorToHttpError(err), err.Error())
		return
	}

	c.JSON(http.StatusOK, o)
}

// ObjectsGet gets a list of objects
func (s *serviceImpl) ObjectsGet(c *gin.Context) {

	ns := c.Param("ns")

	// which components from the object are being requested?
	attribs := c.DefaultQuery("attribs", "none")
	components := decodeComponents(attribs)

	var req uvaeasystore.GetObjectsRequest
	if jsonErr := c.BindJSON(&req); jsonErr != nil {
		log.Printf("ERROR: Unable to parse request: %s", jsonErr.Error())
		c.String(http.StatusBadRequest, uvaeasystore.ErrDeserialize.Error())
		return
	}

	// log request info
	if s.cfg.Debug == true {
		log.Printf("INFO: get objects request [%s] (attribs %s)", ns, attribs)
		log.Printf("DEBUG: req [%s]", spew.Sdump(req))
	}

	// grab the access lock
	//accessLock(id)
	//defer accessUnlock(id)

	results, err := s.es.ObjectGetByKeys(ns, req.Ids, components)
	if err != nil {
		if errors.Is(err, uvaeasystore.ErrNotFound) {
			c.String(http.StatusNotFound, uvaeasystore.ErrNotFound.Error())
			return
		}
		c.String(mapEsErrorToHttpError(err), err.Error())
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

func (s *serviceImpl) ObjectsSearch(c *gin.Context) {

	ns := c.Param("ns")

	// which components from the object are being requested?
	attribs := c.DefaultQuery("attribs", "none")
	components := decodeComponents(attribs)

	var req uvaeasystore.EasyStoreObjectFields
	if jsonErr := c.BindJSON(&req); jsonErr != nil {
		log.Printf("ERROR: Unable to parse request: %s", jsonErr.Error())
		c.String(http.StatusBadRequest, uvaeasystore.ErrDeserialize.Error())
		return
	}

	// log request info
	if s.cfg.Debug == true {
		log.Printf("INFO: find request [%s] (attribs %s)", ns, attribs)
		log.Printf("DEBUG: req [%s]", spew.Sdump(req))
	}

	results, err := s.es.ObjectGetByFields(ns, req, components)
	if err != nil {
		if errors.Is(err, uvaeasystore.ErrNotFound) {
			c.String(http.StatusNotFound, uvaeasystore.ErrNotFound.Error())
			return
		}
		c.String(mapEsErrorToHttpError(err), err.Error())
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

func (s *serviceImpl) ObjectCreate(c *gin.Context) {

	ns := c.Param("ns")

	req := uvaeasystore.NewEasyStoreObject("", "")
	if jsonErr := c.BindJSON(&req); jsonErr != nil {
		log.Printf("ERROR: Unable to parse request: %s", jsonErr.Error())
		c.String(http.StatusBadRequest, uvaeasystore.ErrDeserialize.Error())
		return
	}

	// validate that the namespace is consistent
	if req.Namespace() != ns {
		log.Printf("ERROR: inconsistent namespaces in request %s/%s", req.Namespace(), ns)
		c.String(http.StatusBadRequest, uvaeasystore.ErrBadParameter.Error())
		return
	}

	// log request info
	if s.cfg.Debug == true {
		log.Printf("INFO: create request [%s/%s]", ns, req.Id())
		log.Printf("DEBUG: req [%s]", spew.Sdump(req))
	}

	// grab the access lock
	accessLock(req.Id())
	defer accessUnlock(req.Id())

	o, err := s.es.ObjectCreate(req)
	if err != nil {
		c.String(mapEsErrorToHttpError(err), err.Error())
		return
	}

	c.JSON(http.StatusCreated, o)
}

func (s *serviceImpl) ObjectUpdate(c *gin.Context) {

	ns := c.Param("ns")
	id := c.Param("id")

	// which components from the object are being requested?
	attribs := c.DefaultQuery("attribs", "none")
	components := decodeComponents(attribs)

	req := uvaeasystore.NewEasyStoreObject("", "")
	if jsonErr := c.BindJSON(&req); jsonErr != nil {
		log.Printf("ERROR: Unable to parse request: %s", jsonErr.Error())
		c.String(http.StatusBadRequest, uvaeasystore.ErrDeserialize.Error())
		return
	}

	// validate that the namespace is consistent
	if req.Namespace() != ns {
		log.Printf("ERROR: inconsistent namespaces in request %s/%s", req.Namespace(), ns)
		c.String(http.StatusBadRequest, uvaeasystore.ErrBadParameter.Error())
		return
	}

	// validate that the id is consistent
	if req.Id() != id {
		log.Printf("ERROR: inconsistent id in request %s/%s", req.Id(), id)
		c.String(http.StatusBadRequest, uvaeasystore.ErrBadParameter.Error())
		return
	}

	// log request info
	if s.cfg.Debug == true {
		log.Printf("INFO: update request [%s/%s] (attribs %s)", ns, id, attribs)
		log.Printf("DEBUG: req [%s]", spew.Sdump(req))
	}

	// grab the access lock
	accessLock(req.Id())
	defer accessUnlock(req.Id())

	o, err := s.es.ObjectUpdate(req, components)
	if err != nil {
		c.String(mapEsErrorToHttpError(err), err.Error())
		return
	}

	c.JSON(http.StatusOK, o)
}

// ObjectDelete deletes a single object
func (s *serviceImpl) ObjectDelete(c *gin.Context) {

	ns := c.Param("ns")
	id := c.Param("id")

	// need to include the vtag
	vtag := c.DefaultQuery("vtag", "unknown")

	// which components from the object are being requested?
	attribs := c.DefaultQuery("attribs", "none")
	components := decodeComponents(attribs)

	// log request info
	if s.cfg.Debug == true {
		log.Printf("INFO: delete request [%s/%s] (attribs %s)", ns, id, attribs)
	}

	// grab the access lock
	accessLock(id)
	defer accessUnlock(id)

	obj := uvaeasystore.ProxyEasyStoreObject(ns, id, vtag)
	_, err := s.es.ObjectDelete(obj, components)
	if err != nil {
		if errors.Is(err, uvaeasystore.ErrNotFound) {
			c.String(http.StatusNotFound, uvaeasystore.ErrNotFound.Error())
			return
		}
		c.String(mapEsErrorToHttpError(err), err.Error())
		return
	}

	// standard delete response
	r := emptyStruct{}
	c.JSON(http.StatusNoContent, r)
}

func (s *serviceImpl) FileGet(c *gin.Context) {
	c.String(http.StatusNotImplemented, "not implemented yet")
	return
}

func (s *serviceImpl) FileCreate(c *gin.Context) {

	ns := c.Param("ns")
	id := c.Param("id")

	req := uvaeasystore.NewEasyStoreBlob("", "", nil)
	if jsonErr := c.BindJSON(&req); jsonErr != nil {
		log.Printf("ERROR: Unable to parse request: %s", jsonErr.Error())
		c.String(http.StatusBadRequest, uvaeasystore.ErrDeserialize.Error())
		return
	}

	err := s.es.FileCreate(ns, id, req)
	if err != nil {
		c.String(mapEsErrorToHttpError(err), err.Error())
		return
	}

	// standard delete response
	r := emptyStruct{}
	c.JSON(http.StatusNoContent, r)
}

func (s *serviceImpl) FileUpdate(c *gin.Context) {

	ns := c.Param("ns")
	id := c.Param("id")

	req := uvaeasystore.NewEasyStoreBlob("", "", nil)
	if jsonErr := c.BindJSON(&req); jsonErr != nil {
		log.Printf("ERROR: Unable to parse request: %s", jsonErr.Error())
		c.String(http.StatusBadRequest, uvaeasystore.ErrDeserialize.Error())
		return
	}

	err := s.es.FileUpdate(ns, id, req)
	if err != nil {
		c.String(mapEsErrorToHttpError(err), err.Error())
		return
	}

	// standard delete response
	r := emptyStruct{}
	c.JSON(http.StatusNoContent, r)
}

func (s *serviceImpl) FileRename(c *gin.Context) {

	ns := c.Param("ns")
	id := c.Param("id")
	name := c.Param("name")
	newName := c.Query("new")

	err := s.es.FileRename(ns, id, name, newName)
	if err != nil {
		c.String(mapEsErrorToHttpError(err), err.Error())
		return
	}

	// standard delete response
	r := emptyStruct{}
	c.JSON(http.StatusNoContent, r)
}

func (s *serviceImpl) FileDelete(c *gin.Context) {
	ns := c.Param("ns")
	id := c.Param("id")
	name := c.Param("name")

	err := s.es.FileDelete(ns, id, name)
	if err != nil {
		c.String(mapEsErrorToHttpError(err), err.Error())
		return
	}

	// standard delete response
	r := emptyStruct{}
	c.JSON(http.StatusNoContent, r)
}

func mapEsErrorToHttpError(err error) int {

	if strings.Contains(err.Error(), uvaeasystore.ErrBadParameter.Error()) {
		log.Printf("ERROR: %s", err.Error())
		return http.StatusBadRequest
	}
	if strings.Contains(err.Error(), uvaeasystore.ErrFileNotFound.Error()) {
		log.Printf("WARNING: %s", err.Error())
		return http.StatusNotFound
	}
	if strings.Contains(err.Error(), uvaeasystore.ErrNotFound.Error()) {
		log.Printf("WARNING: %s", err.Error())
		return http.StatusNotFound
	}
	if strings.Contains(err.Error(), uvaeasystore.ErrStaleObject.Error()) {
		log.Printf("WARNING: %s", err.Error())
		return http.StatusConflict
	}
	if strings.Contains(err.Error(), uvaeasystore.ErrAlreadyExists.Error()) {
		log.Printf("WARNING: %s", err.Error())
		return http.StatusConflict
	}

	log.Printf("ERROR: %s", err.Error())
	return http.StatusInternalServerError
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
