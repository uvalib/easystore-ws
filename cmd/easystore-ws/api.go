package main

import "github.com/uvalib/easystore/uvaeasystore"

type emptyStruct struct{}

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
