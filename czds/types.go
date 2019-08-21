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
	"fmt"
	"regexp"
	"time"
)

var (
	// AuthURL is the Authentication URL for ICANN CZDS REST API
	AuthURL = "https://account-api.icann.org/api/authenticate"
	// APIURL is the Base API URL for ICANN CZDS REST API
	APIURL = "https://czds-api.icann.org"
	// UserAgent is the default user agent string to use
	UserAgent = "czdstool/0.1 mailto:tools@libertad.link"
	// ErrNoCred is returned when username or password are empty
	ErrNoCred = fmt.Errorf("Missing or invalid credentials supplied")
	// ErrBadRequest is returned by the AUTH API when a malformed request is sent
	ErrBadRequest = fmt.Errorf("Malformed request reported by API")
	// ErrInvalidCredentials is returned by the AUTH API when invalid or incorrect
	// credentials are used
	ErrInvalidCredentials = fmt.Errorf("Invalid credentials provided to API")
	// ErrUnsupportedContent is returned by the AUTH API when a content type other
	// that application/json is sent or requested
	ErrUnsupportedContent = fmt.Errorf("Unsupported content type sent to API")
	// ErrTooManyRequest is returned by the AUTH API when rate limiting is in
	// effect, after about 8 attempts per 5 minute window
	ErrTooManyRequest = fmt.Errorf("Too many authentication requests sent to API")
	// ErrInternalServer is returned by the AUTH API when an internal error on the
	// server side has been triggered.
	ErrInternalServer = fmt.Errorf("Internal error in API")
	// ErrNoPermission is returned when the credentials don't allow accessing the
	// requested TLD file.
	ErrNoPermission = fmt.Errorf("Access to zone file denied to this user")
	// ErrTCRequired is returned when the user accessing the system has not
	// acknowledged the most recent terms and conditions to use the CZDS. The user
	// must login via a browser and accept the ToCs in order to proceed.
	ErrTCRequired = fmt.Errorf("This user has not accepted the new terms and conditions to access the CZDS")
	// zoneRegexp is a regular expression used to obtain the TLD name given an URL
	zoneRegexp = regexp.MustCompile(`/([^/]+)\.zone$`)
)

// Details encodes the information about a zone file as returned by the ICANN
// CZDS REST API
type Details struct {
	Name          string    `json:"name"`
	LastModified  time.Time `json:"last_modified"`
	ContentLength int       `json:"content_length"`
	ContentType   string    `json:"content_type"`
}

// Meta encodes the TLD metadata provided by the ICANN CADS REST API
type Meta struct {
	Name   string `json:"tld"`
	Ulabel string `json:"ulable"`
	Status string `json:"currentStatus"`
	SFTP   bool   `json:"sftp"`
}
