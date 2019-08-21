// Package czds provides a simple abstraction over ICANN CZDS REST API along
// with various utility functions to ease writing clients.
package czds

// Copyright © 2018 Luis E. Muñoz <github@lem.click>
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Sess is the descriptor for an active API session with the ICANN CZDS
type Sess struct {
	AccessToken string `json:"accessToken"`
}

// NewSession returns a virgin Sess
func NewSession() *Sess {
	return &Sess{}
}

// Store serializes and stores the content of the session for later
// reuse. It can be used to authenticate once and reuse the credentials at later
// times. Note that the file contains an unencrypted token. Set permissions
// accordingly.
func (s *Sess) Store(fName string, permission os.FileMode) error {
	if s.AccessToken == "" {
		return fmt.Errorf("cannot store uninitialized Session object")
	}

	b, err := json.Marshal(s)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fName, b, permission)
	return err
}

// Fetch reads the session content from a file produced by an earlier invocation of Store()
func (s *Sess) Fetch(fName string) error {
	buf, err := ioutil.ReadFile(fName)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, s)
}

// String satisfies the String() interface
func (s *Sess) String() string {
	buf, _ := json.Marshal(s)
	return string(buf)
}

// Login attempts to authenticate to the ICANN REST API, obtaining a JWT which
// is stored in the returned Sess object
func (s *Sess) Login(username, password string) error {
	if username == "" || password == "" {
		return ErrNoCred
	}

	cred := map[string]string{
		"username": username,
		"password": password,
	}

	payload, _ := json.Marshal(cred)
	pr := bytes.NewReader(payload)

	resp, err := http.Post(AuthURL, "application/json", pr)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case 200:
		// Do nothing, simply leak out of the switch
	case 400:
		return ErrBadRequest
	case 401:
		return ErrInvalidCredentials
	case 415:
		return ErrUnsupportedContent
	case 429:
		return ErrTooManyRequest
	case 500:
		return ErrInternalServer
	default:
		return fmt.Errorf("Unexpected status code %d returned by AUTH API %s", resp.StatusCode, AuthURL)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, s)
	if err != nil {
		return err
	}

	return nil
}

// TLDFromURL returns the TLD name associated with an URL as returned by List()
func TLDFromURL(u string) string {
	m := zoneRegexp.FindAllStringSubmatch(u, -1)
	if len(m) == 1 && len(m[0]) == 2 {
		return m[0][1]
	}
	return ""
}

// URLFromTLD returns the URL associated with a TLD name
func URLFromTLD(t string) string {
	return fmt.Sprintf("%s/czds/downloads/%s.zone", APIURL, strings.ToLower(t))
}

func (s *Sess) setupRequest(req *http.Request) {
	req.Header.Add("User-Agent", UserAgent)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.AccessToken))
}

// Download makes a request for a TLD zone file using the ICANN CZDS REST API
// and returns a Reader that can be used to read the zone file contents.
func (s *Sess) Download(u string) (*io.ReadCloser, error) {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	s.setupRequest(req)

	cl := new(http.Client)
	resp, err := cl.Do(req)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case 200:
		// Do nothing, simply leak out of the switch
	case 400:
		return nil, ErrBadRequest
	case 403:
		return nil, ErrNoPermission
	case 409:
		return nil, ErrTCRequired
	default:
		return nil, fmt.Errorf("Unexpected status code %d returned by AUTH API %s", resp.StatusCode, u)
	}

	return &resp.Body, nil
}

// GetTLDsMetadata fetches the metadata for all visible TLDs via the CZDS REST
// API
func (s *Sess) GetTLDsMetadata() (map[string]Meta, error) {
	req, err := http.NewRequest("GET", "https://czds-api.icann.org/czds/tlds", nil)
	if err != nil {
		return nil, err
	}

	s.setupRequest(req)

	cl := new(http.Client)
	resp, err := cl.Do(req)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case 200:
		// Do nothing, simply leak out of the switch
	case 400:
		return nil, ErrBadRequest
	case 500:
		return nil, ErrInternalServer
	default:
		return nil, fmt.Errorf("Unexpected status code %d returned by CZDS API", resp.StatusCode)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tldList []Meta

	err = json.Unmarshal(body, &tldList)
	if err != nil {
		return nil, err
	}

	ret := make(map[string]Meta, len(tldList))
	for _, t := range tldList {
		ret[t.Name] = t
	}

	return ret, nil
}

// Details queries the information about a TLD zone file as fetched via the
// ICANN CZDS REST API
func (s *Sess) Details(u string) (*Details, error) {
	req, err := http.NewRequest("HEAD", u, nil)
	if err != nil {
		return nil, err
	}

	s.setupRequest(req)

	cl := new(http.Client)
	resp, err := cl.Do(req)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case 200:
		// Do nothing, simply leak out of the switch
	case 400:
		return nil, ErrBadRequest
	case 500:
		return nil, ErrInternalServer
	default:
		return nil, fmt.Errorf("Unexpected status code %d returned by CZDS API %s", resp.StatusCode, u)
	}

	resp.Body.Close()

	ret := Details{
		Name:        TLDFromURL(u),
		ContentType: resp.Header.Get("Content-Type"),
	}

	if l, err := strconv.Atoi(resp.Header.Get("Content-Length")); err == nil {
		ret.ContentLength = l
	}

	if lm, err := time.Parse("Mon, 02 Jan 2006 15:04:05 MST", resp.Header.Get("Last-Modified")); err == nil {
		ret.LastModified = lm
	}

	return &ret, nil
}

// List provides a list of TLD zone file URLs available for download
func (s *Sess) List() (*[]string, error) {
	theURL := fmt.Sprintf("%s/%s", APIURL, "czds/downloads/links")

	req, err := http.NewRequest("GET", theURL, nil)
	if err != nil {
		return nil, err
	}

	s.setupRequest(req)

	cl := new(http.Client)
	resp, err := cl.Do(req)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case 200:
		// Do nothing, simply leak out of the switch
	case 400:
		return nil, ErrBadRequest
	case 500:
		return nil, ErrInternalServer
	default:
		return nil, fmt.Errorf("Unexpected status code %d returned by CZDS API %s", resp.StatusCode, theURL)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	ret := []string{}

	err = json.Unmarshal(body, &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (d *Details) String() string {
	buf, _ := json.Marshal(d)
	return string(buf)
}

func (m *Meta) String() string {
	return fmt.Sprintf("tld=%s ulabel=%s status=%s sftp=%t",
		m.Name, m.Ulabel, m.Status, m.SFTP)
}
