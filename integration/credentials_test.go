//
// Copyright 2019 Chef Software, Inc.
// Author: Salim Afiune <afiune@chef.io>
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

package integration

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	DefaultChefServerUser         = "foo"
	DefaultChefServerOrganization = "bar"
)

func createCredentialsConfig() string {
	dir, err := ioutil.TempDir("", "credentials")
	if err != nil {
		panic(err)
	}

	var (
		chefdir    = filepath.Join(dir, ".chef")
		credsFile  = filepath.Join(chefdir, "credentials")
		defaultPem = filepath.Join(chefdir, fmt.Sprintf("%s.pem", DefaultChefServerUser))
		devPem     = filepath.Join(chefdir, "dev.pem")
		emptyPem   = filepath.Join(chefdir, "empty.pem")
	)
	if err := os.MkdirAll(chefdir, os.ModePerm); err != nil {
		panic(err)
	}

	creds := []byte(`[default]
client_name = "` + DefaultChefServerUser + `"
client_key = "` + defaultPem + `"
chef_server_url = "http://localhost/organizations/` + DefaultChefServerOrganization + `"

[dev]
client_name = "dev"
client_key = "` + devPem + `"
chef_server_url = "http://localhost/organizations/dev"

[empty]
client_name = "empty"
client_key = "` + emptyPem + `"
chef_server_url = "http://localhost/organizations/empty"
`)
	err = ioutil.WriteFile(credsFile, creds, 0644)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(defaultPem, key(), 0644)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(devPem, key(), 0644)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(emptyPem, []byte(""), 0644)
	if err != nil {
		panic(err)
	}
	return dir
}

// mocked key
func key() []byte {
	return []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAuip6wlbkvuHvFl4m3Fzpqk3eSCBOudfubLdVZb8bWa0UM++d
fipPJadjDkrf1tlvKhjjoUF1XTizhI1xholJTRm92AjNi+rLWuCcn3Zih58xHTrl
S6+V8VHzmUB9iEDc7iyCdkXabDmoKhvu6CgagrFMN2maBjX7toaTOsAeGahVz3bv
inO/V61ea1an2TCgJ/Ovp77728kmPYQgQAceewelxgx/FwsroPVS0XFlBaRme8bg
Lc2VFJWeID27905ecPXQOZCfm5H2aCfpmzNIwHZ9WRpvOnpQxOXYYsimRGiAhwYA
zkC/FoCgny6ohMjhisibfHI6g7N+Y/iaWBImBwIDAQABAoIBAQCCS+sBi+mrw9wf
zqPqRclxXfC+kIYpQn1ob+SAQwJ0gFQMiZ+0Rw6ALyiAP11tNV+9mg/vtC3doirb
Elgrrni0UtjxlC+wxxOvNlfIsAYEICIy8B6+G1WZwh752w5BSAyZUmO5PejDKJOP
bV+H81GiuU671dhskmnrdUMksoQeteU0Xc1Uc8tYEAp7u/5KEA7WUCF5KiDvp5VC
5YSHC+slrIi0D+2xyWrPMFNqASN05Ke4yXhFEl9tuwiXUFSWahVlyY+mbsfE7xCX
GKPXjXpRFDIg/3EOdL0Cy1U6UUAFkYvYfU08UpQpxNA+ffgUcXbUfbLNcaCHfRl0
L/cPXEmhAoGBAO1/Q6M3sfNtyTMSaRoYxX+k8eV4A2qGy8B1rWv/oNV8NLAdbRBL
9E8QPk4TLSnTf9tK34JqGpUXHktYRvjFasB43V+LGhmrcflQ1YEeGodm3iiBZAuE
7KUX6x71/n3pHiOBea32hDBrGd2MLsX31Nwzo3e152JxLwZifpl48VsxAoGBAMir
cijumQLLIyB1urn1lQqtJm/Cd8jXypwkv6m0qD3bl5/shOEUTqnuGOnXJk2AcfNG
uv2TmP07gjz8JGTqInYgVHZrPd5Cu16TUl6TVezLu0Mu0yjpqlaSar4gTx1UT5ic
annHVa784VQnS/85WzVgldqvPSNsGhlDyfwkf9a3AoGBAJiUQmgByBmUVsaw9UUG
1RuEZMP/rnIp14z2DUxtFm8RNOhQf1kQ8ww4a07Nkx5j+qhwGdg3Qoy2JYhSVoZM
jqDJBa/0Nfh35Ok/vWsOZAzJUcDEH/omk8Ic87kYYT+THQHClOHmllZk+GEVRpd4
+Q/fPQ4Tl2vvOz7m2F7RDH6BAoGAS3oA5FhqANz7B1iAtTUjq/JYhKy2dTqFIJnJ
5UDoDuwraaGCkU4cEFpX0Ix2AayQL5qo9nuvjX/2io2j+rj94URjwG6xxImBBB+R
WbU9GmW+t5RDJB5PTWSg9YYde8Ccd6BNhCRvm/PNpONq+EJQhhEgDDLhYhNk9Z/D
tyzbUJ0CgYB9XAHT44ZbMVF6mF9egdx8HMQg0KgjUL8mAi5gvlbZjLyxpS0nnLhz
mjnO50tNmJ5OE3QGOEMr9Tzs6Sj7g7pv1McWtCpuvgAg8R4qL0a7MJ8qGSK29RP8
rfCdVg0GCAEjF9Dx+rdn4rWwP/8yJ5I61pWUDgmhKXyki94nPZt6Zg==
-----END RSA PRIVATE KEY-----`)
}
