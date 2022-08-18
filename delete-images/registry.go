package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var REGISTRY_URL = "https://YOUR_REGISTRY_URL_HERE/v2/"

func getRepositories() []interface{} {
	var rep = restGet(REGISTRY_URL+"_catalog", false, true)
	repositories := rep["repositories"].([]interface{})
	return repositories
}

func getTags(repositorie string) []interface{} {
	var tag = restGet(REGISTRY_URL+repositorie+"/tags/list", false, true)
	tags := tag["tags"].([]interface{})
	return tags
}

func getTagData(repositorie string, tag string) time.Time {
	//return jsonSlurper.parseText(restGet(REGISTRY_URL + repositorie + "/manifests/" + tag, false, false)["history"].first()["v1Compatibility"]).created
	var manifest = restGet(REGISTRY_URL+repositorie+"/manifests/"+tag, false, false)

	var v1Compatibility map[string]interface{}

	if manifest["history"] == nil {
		return time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	json.Unmarshal([]byte(manifest["history"].([]interface{})[0].(map[string]interface{})["v1Compatibility"].(string)), &v1Compatibility)

	createdAt, _ := time.Parse(time.RFC3339, v1Compatibility["created"].(string))

	return createdAt
}

func getManifest(repositorie string, tag string, print bool) string {
	digest := restGet(REGISTRY_URL+repositorie+"/manifests/"+tag, true, print)["Docker-Content-Digest"].([]string)[0]
	return digest
}

func getTagSize(repositorie string, tag string) float64 {

	//Busca o arquivo de manifest que contem a lista de layers da imagem
	layers := restGet(REGISTRY_URL+repositorie+"/manifests/"+tag, false, false)["fsLayers"].([]interface{})

	var totalSizeBytes int = 0

	//Agrega o tamanho de cada uma das layers
	for _, element := range layers {
		layerDigest := (element.(map[string]interface{})["blobSum"]).(string)

		url := REGISTRY_URL + repositorie + "/blobs/" + layerDigest

		client := &http.Client{}
		//Request type HEAD para não trazer nada no body, optimizando a consulta conforme documentação: https://docs.docker.com/registry/spec/api/#blob
		req, err := http.NewRequest("HEAD", url, nil)
		perror(err)

		resp, err := client.Do(req)
		perror(err)
		defer resp.Body.Close()

		totalSizeBytes += int(resp.ContentLength)
	}

	megaBytes := (float64(totalSizeBytes) / float64(1024)) / float64(1024)

	return megaBytes
}

func deleteDigest(repositorie string, digest string) string {
	url := REGISTRY_URL + repositorie + "/manifests/" + digest

	fmt.Println("DELETE: " + url)

	client := &http.Client{}
	req, err := http.NewRequest("DELETE", url, nil)
	perror(err)

	resp, err := client.Do(req)
	perror(err)
	defer resp.Body.Close()

	return resp.Status
}

func restGet(url string, withAcceptHeader bool, print bool) map[string]interface{} {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	if print {
		fmt.Println("GET: " + url)
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	perror(err)

	if withAcceptHeader {
		req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	}

	resp, err := client.Do(req)
	perror(err)
	defer resp.Body.Close()

	if withAcceptHeader {
		return map[string]interface{}{"Docker-Content-Digest": resp.Header["Docker-Content-Digest"]}
	}

	jsonDataFromHttp, err := ioutil.ReadAll(resp.Body)
	perror(err)

	var obj map[string]interface{}
	_ = json.Unmarshal([]byte(jsonDataFromHttp), &obj)

	return obj
}

func perror(err error) {
	if err != nil {
		panic(err)
	}
}
