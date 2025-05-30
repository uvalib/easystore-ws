package main

import "github.com/uvalib/easystore/uvaeasystore"

type emptyStruct struct{}

type requestError struct {
	Message string `json:"message"`
	Details string `json:"details"`
}

type getObjectsRequest struct {
	Ids []string `json:"ids"`
}

type getObjectsResponse struct {
	Results []uvaeasystore.EasyStoreObject `json:"results"`
}

type searchObjectsResponse struct {
	Results []uvaeasystore.EasyStoreObject `json:"results"`
}

//
// end of file
//
