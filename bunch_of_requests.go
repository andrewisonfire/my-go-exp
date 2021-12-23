package main

import (
	"io/ioutil"
	"log"
	"net/http"
)

func runCode(){
	url := "https://httpbin.org/ip"

	responseChannel := make(chan responseData)
	semaphoreChannel := make(chan struct{}, semaphoreChannelLength)
	requestsNumber := 1000
	for i := 0; i < requestsNumber; i++ {
		go asyncGetData(
			requestsData{
				Url: url,
			},
			responseChannel,
			semaphoreChannel,
		)
	}
	defer func() {
		close(responseChannel)
		close(semaphoreChannel)
	}()
	println("end requests")
	var results []responseData
	for {
		res := <-responseChannel
		results = append(results, res)
		if len(results) == requestsNumber {
			break
		}
	}
	println("end save response")
}

func asyncGetData(requestsData requestsData, responseChannel chan responseData, semaphoreChannel chan struct{}) {
	semaphoreChannel <- struct{}{}
	response, err := http.Get(requestsData.Url)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalln(err)
	}
	responseChannel <- responseData{
		Body: string(body),
	}
	<-semaphoreChannel
}

type requestsData struct {
	Url string
}

type responseData struct {
	Body string
}

const semaphoreChannelLength = 10
