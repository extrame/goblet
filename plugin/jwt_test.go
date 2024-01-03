package plugin

import (
	"fmt"
	"strings"
	"testing"

	"github.com/extrame/jose/jws"
)

func TestJwt(t *testing.T) {
	var claims = make(jws.Claims)
	claims.Set("userId", "1")
	method := jws.GetSigningMethod("HS512")
	j := jws.NewJWT(claims, method)
	j.Claims().SetIssuer("test")

	b, err := j.Serialize([]byte("test"))
	if err == nil {
		fmt.Printf("Bearer %s\n", string(b))
		auth := strings.TrimPrefix(string(b), "Bearer ")
		token, err := jws.ParseJWT([]byte(auth))
		if err == nil {
			err = token.Validate([]byte("test"))
			if err == nil {
				fmt.Println("Result is:", token.Claims().Get("userId").(string))
			} else {
				t.Fatal(err)
			}
		} else {
			t.Fatal(err)
		}
	} else {
		t.Fatal(err)
	}
}

func TestJwt2(t *testing.T) {
	encrypted := `eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJhY2Nlc3NfdG9rZW4iOiJtZ21Cd2ZzMmdmYmN2RmRMSkt2Y3JkV3VGRHRZcWRkbyIsImFwcF9pZCI6IlFWbDkxd2daNEtIbm81OEYiLCJleHAiOjE3MDE0ODczMjAsImp0aSI6ImNlNTk5NTM3LTllNjQtNDFkMS04ZDRmLTY2OGNjZmQyYjM1YyJ9.PMxg-q7GhV8wAFD2M4s6V9vcO0Fa669wYb13rRk5Bn8`

	token, err := jws.ParseJWT([]byte(encrypted))
	if err == nil {
		// var secret []byte
		// secret, err = base64.StdEncoding.DecodeString("NTQyMzRjODM1ODMwOWI0NGM0YzdhMjE4ZTVmMjM5OWM=")
		err = token.Validate([]byte("54234c8358309b44c4c7a218e5f2399c"))
		if err == nil {
			fmt.Println("Result is:", token.Claims().Get("access_token").(string))
		} else {
			t.Fatal(err)
		}
	} else {
		t.Fatal(err)
	}
}
