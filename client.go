package money

import (
	"log"
	"net/http"
)

type client interface {
	Do(*http.Request) (*http.Response, error)
}

type Options struct {
	Finish bool
}

type moneyClient struct {
	finish bool
	client
	moneyLog       log.Logger
	responseWriter http.ResponseWriter
}

// TODO: implement this. pipeReq pipes the request's headers into the response writer
func pipeReq(rw http.ResponseWriter, resp *http.Request) {
	rw.Header().Set(MoneySpansHeader, resp.Header.Get(MoneySpansHeader))
}

func NewMoneyTransactor(o *Options) moneyClient {
	return moneyClient{finish: o.Finish}
}

func (mc moneyClient) Do(request *http.Request) (*http.Response, error) {
	tracker, err := ExtractTrackerFromRequest(request)
	if err != nil {
		resp, err := mc.client.Do(request)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}

	// set the request's MoneyHeader with a new trace
	request = SetRequestMoneyHeader(tracker, request)
	resp, err := mc.client.Do(request)
	if err != nil {
		return nil, err
	}

	tracker, err = ExtractTrackerFromResponse(resp)
	if err != nil {
		return nil, err
	}

	/*
		// if span does not participate in round trips
		if tracker.CheckOneWay() {
			maps, err := tracker.SpansMap()
			if err != nil {
				return nil, err
			}

			// mc.moneyLog.Log(logging.MessageKey(), mapsToStringResult(maps))
		}

		/*:w
		// if this client is an end node write the tracker maps to response writer
		//
		// this is turned on at options when this client is created
		if tracker.Finisher() {
			maps, err := tracker.SpansMap()
			if err != nil {
				return nil, err
			}

			resp = SetResponseMoneyHeader(maps, resp)
			// TODO: pipe response to response writer and write spans to headers.
		}
	*/

	return resp, nil
}

func (mc moneyClient) Monetize(next client) client {
	return moneyClient{client: next}
}

func (mc moneyClient) Finisher() bool {
	return mc.finish
}