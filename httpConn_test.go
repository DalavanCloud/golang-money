// Copyright 2016 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// End Copyright

package money

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	//"strings"
	"testing"
	"time"
)

func setupMoneyArray() (*Money, *Money) {
	m1 := new(Money)
	m1.spanId = 23456
	m1.traceId = "money for nothing"
	m1.parentId = 6748
	m1.spanName = "spanner"
	m1.startTime, _ = time.Parse(time.RFC3339Nano, "0")
	m1.spanDuration = 5
	m1.errorCode = 202
	m1.spanSuccess = true

	m2 := new(Money)
	m2.spanId = 98765
	m2.traceId = "money for something"
	m2.parentId = 23456
	m2.spanName = "namest"
	m2.startTime, _ = time.Parse(time.RFC3339Nano, "10")
	m2.spanDuration = 7
	m2.errorCode = 257
	m2.spanSuccess = true

	return m1, m2
}

func TestDelAllMNYHeaders(t *testing.T) {
	m1, m2 := setupMoneyArray()
	var mnys []*Money
	mnys = append(mnys, m1)
	mnys = append(mnys, m2)

	server := httptest.NewServer(
		http.HandlerFunc(
			func(rw http.ResponseWriter, req *http.Request) {
				for x := range mnys {
					rw.Header().Add(HEADER, mnys[x].ToString())
				}
				DelAllMNYHeaders(rw)
			},
		),
	)
	defer server.Close()

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	httpClient := &http.Client{Transport: transport}
	newReq, e := http.NewRequest("GET", server.URL, nil)
	if e != nil {
		log.Error(e)
	}
	resp, e := httpClient.Do(newReq)
	if e != nil {
		log.Error(e)
	}
	defer resp.Body.Close()

	count := 0
	for k, v := range resp.Header {
		if k == HEADER {
			for i := 0; i < len(v); i++ {
				for _, m := range mnys {
					if m.ToString() == v[i] {
						count++
						break
					}
				}
			}
		}
	}

	if count != 0 {
		t.Errorf("Expected Money header count not correct, Got: %v, Expected: 0", count)
	}
}

func TestAddAllMNYHeaders(t *testing.T) {
	m1, m2 := setupMoneyArray()
	var mnys []*Money
	mnys = append(mnys, m1)
	mnys = append(mnys, m2)

	server := httptest.NewServer(
		http.HandlerFunc(
			func(rw http.ResponseWriter, req *http.Request) {
				AddAllMNYHeaders(rw, mnys)
			},
		),
	)
	defer server.Close()

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	httpClient := &http.Client{Transport: transport}
	newReq, e := http.NewRequest("GET", server.URL, nil)
	if e != nil {
		log.Error(e)
	}
	resp, e := httpClient.Do(newReq)
	if e != nil {
		log.Error(e)
	}
	defer resp.Body.Close()

	count := 0
	for k, v := range resp.Header {
		if k == HEADER {
			for i := 0; i < len(v); i++ {
				for _, m := range mnys {
					if m.ToString() == v[i] {
						count++
						break
					}
				}
			}
		}
	}

	if count != len(mnys) {
		t.Errorf("Expected Money header count not correct, Got: %v, Expected: %v", count, len(mnys))
	}
}

func TestMyGrandParentIs(t *testing.T) {
	m1, m2 := setupMoneyArray()
	var mnys []Money
	mnys = append(mnys, *m1)
	mnys = append(mnys, *m2)
	var id int64 = 98765
	var expectedId int64 = 23456

	mnys, mnyId, found := myGrandParentIs(mnys, id)

	if !found {
		t.Error("Descendent not found")
	}
	if mnyId != expectedId {
		t.Error("Wrong descendent found")
	}
	if len(mnys) > 1 {
		t.Error("Descendent not removed")
	}

	for _, m := range mnys {
		if m.spanId == id {
			t.Error("Current id was not removed correctly")
		}
	}

}

func TestCopyOfMNY(t *testing.T) {
	m1, m2 := setupMoneyArray()
	var mnys []*Money
	mnys = append(mnys, m1)
	mnys = append(mnys, m2)
	newMnys := copyOfMNY(mnys)

	if reflect.TypeOf(newMnys) == reflect.TypeOf(mnys) {
		t.Errorf("copy []Money is not a copy")
	}

	for _, m := range mnys {
		found := false
		for _, n := range newMnys {
			if m.spanId == n.spanId {
				found = true
				break
			}
		}

		if !found {
			t.Error("Money values are not the same")
			break
		}
	}
}

func TestGetCurrentHeader(t *testing.T) {
	m1, m2 := setupMoneyArray()
	var mnys []*Money
	mnys = append(mnys, m1)
	mnys = append(mnys, m2)
	var expectedId int64 = 98765

	mny := GetCurrentHeader(mnys)

	if mny.spanId != expectedId {
		t.Errorf("Incorrect header found.  expected: %v, got: %v\n", expectedId, mny.spanId)
	}
}

/*
func TestStart(t *testing.T) {
	m1, m2 := setupMoneyArray()
	var mnys []*Money
	mnys = append(mnys, m1)
	mnys = append(mnys, m2)

	spanName := "TEST_SPAN_NAME"
	server := httptest.NewServer(
		http.HandlerFunc(
			func(rw http.ResponseWriter, req *http.Request) {
				Start(rw, req, spanName)
			},
		),
	)
	defer server.Close()

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	httpClient := &http.Client{Transport: transport}
	newReq, e := http.NewRequest("GET", server.URL, nil)
	if e != nil {
		log.Error(e)
	}
	resp, e := httpClient.Do(newReq)
	if e != nil {
		log.Error(e)
	}
	defer resp.Body.Close()
	
	count := 0
	foundNewSpan := false
	for k, v := range resp.Header {
		if k == HEADER {
			for i := 0; i < len(v); i++ {
				for _, m := range mnys {
					if m.ToString() == v[i] {
						count++
						break
					}
				}
				
				if strings.Contains(v[i], spanName) {
					foundNewSpan = true
				}

			}
		}
	}
	
	if count != (len(mnys) + 1) {
		t.Errorf("Expected Money header count not correct, Got: %v, Expected: %v", count, (len(mnys)+1))
	}
	
	if !foundNewSpan {
		t.Errorf("New Money span was not found.  Expected to find Money span with span name: %v", spanName)
	}
}
*/