package chef

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/ctdk/goiardi/chefcrypto"
	. "github.com/smartystreets/goconvey/convey"
)

type keyPair struct {
	private,
	public,
	kind string
}

const (
	userid     = "tester"
	requestURL = "http://localhost:80"

	// Generated from
	// openssl genrsa -out privkey.pem 2048
	// perl -pe 's/\n/\\n/g' privkey.pem
	privateKeyPKCS1 = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAx12nDxxOwSPHRSJEDz67a0folBqElzlu2oGMiUTS+dqtj3FU
h5lJc1MjcprRVxcDVwhsSSo9948XEkk39IdblUCLohucqNMzOnIcdZn8zblN7Cnp
W03UwRM0iWX1HuwHnGvm6PKeqKGqplyIXYO0qlDWCzC+VaxFTwOUk31MfOHJQn4y
fTrfuE7h3FTElLBu065SFp3dPICIEmWCl9DadnxbnZ8ASxYQ9xG7hmZduDgjNW5l
3x6/EFkpym+//D6AbWDcVJ1ovCsJL3CfH/NZC3ekeJ/aEeLxP/vaCSH1VYC5VsYK
5Qg7SIa6Nth3+RZz1hYOoBJulEzwljznwoZYRQIDAQABAoIBADPQol+qAsnty5er
PTcdHcbXLJp5feZz1dzSeL0gdxja/erfEJIhg9aGUBs0I55X69VN6h7l7K8PsHZf
MzzJhUL4QJJETOYP5iuVhtIF0I+DTr5Hck/5nYcEv83KAvgjbiL4ZE486IF5awnL
2OE9HtJ5KfhEleNcX7MWgiIHGb8G1jCqu/tH0GI8Z4cNgUrXMbczGwfbN/5Wc0zo
Dtpe0Tec/Fd0DLFwRiAuheakPjlVWb7AGMDX4TyzCXfMpS1ul2jk6nGFk77uQozF
PQUawCRp+mVS4qecgq/WqfTZZbBlW2L18/kpafvsxG8kJ7OREtrb0SloZNFHEc2Q
70GbgKECgYEA6c/eOrI3Uour1gKezEBFmFKFH6YS/NZNpcSG5PcoqF6AVJwXg574
Qy6RatC47e92be2TT1Oyplntj4vkZ3REv81yfz/tuXmtG0AylH7REbxubxAgYmUT
18wUAL4s3TST2AlK4R29KwBadwUAJeOLNW+Rc4xht1galsqQRb4pUzkCgYEA2kj2
vUhKAB7QFCPST45/5q+AATut8WeHnI+t1UaiZoK41Jre8TwlYqUgcJ16Q0H6KIbJ
jlEZAu0IsJxjQxkD4oJgv8n5PFXdc14HcSQ512FmgCGNwtDY/AT7SQP3kOj0Rydg
N02uuRb/55NJ07Bh+yTQNGA+M5SSnUyaRPIAMW0CgYBgVU7grDDzB60C/g1jZk/G
VKmYwposJjfTxsc1a0gLJvSE59MgXc04EOXFNr4a+oC3Bh2dn4SJ2Z9xd1fh8Bur
UwCLwVE3DBTwl2C/ogiN4C83/1L4d2DXlrPfInvloBYR+rIpUlFweDLNuve2pKvk
llU9YGeaXOiHnGoY8iKgsQKBgQDZKMOHtZYhHoZlsul0ylCGAEz5bRT0V8n7QJlw
12+TSjN1F4n6Npr+00Y9ov1SUh38GXQFiLq4RXZitYKu6wEJZCm6Q8YXd1jzgDUp
IyAEHNsrV7Y/fSSRPKd9kVvGp2r2Kr825aqQasg16zsERbKEdrBHmwPmrsVZhi7n
rlXw1QKBgQDBOyUJKQOgDE2u9EHybhCIbfowyIE22qn9a3WjQgfxFJ+aAL9Bg124
fJIEzz43fJ91fe5lTOgyMF5TtU5ClAOPGtlWnXU0e5j3L4LjbcqzEbeyxvP3sn1z
dYkX7NdNQ5E6tcJZuJCGq0HxIAQeKPf3x9DRKzMnLply6BEzyuAC4g==
-----END RSA PRIVATE KEY-----
`
	// Generated from
	// openssl rsa -in privkey.pem -pubout -out pubkey.pem
	// perl -pe 's/\n/\\n/g' pubkey.pem
	publicKeyPKCS1 = `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAx12nDxxOwSPHRSJEDz67
a0folBqElzlu2oGMiUTS+dqtj3FUh5lJc1MjcprRVxcDVwhsSSo9948XEkk39Idb
lUCLohucqNMzOnIcdZn8zblN7CnpW03UwRM0iWX1HuwHnGvm6PKeqKGqplyIXYO0
qlDWCzC+VaxFTwOUk31MfOHJQn4yfTrfuE7h3FTElLBu065SFp3dPICIEmWCl9Da
dnxbnZ8ASxYQ9xG7hmZduDgjNW5l3x6/EFkpym+//D6AbWDcVJ1ovCsJL3CfH/NZ
C3ekeJ/aEeLxP/vaCSH1VYC5VsYK5Qg7SIa6Nth3+RZz1hYOoBJulEzwljznwoZY
RQIDAQAB
-----END PUBLIC KEY-----
`
	// Generated from
	// openssl dsaparam -out dsaparam.pem 2048
	// openssl gendsa  -out privkey.pem dsaparam.pem
	// perl -pe 's/\n/\\n/g' privkey.pem
	badPrivateKeyPKCS1 = `
-----BEGIN DSA PRIVATE KEY-----
MIIDVgIBAAKCAQEApv0SsaKRWyn0IrbI6i547c/gldLQ3vB5xoSuTkVOvmD3HfuE
EVPKMS+XKlhgHOJy677zYNKUOIR78vfDVr1M89w19NSic81UwGGaOkrjQWOkoHaA
BS4046AzYKWqHWQNn9dm7WdQlbMBcBv9u+J6EqlzstPwWVaRdbAzyPtwQZRF5WfC
OcrQr8XpXbKsPh55FzfvFpu4KEKTY+8ynLz9uDNW2iAxj9NtRlUHQNqKQvjQsr/8
4pVrEBh+CnzNrmPXQIbyxV0y8WukAo3I3ZXK5nsUcJhFoVCRx4aBlp9W96mYZ7OE
dPCkFsoVhUNFo0jlJhMPODR1NXy77c4v1Kh6xwIhAJwFm6CQBOWJxZdGo2luqExE
acUG9Hkr2qd0yccgs2tFAoIBAQCQJCwASD7X9l7nZyZvJpXMe6YreGaP3VbbHCz8
GHs1P5exOausfJXa9gRLx2qDW0sa1ZyFUDnd2Dt810tgAhY143lufNoV3a4IRHpS
Fm8jjDRMyBQ/BrLBBXgpwiZ9LHBuUSeoRKY0BdyRsULmcq2OaBq9J38NUblWSe2R
NjQ45X6SGgUdHy3CrQtLjCA9l8+VPg3l05IBbXIhVSllP5AUmMG4T9x6M7NHEoSr
c7ewKSJNvc1C8+G66Kfz8xcChKcKC2z1YzvxrlcDHF+BBLw1Ppp+yMBfhQDWIZfe
6tpiKEEyWoyi4GkzQ+vooFIriaaL+Nnggh+iJ7BEUByHBaHnAoIBAFUxSB3bpbbp
Vna0HN6b+svuTCFhYi9AcmI1dcyEFKycUvZjP/X07HvX2yrL8aGxMJgF6RzPob/F
+SZar3u9Fd8DUYLxis6/B5d/ih7GnfPdChrDOJM1nwlferTGHXd1TBDzugpAovCe
JAjXiPsGmcCi9RNyoGib/FgniT7IKA7s3yJAzYSeW3wtLToSNGFJHn+TzFDBuWV4
KH70bpEV84JIzWo0ejKzgMBQ0Zrjcsm4lGBtzaBqGSvOrlIVFuSWFYUxrSTTxthQ
/JYz4ch8+HsQC/0HBuJ48yALDCVKsWq4Y21LRRJIOC25DfjwEYWWaKNGlDDsJA1m
Y5WF0OX+ABcCIEXhrzI1NddyFwLnfDCQ+sy6HT8/xLKXfaipd2rpn3gL
-----END DSA PRIVATE KEY-----
`
	// Generated from
	// openssl genpkey -out rsakey.pem -algorithm RSA -pkeyopt rsa_keygen_bits:2048
	// openssl genrsa -out privkey.pem 2048
	privateKeyPKCS8 = `
-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDNjtxSUP5FjiD9
a0KXByeLPE1y5d7G1WpJOo6YgAJjFUFPYs8+EtF7MzWpxvcRQEuYgrR7K5E7ZmSk
uM3fg+kWessqrc8qZLx3LFVv7C2O2IT0s2riHjBbBOjLbM0Ps9uX5u5vgyIOlEGz
o1dw5AMDi52QjjfROMML7WqRLMY7jcRuK7IpL5UhnAtKnOrakHSzxMHqIC2ZQnsJ
Es2Rnj7ihgr6VZ66FEEUcIqbUwZDEHYsamkg4bCFHB+P925FeZfQtBDBGlFGeNSs
mDOKrw66I2wDdq/BZ7MN3y/tdpda0H+95qYRye2FeyL9uSoREWaAv5PemQYGt2wc
xmkNoImRAgMBAAECggEABFJ2q3xsfEXqx6lTsx1BZZoU/s96ia+/Fl8W1HoMkszF
nMe1F9cJdI+1FybJ1yEE9eX5qYVW/mq+vv/rxEFfy0s1rmYNLxUDKXZTLZFHu/Mt
iH+lRa/g0GkgA/b7sNLVUTJX3RxiwO+5Ge/bTNJehdqPq5Rx9AI/h6asUPUiDep5
gy22eGh8hNYXrDvZxQBe8stVw11PSItn5pgYTtlLW+AxdR5r17JvIsxbdX+nceEK
KWiS8YvkPJwlhIskMu8nBlc62efk6R8bVIRCrgbn87KNe/SmOTgUvgdw5zL5UxU7
m3IMdy7Cl9+0h7AYKUha2d05cAw5nEvmcJlOGjwygQKBgQD4vOuEJXcjjOYzOwEO
DbCfExCl9KnCCOJewq37AxBWreo3gWu4+S4RxSnsEN7NPQGj4JfePr/gSzcr0Zkb
wDZc1jVIUdh5eyE1ABvJWnyfYducKF1j5hO0XJNlHqg1+5DhtycsQRlsbiMDEUxk
1S/zMMg3Af/y87Su/wmnZdCo+QKBgQDTjzY2iaxhut3gi8vKzrS+YAAsjHw+ohT5
WVgFp+TP1lFEjV8hLhWWvnbfluFItzLawjYNpckNcHEA/cgTtsy2baEdrkhhFIj0
1FF2xIYJzHucHZT9e8hMU6FyoX/iqXSfA9bmc5LSV/Bi6nN8hneIcz/x/Vt1z3qd
EeUgHYnjWQKBgGwR2NnPVVYSz6mOh0TN2eEjbWZNSLxPE9tMBj8684xVf5+iEWWK
jeOWoEI6ijLtwJqs6A7dgIw44b2eEUGnX3cycm/7b2xIfQMECw6Oy/qLj9jnCLxw
qDsCxd93VGov5KDM7K4jkqIzr+6TQ3fD0FN+7F5J9iRekjA+Crm6WNAxAoGBAJkC
84rueCcXKHLHqVW9uywV8wpFcXc7c0AFRoyQqgVIVO7n8O3mjubASuncDoSxO67M
2Jt2VLvLn2/AHX1ksRsgn28AJolQeN3a0jC8YtWjd6OqIaBUbsIFmrd15zDgruBz
vnJfFMndoJdqSqy99KZT9OPpAsVqkpwX3UglFR3BAoGBAJLMwZ1bKqIH1BrZhSdx
dtDSoMoQsg+5mWVx5DXSyN4cgkykfbIqAPh8xe6hDFUwMBPniVj9D1c67YYPs/7/
9UtZHPN4s55Li7gJ4tGIpRkcThMEbdBE9rBzgFdNSPloBzwJgC4/XgWR6ZQr6zXD
CD/2ADbs1OybuNTkDSiPdw9K
-----END PRIVATE KEY-----
	`
	// Generated from
	// openssl rsa -in privkey.pem -pubout -out pubkey.pem
	// perl -pe 's/\n/\\n/g' pubkey.pem
	publicKeyPKCS8 = `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzY7cUlD+RY4g/WtClwcn
izxNcuXextVqSTqOmIACYxVBT2LPPhLRezM1qcb3EUBLmIK0eyuRO2ZkpLjN34Pp
FnrLKq3PKmS8dyxVb+wtjtiE9LNq4h4wWwToy2zND7Pbl+bub4MiDpRBs6NXcOQD
A4udkI430TjDC+1qkSzGO43EbiuyKS+VIZwLSpzq2pB0s8TB6iAtmUJ7CRLNkZ4+
4oYK+lWeuhRBFHCKm1MGQxB2LGppIOGwhRwfj/duRXmX0LQQwRpRRnjUrJgziq8O
uiNsA3avwWezDd8v7XaXWtB/veamEcnthXsi/bkqERFmgL+T3pkGBrdsHMZpDaCJ
kQIDAQAB
-----END PUBLIC KEY-----
`
)

var (
	testRequiredHeaders = []string{
		"X-Ops-Timestamp",
		"X-Ops-UserId",
		"X-Ops-Sign",
		"X-Ops-Content-Hash",
		"X-Ops-Authorization-1",
		"X-Ops-Request-Source",
	}

	mux      *http.ServeMux
	server   *httptest.Server
	client   *Client
	keyPairs = []keyPair{
		{
			privateKeyPKCS1,
			publicKeyPKCS1,
			"PKCS1",
		},
		{
			privateKeyPKCS8,
			publicKeyPKCS8,
			"PKCS8",
		},
	}
)

// Gave up trying to implement this myself
// nopCloser came from https://groups.google.com/d/msg/golang-nuts/J-Y4LtdGNSw/wDSYbHWIKj0J
// yay for sharing
// nopCloser creates a io.ReadCloser to satisfy the request.Body input
type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
	client, _ = NewClient(&Config{
		Name:                  userid,
		Key:                   privateKeyPKCS1,
		BaseURL:               server.URL,
		AuthenticationVersion: "1.0",
	})
}

func setupWithKey(privateKey string) {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
	client, _ = NewClient(&Config{
		Name:                  userid,
		Key:                   privateKey,
		BaseURL:               server.URL,
		AuthenticationVersion: "1.0",
	})
}

func teardown() {
	server.Close()
}

func createServer(key *keyPair) *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(
			func(rw http.ResponseWriter, req *http.Request) {
				checkHeader(key, rw, req)
			},
		),
	)
}

func createTLSServer(key *keyPair) *httptest.Server {
	return httptest.NewTLSServer(
		http.HandlerFunc(
			func(rw http.ResponseWriter, req *http.Request) {
				checkHeader(key, rw, req)
			},
		),
	)
}

// publicKeyFromString parses an RSA public key from a string
func publicKeyFromString(key []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(key)
	if block == nil {
		return nil, fmt.Errorf("block size invalid for '%s'", string(key))
	}
	rsaKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return rsaKey.(*rsa.PublicKey), nil
}

func makeAuthConfig(privateKey string) (*AuthConfig, error) {
	pk, err := PrivateKeyFromString([]byte(privateKey))
	if err != nil {
		return nil, err
	}

	ac := &AuthConfig{
		PrivateKey: pk,
		ClientName: userid,
	}
	return ac, nil
}

func TestAuthConfig(t *testing.T) {
	for _, keys := range keyPairs {
		_, err := makeAuthConfig(keys.private)
		if err != nil {
			t.Error("Failed to create AuthConfig struct from privatekeys and stuff", err)
		}
	}
}

func TestBase64BlockEncodeNoLimit(t *testing.T) {
	for _, keys := range keyPairs {
		ac, _ := makeAuthConfig(keys.private)
		var content string
		for _, key := range []string{"header1", "header2", "header3"} {
			content += fmt.Sprintf("%s:blahblahblah\n", key)
		}
		content = strings.TrimSuffix(content, "\n")

		signature, _ := GenerateSignature(ac.PrivateKey, content)
		Base64BlockEncode(signature, 0)
	}
}

func TestSignRequestBadSignature(t *testing.T) {
	for _, keys := range keyPairs {
		ac, err := makeAuthConfig(keys.private)
		request, err := http.NewRequest("GET", requestURL, nil)
		ac.PrivateKey.PublicKey.N = big.NewInt(23234728432324)

		err = ac.SignRequest(request)
		if err == nil {
			t.Fatal("failed to generate failed signature")
		}
	}
}

func TestSignRequestNoBody(t *testing.T) {
	for _, keys := range keyPairs {
		func() {
			setupWithKey(keys.private)
			defer teardown()
			ac, err := makeAuthConfig(keys.private)
			request, err := client.NewRequest("GET", requestURL, nil)

			err = ac.SignRequest(request)
			if err != nil {
				t.Fatal("failed to generate RequestHeaders")
			}
			count := 0
			for _, requiredHeader := range testRequiredHeaders {
				for header := range request.Header {
					if strings.ToLower(requiredHeader) == strings.ToLower(header) {
						count++
						break
					}
				}
			}
			if count != len(testRequiredHeaders) {
				t.Errorf("apiRequestHeaders didn't return all of testRequiredHeaders received: %+v required %+v", request.Header, testRequiredHeaders)
			}
		}()
	}
}

func TestSignRequestBody(t *testing.T) {
	for _, keys := range keyPairs {
		func() {
			ac, err := makeAuthConfig(keys.private)
			if err != nil {
				t.Fatal(err)
			}
			setupWithKey(keys.private)
			defer teardown()

			// Gave up trying to implement this myself
			// nopCloser came from https://groups.google.com/d/msg/golang-nuts/J-Y4LtdGNSw/wDSYbHWIKj0J
			// yay for sharing
			requestBody := strings.NewReader("somecoolbodytext")
			request, err := client.NewRequest("GET", requestURL, requestBody)

			err = ac.SignRequest(request)
			if err != nil {
				t.Fatal("failed to generate RequestHeaders")
			}
			count := 0
			for _, requiredHeader := range testRequiredHeaders {
				for header := range request.Header {
					if strings.ToLower(requiredHeader) == strings.ToLower(header) {
						count++
						break
					}
				}
			}
			if count != len(testRequiredHeaders) {
				t.Error("apiRequestHeaders didn't return all of testRequiredHeaders")
			}
		}()
	}
}

// <3 goiardi
// Test our headers as goiardi would
// https://github.com/ctdk/goiardi/blob/master/authentication/authentication.go
// func checkHeader(user_id string, r *http.Request) string {
func checkHeader(key *keyPair, rw http.ResponseWriter, req *http.Request) {
	user_id := req.Header.Get("X-OPS-USERID")
	// Since we don't have a real client or user to check against,
	// we'll just verify that input user = output user
	// user, err := actor.GetReqUser(user_id)
	// if err != nil {
	if user_id != userid {
		fmt.Fprintf(rw, "Failed to authenticate as %s with key standard %s", user_id, key.kind)
	}

	contentHash := req.Header.Get("X-OPS-CONTENT-HASH")
	if contentHash == "" {
		fmt.Fprintf(rw, "no content hash provided (%s)", key.kind)
	}

	authTimestamp := req.Header.Get("x-ops-timestamp")
	if authTimestamp == "" {
		fmt.Fprintf(rw, "no timestamp header provided (%s)", key.kind)
	}
	// TODO: Will want to implement this later
	//  else {
	//  // check the time stamp w/ allowed slew
	//  tok, terr := checkTimeStamp(authTimestamp, config.Config.TimeSlewDur)
	//  if !tok {
	//    return terr
	//  }
	// }

	// Eventually this may be put to some sort of use, but for now just
	// make sure that it's there. Presumably eventually it would be used to
	// use algorithms other than sha1 for hashing the body, or using a
	// different version of the header signing algorithm.
	xopssign := req.Header.Get("x-ops-sign")
	var apiVer string
	var hashChk []string
	if xopssign == "" {
		fmt.Fprintf(rw, "missing X-Ops-Sign header (%s)", key.kind)
	} else {
		re := regexp.MustCompile(`version=(\d+\.\d+)`)
		shaRe := regexp.MustCompile(`algorithm=(\w+)`)
		if verChk := re.FindStringSubmatch(xopssign); verChk != nil {
			apiVer = verChk[1]
			if apiVer != "1.0" && apiVer != "1.1" {
				fmt.Fprintf(rw, "Bad version number '%s' in X-Ops-Header with crypto standard %s", apiVer, key.kind)

			}
		} else {
			fmt.Fprintf(rw, "malformed version in X-Ops-Header with crypto standard %s", key.kind)
		}

		// if algorithm is missing, it uses sha1. Of course, no other
		// hashing algorithm is supported yet...
		if hashChk = shaRe.FindStringSubmatch(xopssign); hashChk != nil {
			if hashChk[1] != "sha1" {
				fmt.Fprintf(rw, "Unsupported hashing algorithm '%s' specified in X-Ops-Header with crypto standard %s", hashChk[1], key.kind)
			}
		}
	}

	signedHeaders, sherr := assembleSignedHeader(req)
	if sherr != nil {
		fmt.Fprintf(rw, sherr.Error())
	}

	_, err := HeaderDecrypt(key.public, signedHeaders)
	if err != nil {
		fmt.Fprintf(rw, "unexpected header decryption error '%s' with crypto standard %s", err, key.kind)
	}
}

func TestRequest(t *testing.T) {
	for _, keys := range keyPairs {
		func() {
			ac, err := makeAuthConfig(keys.private)
			server := createServer(&keys)
			defer server.Close()
			setupWithKey(keys.private)
			defer teardown()

			request, err := client.NewRequest("GET", server.URL, nil)

			err = ac.SignRequest(request)
			if err != nil {
				t.Fatal("failed to generate RequestHeaders")
			}

			client := &http.Client{}
			response, err := client.Do(request)
			if err != nil {
				t.Error(err)
			}

			if response.StatusCode != 200 {
				t.Error("Non 200 return code: " + response.Status)
			}

			buf := new(bytes.Buffer)
			buf.ReadFrom(response.Body)
			bodyStr := buf.String()

			if bodyStr != "" {
				t.Error(bodyStr)
			}
		}()
	}
}

func TestRequestToEndpoint(t *testing.T) {
	for _, keys := range keyPairs {
		func() {
			ac, err := makeAuthConfig(keys.private)
			server := createServer(&keys)
			defer server.Close()

			requestBody := strings.NewReader("somecoolbodytext")
			request, err := client.NewRequest("GET", server.URL+"/clients", requestBody)

			err = ac.SignRequest(request)
			if err != nil {
				t.Fatal("failed to generate RequestHeaders")
			}

			client := &http.Client{}
			response, err := client.Do(request)
			if err != nil {
				t.Error(err)
			}

			if response.StatusCode != 200 {
				t.Error("Non 200 return code: " + response.Status)
			}

			buf := new(bytes.Buffer)
			buf.ReadFrom(response.Body)
			bodyStr := buf.String()

			if bodyStr != "" {
				t.Error(bodyStr)
			}
		}()
	}
}

func TestTLSValidation(t *testing.T) {
	for _, keys := range keyPairs {
		func() {
			ac, err := makeAuthConfig(keys.private)
			if err != nil {
				panic(err)
			}
			// Self-signed server
			server := createTLSServer(&keys)
			defer server.Close()

			// Without RootCAs, TLS validation should fail
			chefClient, _ := NewClient(&Config{
				Name:    userid,
				Key:     keys.private,
				BaseURL: server.URL,
			})

			request, err := chefClient.NewRequest("GET", server.URL, nil)
			err = ac.SignRequest(request)
			if err != nil {
				t.Fatal("failed to generate RequestHeaders")
			}

			client := chefClient.client
			response, err := client.Do(request)
			if err == nil {
				t.Fatal("Request should fail due to TLS certification validation failure")
			}

			// Success with RootCAs containing the server's certificate
			certPool := x509.NewCertPool()
			certPool.AddCert(server.Certificate())
			chefClient, _ = NewClient(&Config{
				Name:    userid,
				Key:     keys.private,
				BaseURL: server.URL,
				RootCAs: certPool,
			})

			request, err = chefClient.NewRequest("GET", server.URL, nil)
			err = ac.SignRequest(request)
			if err != nil {
				t.Fatal("failed to generate RequestHeaders")
			}

			client = chefClient.client
			response, err = client.Do(request)
			if err != nil {
				t.Fatal(err)
			}

			if response.StatusCode != 200 {
				t.Error("Non 200 return code: " + response.Status)
			}

			buf := new(bytes.Buffer)
			buf.ReadFrom(response.Body)
			bodyStr := buf.String()

			if bodyStr != "" {
				t.Error(bodyStr)
			}
		}()
	}
}

// More Goiardi <3
func assembleSignedHeader(r *http.Request) (string, error) {
	sHeadStore := make(map[int]string)
	authHeader := regexp.MustCompile(`(?i)^X-Ops-Authorization-(\d+)`)
	for k := range r.Header {
		if c := authHeader.FindStringSubmatch(k); c != nil {
			/* Have to put it into a map first, then sort, in case
			 * the headers don't come out in the right order */
			// skipping this error because we shouldn't even be
			// able to get here with something that won't be an
			// integer. Famous last words, I'm sure.
			i, _ := strconv.Atoi(c[1])
			sHeadStore[i] = r.Header.Get(k)
		}
	}
	if len(sHeadStore) == 0 {
		return "", errors.New("no authentication headers found")
	}

	sH := make([]string, len(sHeadStore))
	sHlimit := len(sH)
	for k, v := range sHeadStore {
		if k > sHlimit {
			return "", errors.New("malformed authentication headers")
		}
		sH[k-1] = v
	}
	signedHeaders := strings.Join(sH, "")

	return signedHeaders, nil
}

func TestGenerateHash(t *testing.T) {
	input, output := HashStr("hi"), "witfkXg0JglCjW9RssWvTAveakI="

	Convey("correctly hashes a given input string", t, func() {
		So(input, ShouldEqual, output)
	})
}

// BUG(fujin): @bradbeam: this doesn't make sense to me.
func TestGenerateSignatureError(t *testing.T) {
	for _, keys := range keyPairs {
		ac, _ := makeAuthConfig(keys.private)

		// BUG(fujin): what about the 'hi' string is not meant to be signable?
		sig, err := GenerateSignature(ac.PrivateKey, "hi")

		Convey("sig should be empty?", t, func() {
			So(sig, ShouldNotBeEmpty)
		})

		Convey("errors for an unknown reason to fujin", t, func() {
			So(err, ShouldBeNil)
		})
	}
}

func TestSignatureContent(t *testing.T) {
	pk, _ := PrivateKeyFromString([]byte(privateKeyPKCS1))
	ac := &AuthConfig{
		PrivateKey:            pk,
		ClientName:            userid,
		AuthenticationVersion: "1.0",
	}
	vals := map[string]string{
		"Method":                   "GET",
		"Accept":                   "application/json",
		"Hashed Path":              "FaX3AVJLlDDqHB7giEG/2EbBsR0=",
		"X-Chef-Version":           ChefVersion,
		"X-Ops-Server-API-Version": "1",
		"X-Ops-Timestamp":          "1990-12-31T15:59:60-08:00",
		"X-Ops-UserId":             ac.ClientName,
		"X-Ops-Content-Hash":       "Content-Hash",
	}
	expected := "Method:GET\nHashed Path:FaX3AVJLlDDqHB7giEG/2EbBsR0=\nX-Ops-Content-Hash:Content-Hash\nX-Ops-Timestamp:1990-12-31T15:59:60-08:00\nX-Ops-UserId:tester"

	content := ac.SignatureContent(vals)
	if expected != content {
		t.Errorf("Unexpected content wanted: %+v\n delivered: %+v", expected, content)
	}
}

func TestRequestError(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":["Not Available"]}`, http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	resp, _ := http.Get(ts.URL)
	err := CheckResponse(resp)
	cerr, err := ChefError(err)
	matched, err := regexp.MatchString(`^GET http://127.0.0.1:\d+: 503`, cerr.Error())
	if !matched {
		t.Errorf("Request Error returned %+v instead of GET URL 503", cerr.Error())
	}
	if cerr.StatusCode() != http.StatusServiceUnavailable {
		t.Errorf("Request Error returned code %+v instead of %+v\n", cerr.StatusCode(), http.StatusServiceUnavailable)
	}
	if cerr.StatusMethod() != "GET" {
		t.Errorf("Request Error returned method %+v instead of %+v\n", cerr.StatusMethod(), "GET")
	}
	if cerr.StatusMsg() != "Not Available" {
		t.Errorf("Request Error returned msg '%+v' instead of %+v\n", cerr.StatusMsg(), "Not Available")
	}
	if strings.TrimSpace(string(cerr.StatusText())) != `{"error":["Not Available"]}` {
		t.Errorf("Request Error returned text %+v instead of %+v\n", string(cerr.StatusText()), "Not Available")
	}
	matched, err = regexp.MatchString(`http://127.0.0.1:\d+`, cerr.StatusURL().String())
	if !matched {
		t.Errorf("Request Error returned URL %+v instead of %+v\n", cerr.StatusURL(), "http://127.0.0.1")
	}
	matched, err = regexp.MatchString(`http://127.0.0.1:\d+`, cerr.Error())
	if !matched {
		t.Errorf("Request URL returned %+v instead of %+v\n", cerr.StatusURL().String(), "http://127.0.0.1*")
	}
}

func TestNewClient(t *testing.T) {
	cfg := &Config{Name: "testclient", Key: privateKeyPKCS1, SkipSSL: false, Timeout: 1}
	c, err := NewClient(cfg)
	if err != nil {
		t.Error("Couldn't make a valid client...\n", err)
	}
	// simple validations on the created client
	if c.Auth.ClientName != "testclient" {
		t.Error("unexpected client name: ", c.Auth.ClientName)
	}
	if c.client.Timeout != time.Duration(1)*time.Second {
		t.Error("unexpected timeout value: ", c.client.Timeout)
	}

	// Bad PEM should be an error
	cfg = &Config{Name: "blah", Key: "not a key", SkipSSL: false}
	c, err = NewClient(cfg)
	if err == nil {
		t.Error("Built a client from a bad key string")
	}

	// Not a proper key should be an error
	cfg = &Config{Name: "blah", Key: badPrivateKeyPKCS1, SkipSSL: false}
	c, err = NewClient(cfg)
	if err == nil {
		t.Error("Built a client from a bad key string")
	}

	// TODO: Test the value of Authentication assisgned
}

func TestNewRequest(t *testing.T) {
	for _, keys := range keyPairs {
		var err error
		server := createServer(&keys)
		cfg := &Config{Name: "testclient", Key: keys.private, SkipSSL: false}
		c, _ := NewClient(cfg)
		defer server.Close()

		request, err := c.NewRequest("GET", server.URL, nil)
		if err != nil {
			t.Error("HRRRM! we tried to make a request but it failed :`( ", err)
		}

		resp, err := c.Do(request, nil)
		if resp.StatusCode != 200 {
			t.Error("Non 200 return code: ", resp.Status)
		}

		// This should fail because we've got an invalid URI
		_, err = c.NewRequest("GET", "%gh&%ij", nil)
		if err == nil {
			t.Error("This terrible request thing should fail and it didn't")
		}

		// This should fail because there is no TOODLES! method :D
		request, err = c.NewRequest("TOODLES!", "", nil)
		_, err = c.Do(request, nil)
		if err == nil {
			t.Error("This terrible request thing should fail and it didn't")
		}
	}
}

func TestDo_badjson(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/hashrocket", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, " pigthrusters => 100%% ")
	})

	stupidData := struct{}{}
	request, err := client.NewRequest("GET", "hashrocket", nil)
	_, err = client.Do(request, &stupidData)
	if err == nil {
		t.Error(err)
	}
}

// Add Content-Type tests

func TestDoText(t *testing.T) {
	setup()
	defer teardown()

	pigText := " pigthrusters => 100 "
	mux.HandleFunc("/hashrocket", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain")
		fmt.Fprintf(w, pigText)
	})

	var getdata string
	request, err := client.NewRequest("GET", "hashrocket", nil)
	_, err = client.Do(request, &getdata)
	if err != nil {
		t.Error(err)
	}
	if getdata != pigText {
		t.Errorf("Plain text got unexpected string: %+v expected: %+v\n", getdata, pigText)
	}
}

func TestDoJSON(t *testing.T) {
	setup()
	defer teardown()

	jsonText := `{"key": "value"}`
	mux.HandleFunc("/hashrocket", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		fmt.Fprintf(w, jsonText)
	})

	getdata := map[string]string{}
	wantdata := map[string]string{"key": "value"}
	request, err := client.NewRequest("GET", "hashrocket", nil)
	_, err = client.Do(request, &getdata)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(getdata, wantdata) {
		t.Errorf("JSON data got unexpected string: %+v expected: %+v\n", getdata, wantdata)
	}
}

func TestDoDefaultParse(t *testing.T) {
	setup()
	defer teardown()

	jsonText := `{"key": "value"}`
	mux.HandleFunc("/hashrocket", func(w http.ResponseWriter, r *http.Request) {
		// Note: deliberately using a non standard text type
		w.Header().Add("Content-Type", "none/here")
		fmt.Fprintf(w, jsonText)
	})

	getdata := map[string]string{}
	wantdata := map[string]string{"key": "value"}
	request, err := client.NewRequest("GET", "hashrocket", nil)
	_, err = client.Do(request, &getdata)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(getdata, wantdata) {
		t.Errorf("JSON data got unexpected string: %+v expected: %+v\n", getdata, wantdata)
	}
}

func TestBasicAuthHeader(t *testing.T) {
	setup()
	defer teardown()
	req, _ := client.NewRequest("GET", "http://dummy", nil)
	basicAuthHeader(req, "stduser", "stdpassword")
	basicHeader := req.Header.Get("Authorization")
	if basicHeader != "Basic c3RkdXNlcjpzdGRwYXNzd29yZA==" {
		t.Error("BasicAuthHeader credentials not calculated properly")
	}
}

func TestBasicAuth(t *testing.T) {
	header := basicAuth("stduser", "stdpassword")
	if header != "c3RkdXNlcjpzdGRwYXNzd29yZA==" {
		t.Error("BasicAuth credentials not calculated properly")
	}
}

func TestHeaderValue(t *testing.T) {
	for _, keys := range keyPairs {
		func() {
			ac, err := makeAuthConfig(keys.private)
			if err != nil {
				t.Fatal(err)
			}
			setupWithKey(keys.private)
			defer teardown()

			// add 'X-Ops-Request-Source' header value as web if IsWebuiKey is true
			client.IsWebuiKey = true

			requestBody := strings.NewReader("somecoolbodytext")
			request, err := client.NewRequest("GET", requestURL, requestBody)

			err = ac.SignRequest(request)
			if err != nil {
				t.Fatal("failed to generate RequestHeaders")
			}
			if request.Header.Get("X-Ops-Request-Source") != "web" {
				t.Errorf("didn't return header value")
			}

			// Should not add 'X-Ops-Request-Source' header value as web if IsWebuiKey is false
			client.IsWebuiKey = false

			requestBody = strings.NewReader("somecoolbodytext")
			request, err = client.NewRequest("GET", requestURL, requestBody)

			err = ac.SignRequest(request)
			if err != nil {
				t.Fatal("failed to generate RequestHeaders")
			}
			if request.Header.Get("X-Ops-Request-Source") != "" {
				t.Errorf("Returned header value")
			}
		}()
	}
}
